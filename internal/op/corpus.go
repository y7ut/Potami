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
	"github.com/y7ut/potami/internal/job"
	"github.com/y7ut/potami/internal/schema"
	"github.com/y7ut/potami/internal/vector"
)

var (
	corpus      map[string]*schema.Corpus
	corpusOnce  sync.Once
	corpusMutex sync.RWMutex
)

// InitCorpus initializes the corpus
func InitCorpus() {
	corpusOnce.Do(func() {
		corpus = make(map[string]*schema.Corpus)
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
		corpus[c.Name] = c
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
		if c.EmbeddingOptions.Valid {
			if err := json.Unmarshal([]byte(c.EmbeddingOptions.String), &EmbeddingOptions); err != nil {
				return nil, err
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
	c, ok = corpus[name]
	return
}

func UpdateCorpus(c *schema.Corpus) {
	corpusMutex.Lock()
	defer corpusMutex.Unlock()
	corpus[c.Name] = c
}

func RemoveCorpus(name string) {
	corpusMutex.Lock()
	defer corpusMutex.Unlock()
	delete(corpus, name)
}

func GetCorpusList() []*schema.Corpus {
	corpusMutex.RLock()
	defer corpusMutex.RUnlock()

	list := make([]*schema.Corpus, 0, len(corpus))
	for _, c := range corpus {
		list = append(list, c)
	}

	slices.SortFunc(list, func(a, b *schema.Corpus) int {
		return strings.Compare(a.Name, b.Name)
	})

	return list
}

func CreateCorpusFromSchema(corpus *schema.Corpus) (*vector.Corpus, error) {

	c := &vector.Corpus{
		Topic:           corpus.Name,
		Description:     corpus.Description,
		CollectionName:  corpus.CollectionName,
		UseBm25Index:    corpus.UseBm25Index,
		UseContextEmbed: corpus.UseContextEmbed,
		VectorDimension: corpus.VectorDimension,
	}

	embedProviderInit, ok := EmbedProviders[corpus.EmbeddingProvider]
	if !ok {
		return nil, fmt.Errorf("unknown embedding provider: %s", corpus.EmbeddingProvider)
	}
	embedOptionsHelper := job.NewBlankJob()
	c.EmbedProvider = embedProviderInit(embedOptionsHelper)
	embedOptionsHelper.SetOption("dimensions", corpus.VectorDimension)

	contextGenerateLLM, ok := corpus.EmbeddingOptions["context_generate_llm_provider"].(string)
	if ok {
		llMProviderInit, ok := LLMProviders[contextGenerateLLM]
		if !ok {
			return nil, fmt.Errorf("unknown LLM provider: %s", contextGenerateLLM)
		}
		llmOptionsHelper := job.NewBlankJob()
		c.LLMProvider = llMProviderInit(llmOptionsHelper)
		contextGenerateLLMModel, ok := corpus.EmbeddingOptions["context_generate_llm_model"].(string)
		if ok {
			llmOptionsHelper.SetOption("model", contextGenerateLLMModel)
		}
	}
	c.MilvusConnectionManager = MilvusConnectionPool()
	return c, nil
}
