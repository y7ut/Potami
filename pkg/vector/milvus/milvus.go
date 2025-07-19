package milvus

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/sirupsen/logrus"
	"github.com/y7ut/potami/internal/document"
)

// PooledConnection 连接
type PooledConnection struct {
	client   *milvusclient.Client
	params   *MilvusParams
	lastUsed time.Time

	refCount int32 // 引用计数
}

func (m *PooledConnection) acquire() {
	atomic.AddInt32(&m.refCount, 1)
}

func (m *PooledConnection) release() {
	atomic.AddInt32(&m.refCount, -1)
	m.lastUsed = time.Now()
}

// isHealthy 检查连接是否健康
func (m *PooledConnection) isHealthy(ctx context.Context) error {
	// 使用轻量级操作检查连接状态
	_, err := m.client.ListCollections(ctx, milvusclient.NewListCollectionOption())
	return err
}

func (m *PooledConnection) Query(ctx context.Context, id ...int) (document.DocumentCollection, error) {
	m.acquire()
	defer m.release()
	if len(id) == 0 {
		return nil, fmt.Errorf("id is empty")
	}
	queryOption := milvusclient.NewQueryOption(m.params.Collection)
	if m.params.Partition != "" {
		queryOption = queryOption.WithPartitions(m.params.Partition)
	}
	ids := make([]int64, 0)
	for _, i := range id {
		ids = append(ids, int64(i))
	}
	queryOption = queryOption.WithConsistencyLevel(entity.ClStrong).WithIDs(column.NewColumnInt64("id", ids)).WithOutputFields(outputFields...)

	searchResult, err := m.client.Query(ctx, queryOption)
	if err != nil {
		return nil, err
	}
	docs, err := decodeResultSets(searchResult)
	if err != nil {
		return nil, err
	}

	return docs, nil
}

func (m *PooledConnection) Search(ctx context.Context, query string, vectors []float64, limit int) (document.DocumentCollection, error) {
	m.acquire()
	defer m.release()
	// convert float64 to []entity.Vector
	queryVector := make([]float32, 0)
	for _, v := range vectors {
		queryVector = append(queryVector, float32(v))
	}

	textAnnRequest := milvusclient.NewAnnRequest("text_dense", limit, entity.FloatVector(queryVector)).WithAnnParam(index.NewIvfAnnParam(10))

	hybirdRequests := make([]*milvusclient.AnnRequest, 0)
	hybirdRequests = append(hybirdRequests, textAnnRequest)

	if m.params.useBM25 {
		annParam := index.NewSparseAnnParam()
		annParam.WithDropRatio(BM25_SPARSE_DROP_RATIO)
		bm25Request := milvusclient.NewAnnRequest("text_sparse", limit, entity.Text(query)).WithAnnParam(annParam)
		hybirdRequests = append(hybirdRequests, bm25Request)
	}

	if m.params.useContextEmbed {
		contextAnnRequest := milvusclient.NewAnnRequest("context_emb", limit, entity.FloatVector(queryVector)).WithAnnParam(index.NewIvfAnnParam(10))
		hybirdRequests = append(hybirdRequests, contextAnnRequest)
	}

	outputFields := []string{"id", "text", "dynamic_json"}

	if m.params.useContextEmbed {
		outputFields = append(outputFields, "context_text")
	}

	reranker := milvusclient.NewRRFReranker().WithK(100)
	logrus.Debug("Using hybrid search request length: ", len(hybirdRequests))
	searchOption := milvusclient.NewHybridSearchOption(
		m.params.Collection, // collectionName
		limit,               // limit
		hybirdRequests...,
	).WithReranker(reranker).WithOutputFields(outputFields...)

	if m.params.Partition != "" {
		searchOption = searchOption.WithPartitions(m.params.Partition)
	}

	resultSets, err := m.client.HybridSearch(ctx, searchOption)

	if err != nil {
		return nil, err
	}
	hybirdRequestsDocuments, err := decodeResultSets(resultSets...)
	if err != nil {
		return nil, fmt.Errorf("failed to decode result sets: %v", err)
	}

	return hybirdRequestsDocuments, nil
}

func (m *PooledConnection) Upsert(ctx context.Context, documents ...*document.Document) error {
	m.acquire()
	defer m.release()
	if len(documents) == 0 {
		return fmt.Errorf("documents is empty")
	}
	// Prepare data

	textColumnData := make([]string, 0)
	mateDataColumnData := make([][]byte, 0)
	vectorColumnData := make([][]float32, 0)
	for _, doc := range documents {
		textColumnData = append(textColumnData, doc.Text)

		metaDataJsonBuffer := bytes.NewBuffer([]byte{})

		metadataEncoder := json.NewEncoder(metaDataJsonBuffer)
		metadataEncoder.SetIndent("", "  ")
		if doc.MetaData != nil {
			metadataEncoder.Encode(doc.MetaData)
		} else {
			metaDataJsonBuffer.WriteString("{}")
		}

		mateDataColumnData = append(mateDataColumnData, metaDataJsonBuffer.Bytes())

		vector := make([]float32, 0)
		for _, v := range doc.Embed {
			vector = append(vector, float32(v))
		}
		vectorColumnData = append(vectorColumnData, vector)
	}

	textColumn := column.NewColumnVarChar("text", textColumnData)
	dynamicFieldColumn := column.NewColumnJSONBytes("dynamic_json", mateDataColumnData)
	vectorColumn := column.NewColumnFloatVector("text_dense", int(m.params.Dimensions), vectorColumnData)

	// new UpsertOption
	upsertOption := milvusclient.NewColumnBasedInsertOption(m.params.Collection, textColumn, dynamicFieldColumn, vectorColumn)

	if m.params.Partition != "" {
		upsertOption.WithPartition(m.params.Partition)
	}

	if m.params.useContextEmbed {
		contextColumnData := make([]string, 0)
		contextEmbedColumnData := make([][]float32, 0)
		for _, doc := range documents {
			contextColumnData = append(contextColumnData, doc.Context)
			contextEmbed := make([]float32, 0)
			for _, v := range doc.ContextEmbed {
				contextEmbed = append(contextEmbed, float32(v))
			}
			contextEmbedColumnData = append(contextEmbedColumnData, contextEmbed)
		}

		contextColumn := column.NewColumnVarChar("context_text", contextColumnData)
		contextEmbedColumn := column.NewColumnFloatVector("context_emb", int(m.params.Dimensions), contextEmbedColumnData)
		upsertOption.WithColumns(contextColumn, contextEmbedColumn)
	}

	result, err := m.client.Insert(
		ctx,
		upsertOption,
	)
	if err != nil {
		return err
	}

	if result.IDs != nil {
		for i := 0; i < result.IDs.Len(); i++ {
			id, err := result.IDs.GetAsString(i)
			if err != nil {
				continue
			}
			documents[i].ID = id
			fmt.Println("Upserted ID:", id)
		}
	}

	return nil
}

func releaseCollection(ctx context.Context, m *PooledConnection) error {
	logrus.WithFields(logrus.Fields{
		"connection_id": m.params.Hash(),
		"collection":    m.params.Collection,
	}).Debug("Releasing collection")
	return m.client.ReleaseCollection(ctx, milvusclient.NewReleaseCollectionOption(m.params.Collection))
}

func (m *PooledConnection) Close(ctx context.Context) error {
	err := releaseCollection(ctx, m)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"connection_id": m.params.Hash(),
		"collection":    m.params.Collection,
	}).Debug("Closing connection")
	return m.client.Close(ctx)
}

func decodeResultSets(resultSets ...milvusclient.ResultSet) (document.DocumentCollection, error) {
	documents := make(document.DocumentCollection, 0)

	for _, result := range resultSets {
		idCol := result.GetColumn("id")
		dynResCol := result.GetColumn("dynamic_json")
		textCol := result.GetColumn("text")
		contextCol := result.GetColumn("context_text")
		// hybirdDocuments := make([]document.Document, 0)
		for i := 0; i < result.ResultCount; i++ {
			currentDoc := &document.Document{}
			if len(result.Scores) > 0 && result.Scores[i] != 0 {
				currentDoc.Score = float64(result.Scores[i])
			}

			if idCol != nil {
				id, err := idCol.GetAsInt64(i)
				if err != nil {
					return nil, err
				}
				currentDoc.ID = strconv.FormatInt(id, 10)
			}
			if dynResCol != nil {
				jsonBytes, err := dynResCol.GetAsString(i)
				if err != nil {
					return nil, err
				}
				base64Decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(jsonBytes)))
				dynamicJsonByte, err := io.ReadAll(base64Decoder)
				if err != nil {
					return nil, err
				}
				var meta map[string]string
				err = json.Unmarshal(dynamicJsonByte, &meta)
				if err != nil {
					return nil, err
				}
				currentDoc.MetaData = meta
			}
			if textCol != nil {
				text, err := textCol.GetAsString(i)
				if err != nil {
					return nil, err
				}
				currentDoc.Text = text
			}
			if contextCol != nil {
				context, err := contextCol.GetAsString(i)
				if err != nil {
					return nil, err
				}
				currentDoc.Context = context
			}
			// fmt.Println(currentDoc)
			documents = append(documents, currentDoc)
		}
		// documents = append(documents, hybirdDocuments)
	}
	return documents, nil
}
