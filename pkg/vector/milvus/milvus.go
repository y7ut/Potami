package milvus

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/y7ut/potami/internal/document"
)

func (m *PooledConnection) release() {
	atomic.AddInt32(&m.refCount, -1)
}

// isHealthy 检查连接是否健康
func (m *PooledConnection) isHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	// 使用轻量级操作检查连接状态
	_, err := m.client.ListCollections(ctx, milvusclient.NewListCollectionOption())
	return err == nil
}

func (m *PooledConnection) Query(ctx context.Context, id ...int) (document.DocumentCollection, error) {
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
	documents := make([]document.Document, searchResult.ResultCount)

	// Process search results
	for _, column := range searchResult.Fields {
		if column.Name() == "text" {
			for i := 0; i < column.Len(); i++ {
				text, err := column.GetAsString(i)
				if err != nil {
					return nil, err
				}
				documents[i].Text = text
			}
		}
		if column.Name() == "context_text" {
			for i := 0; i < column.Len(); i++ {
				contextText, err := column.GetAsString(i)
				if err != nil {
					return nil, err
				}
				documents[i].Context = contextText
			}
		}
		if column.Name() == "id" {
			for i := 0; i < column.Len(); i++ {
				id, err := column.GetAsInt64(i)
				if err != nil {
					return nil, err
				}
				documents[i].ID = fmt.Sprintf("%d", id)
			}
		}
		if column.Name() == "dynamic_json" {
			for i := 0; i < column.Len(); i++ {
				dynamicJson, err := column.GetAsString(i)
				if err != nil {
					return nil, err
				}
				var meta map[string]string
				err = json.Unmarshal([]byte(dynamicJson), &meta)
				if err != nil {
					return nil, err
				}
				documents[i].MetaData = meta
			}
		}
	}
	return documents, nil
}

func (m *PooledConnection) Search(ctx context.Context, query string, vectors []float64, limit int) (document.DocumentCollection, error) {
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
		annParam.WithDropRatio(0.2)
		bm25Request := milvusclient.NewAnnRequest("text_sparse", 2, entity.Text(query)).WithAnnParam(annParam)
		hybirdRequests = append(hybirdRequests, bm25Request)
	}

	if m.params.useContextEmbed {
		contextAnnRequest := milvusclient.NewAnnRequest("text_dense", limit, entity.FloatVector(queryVector)).WithAnnParam(index.NewIvfAnnParam(10))
		hybirdRequests = append(hybirdRequests, contextAnnRequest)
	}

	outputFields := []string{"id", "text", "dynamic_json"}

	if m.params.useContextEmbed {
		outputFields = append(outputFields, "context_text")
	}

	reranker := milvusclient.NewRRFReranker().WithK(100)

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
	hybirdRequestsDocuments, err := decodeResultSets(resultSets)
	if err != nil {
		return nil, fmt.Errorf("failed to decode result sets: %v", err)
	}
	documents := make([]document.Document, 0)

	for _, collection := range hybirdRequestsDocuments {
		documents = append(documents, collection...)
	}

	return documents, nil
}

func (m *PooledConnection) Upsert(ctx context.Context, documents ...document.Document) error {
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
	vectorColumn := column.NewColumnFloatVector("text_dense", VECTOR_DIMENSION, vectorColumnData)

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

		contextColumn := column.NewColumnString("context", contextColumnData)
		contextEmbedColumn := column.NewColumnFloatVector("context_embed", VECTOR_DIMENSION, contextEmbedColumnData)
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
	log.Printf("release collection: %s \n", m.params.Collection)
	return m.client.ReleaseCollection(ctx, milvusclient.NewReleaseCollectionOption(m.params.Collection))
}

func (m *PooledConnection) Close(ctx context.Context) error {
	err := releaseCollection(ctx, m)
	if err != nil {
		return err
	}
	log.Printf("close connection: %s\n", m.params.Collection)
	return m.client.Close(ctx)
}

func decodeResultSets(resultSets []milvusclient.ResultSet) ([]document.DocumentCollection, error) {
	documents := make([]document.DocumentCollection, 0)

	for _, result := range resultSets {
		idCol := result.GetColumn("id")
		dynResCol := result.GetColumn("dynamic_json")
		textCol := result.GetColumn("text")
		contextCol := result.GetColumn("text_context")
		hybirdDocuments := make([]document.Document, 0)
		for i := 0; i < result.ResultCount; i++ {
			currentDoc := document.Document{
				Score: float64(result.Scores[i]),
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
			hybirdDocuments = append(hybirdDocuments, currentDoc)
		}
		documents = append(documents, hybirdDocuments)
	}
	return documents, nil
}
