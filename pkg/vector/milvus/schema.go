package milvus

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

const (
	BM25_B                 = 0.8
	BM25_K1                = 1.5
	BM25_SPARSE_DROP_RATIO = 0.2
	TEXT_MAX_LENGTH        = 4096
)

var outputFields = []string{"id", "text", "context_text", "dynamic_json"}

func defaultSchema(vectorDimension int64) *entity.Schema {
	return entity.NewSchema().WithDynamicFieldEnabled(true).
		WithField(
			entity.NewField().WithName("id").
				WithIsPrimaryKey(true).
				WithDataType(entity.FieldTypeInt64).
				WithIsAutoID(true),
		).
		WithField(
			entity.NewField().WithName("text_dense").
				WithDataType(entity.FieldTypeFloatVector).
				WithDim(vectorDimension),
		).
		WithField(
			entity.NewField().WithName("text").
				WithDataType(entity.FieldTypeVarChar).
				WithEnableAnalyzer(true).
				WithMaxLength(TEXT_MAX_LENGTH),
		)
}

// createCollection creates a milvus collection with:
// - schema
// - name
// - useBM25
// - useContextEmbed
func createCollection(ctx context.Context, client *milvusclient.Client, collectionName string, useBM25 bool, useContextEmbed bool, vectorDimension int64) error {
	var err error
	var options = make([]milvusclient.CreateIndexOption, 0)

	schema := defaultSchema(vectorDimension)
	if useBM25 {
		// add bm25 and bm25 embedding Function
		schema.WithField(
			entity.NewField().
				WithName("text_sparse").
				WithDataType(entity.FieldTypeSparseVector),
		).WithFunction(
			entity.NewFunction().
				WithName("text_bm25_emb").
				WithInputFields("text").
				WithOutputFields("text_sparse").
				WithType(entity.FunctionTypeBM25),
		)
		// create bm25 index
		bm25IndexOption := milvusclient.NewCreateIndexOption(
			collectionName,
			"text_sparse",
			index.NewSparseInvertedIndex(entity.MetricType(entity.BM25), BM25_SPARSE_DROP_RATIO),
		)
		bm25IndexOption.WithExtraParam("inverted_index_algo", "DAAT_MAXSCORE")
		bm25IndexOption.WithExtraParam("bm25_k1", BM25_K1)
		bm25IndexOption.WithExtraParam("bm25_b", BM25_B)
		bm25IndexOption.WithIndexName("bm25_emb")

		options = append(options, bm25IndexOption)
	}

	// add context, context embedding field and index option
	if useContextEmbed {
		schema.WithField(
			entity.NewField().WithName("context_text").
				WithDataType(entity.FieldTypeVarChar).
				WithEnableAnalyzer(true).
				WithMaxLength(TEXT_MAX_LENGTH),
		).WithField(
			entity.NewField().
				WithName("context_emb").
				WithDataType(entity.FieldTypeFloatVector).
				WithDim(vectorDimension),
		)

		options = append(options, milvusclient.NewCreateIndexOption(
			collectionName,
			"context_emb",
			index.NewAutoIndex(entity.MetricType(entity.IP)),
		).WithIndexName("context_dense_emb"))
	}

	// create text vector index
	textIndexOption := milvusclient.NewCreateIndexOption(
		collectionName,
		"text_dense",
		index.NewAutoIndex(entity.MetricType(entity.IP)),
	)
	textIndexOption.WithIndexName("text_emb")

	options = append(options, textIndexOption)

	err = client.CreateCollection(ctx,
		milvusclient.NewCreateCollectionOption(collectionName, schema).
			WithIndexOptions(options...))

	if err != nil {
		return fmt.Errorf("failed to create collection: %v", err)
	}

	return nil
}
