package op

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/y7ut/potami/internal/db"
	"github.com/y7ut/potami/internal/schema"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/internal/vector"
	"github.com/y7ut/potami/pkg/embedding"
	"github.com/y7ut/potami/pkg/llm"
)

var (
	corpusInits map[string]CorpusBuilder
	corpusInfo  map[string]*schema.Corpus
	corpusOnce  sync.Once
	corpusMutex sync.RWMutex
)

// InitCorpus initializes the corpus
func InitCorpus() {
	corpusOnce.Do(func() {
		corpusInits = make(map[string]CorpusBuilder)
		corpusInfo = make(map[string]*schema.Corpus)
		if err := loadCorpusFromDB(); err != nil {
			logrus.WithError(err).Fatal("Failed to initialize corpus from database")
		}
	})
}

// loadCorpusFromDB loads corpus from the database
func loadCorpusFromDB() error {
	corpusFromDB, err := fetchAndBuildCorpus()
	if err != nil {
		return err
	}

	corpusMutex.Lock()
	defer corpusMutex.Unlock()
	for _, c := range corpusFromDB {
		logrus.Infof("Initializing corpus from DB: %s", c.Name)
		icb, err := GetCorpusBuilder(c)
		if err != nil {
			return err
		}
		corpusInits[c.Name] = icb
		corpusInfo[c.Name] = c
	}
	return nil
}

// fetchAndBuildCorpus fetches all corpus from the database
func fetchAndBuildCorpus() (map[string]*schema.Corpus, error) {
	ctx := context.Background()

	dbCorpus, err := db.GetQueries().ListCorpus(ctx)
	if err != nil {
		return nil, err
	}

	corpusResult := make(map[string]*schema.Corpus)
	for _, c := range dbCorpus {
		EmbeddingOptions := make(map[string]interface{})
		if c.EmbeddingOptions.Valid && c.EmbeddingOptions.String != "" {
			if err := json.Unmarshal([]byte(c.EmbeddingOptions.String), &EmbeddingOptions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal embedding options: %w", err)
			}
		}
		corpusResult[c.Name] = &schema.Corpus{
			Name:               c.Name,
			Description:        c.Description.String,
			CollectionName:     c.CollectionName.String,
			RerankFusionMethod: c.RerankFusionMethod.String,
			EmbeddingProvider:  c.EmbeddingProvider.String,
			EmbeddingModel:     c.EmbeddingModel.String,
			EmbeddingOptions:   EmbeddingOptions,
			VectorStore:        c.VectorStore.String,
			VectorDimension:    c.VectorDimension.Int64,
			VectorMetricType:   c.VectorMetricType.String,
			UseBm25Index:       c.UseBm25Index.Bool,
			UseContextEmbed:    c.UseContextEmbed.Bool,
			ContextEmbedPrompt: c.ContextEmbedPrompt.String,
		}
	}
	return corpusResult, nil
}

func GetCorpus(name string) (c *schema.Corpus, ok bool) {
	corpusMutex.RLock()
	defer corpusMutex.RUnlock()
	c, ok = corpusInfo[name]
	return
}

func UpdateCorpus(c *schema.Corpus) error {
	corpusMutex.Lock()
	defer corpusMutex.Unlock()

	initCorpus, err := GetCorpusBuilder(c)
	if err != nil {
		return err
	}
	corpusInits[c.Name] = initCorpus

	corpusInfo[c.Name] = c
	return nil
}

func RemoveCorpus(name string) {
	corpusMutex.Lock()
	defer corpusMutex.Unlock()
	
	delete(corpusInfo, name)
	delete(corpusInits, name)
}

func GetCorpusList() []*schema.Corpus {
	corpusMutex.RLock()
	defer corpusMutex.RUnlock()

	list := make([]*schema.Corpus, 0, len(corpusInits))
	for _, c := range corpusInfo {
		list = append(list, c)
	}

	slices.SortFunc(list, func(a, b *schema.Corpus) int {
		return strings.Compare(a.Name, b.Name)
	})

	return list
}

type CorpusBuilder func(optionHelper task.Tracer) (*vector.Corpus, error)

func GetCorpusBuilder(corpus *schema.Corpus) (func(optionHelper task.Tracer) (*vector.Corpus, error), error) {

	var embedProviderInit func(optionHelper task.Tracer) embedding.Embed
	embedProviderInit, ok := EmbedProviders[corpus.EmbeddingProvider]
	if !ok {
		return nil, fmt.Errorf("unknown embedding provider: %s", corpus.EmbeddingProvider)
	}

	var llMProviderInit func(optionHelper task.Tracer) llm.Provider
	if contextGenerateLLM, ok := corpus.EmbeddingOptions["context_generate_llm_provider"].(string); ok {
		llMProviderInit, ok = LLMProviders[contextGenerateLLM]
		if !ok {
			return nil, fmt.Errorf("unknown LLM provider: %s", contextGenerateLLM)
		}
	}

	return func(optionHelper task.Tracer) (*vector.Corpus, error) {
		c := &vector.Corpus{
			Topic:           corpus.Name,
			Description:     corpus.Description,
			CollectionName:  corpus.CollectionName,
			UseBm25Index:    corpus.UseBm25Index,
			UseContextEmbed: corpus.UseContextEmbed,
			VectorDimension: corpus.VectorDimension,
		}

		c.EmbedProvider = embedProviderInit(optionHelper)
		if corpus.VectorDimension > 0 {
			optionHelper.SetOption("dimensions", corpus.VectorDimension)
		}

		if corpus.EmbeddingModel != "" {
			optionHelper.SetOption("embedding_model", corpus.EmbeddingModel)
		}

		if llMProviderInit != nil {
			c.LLMProvider = llMProviderInit(optionHelper)
			contextGenerateLLMModel, ok := corpus.EmbeddingOptions["context_generate_llm_model"].(string)
			if ok {
				optionHelper.SetOption("model", contextGenerateLLMModel)
			}
		}
		c.MilvusConnectionManager = MilvusConnectionPool()
		return c, nil
	}, nil
}
