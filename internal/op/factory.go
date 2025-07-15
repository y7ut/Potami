package op

import (
	"sync"

	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/internal/job/chat"
	"github.com/y7ut/potami/internal/job/retrieval"
	"github.com/y7ut/potami/internal/job/tool"
	"github.com/y7ut/potami/internal/schema"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/pkg/embedding"
	"github.com/y7ut/potami/pkg/llm"
	"github.com/y7ut/potami/pkg/search"
)

const DEFAULT_TASK_LEVEL = 100

var (
	streamFactory    *task.StreamTaskFactory
	factorySyncMutex sync.Mutex

	SearchRetrievers = map[string]func(options task.WithOption) search.SearchEngine{
		"tavily": func(options task.WithOption) search.SearchEngine {
			return search.NewTavilySearch(options)
		},
		"google": func(options task.WithOption) search.SearchEngine {
			return search.NewGoogleCustomSearch(options)
		},
	}
	LLMProviders = map[string]func(tracer task.Tracer) llm.Provider{
		"openai": func(tracer task.Tracer) llm.Provider {
			return llm.NewOpenAIProvider(tracer)
		},
		"ollama": func(tracer task.Tracer) llm.Provider {
			return llm.NewOllamaProvider(tracer)
		},
	}
	EmbedProviders = map[string]func(options task.WithOption) embedding.Embed{
		"ollama": func(options task.WithOption) embedding.Embed {
			return embedding.NewOllamaEmbedding(options)
		},
		"openai": func(options task.WithOption) embedding.Embed {
			return embedding.NewOpenAIEmbedding(options)
		},
	}
)

func GetStreamFactory() *task.StreamTaskFactory {
	factorySyncMutex.Lock()
	defer factorySyncMutex.Unlock()
	return streamFactory
}

func InitStreamFactory() {
	roadMap, levelMap, descriptionMap := LoadStream(streams)
	factorySyncMutex.Lock()
	defer factorySyncMutex.Unlock()
	streamFactory = task.NewFactory(
		TaskPool,
		conf.Task.RetryCount,
		conf.Task.Tracking,
		roadMap,
		levelMap,
		descriptionMap,
	)
}
func ReloadStreamFactory() {
	roadMap, levelMap, descriptionMap := LoadStream(streams)
	factorySyncMutex.Lock()
	defer factorySyncMutex.Unlock()
	streamFactory.RoadMap = roadMap
	streamFactory.LevelMap = levelMap
	streamFactory.DescriptionMap = descriptionMap
}

// LoadStream 根据配置加载stream工厂所需的策略
func LoadStream(conf map[string]*schema.Stream) (map[string]func() []task.Job, map[string]int, map[string]string) {
	roadMap := make(map[string]func() []task.Job)
	levelMap := make(map[string]int)
	descriptionMap := make(map[string]string)

	for name, stream := range conf {
		roadMap[name] = func() []task.Job {

			jobs := make([]task.Job, 0)
			for _, job := range stream.Jobs {
				descriptionMap[job.Name] = job.Description
				if job.Type == "prompt" {

					promptJob := &chat.Dialog{
						Intput:   job.Params,
						Output:   job.Output,
						System:   job.SystemPrompt,
						Template: job.Template,
					}

					generateProvider, avaliable := LLMProviders[job.LLMProvider]
					if !avaliable {
						generateProvider = LLMProviders["openai"]
					}

					promptJob.Provider = generateProvider(promptJob)
					promptJob.SetName(job.Name)
					promptJob.SetDescription(job.Description)

					promptJob.SetOption("provider", job.LLMProvider)
					if job.LlmModel != "" {
						promptJob.SetOption("model", job.LlmModel)
					}

					if job.Temperature != 0 {
						promptJob.SetOption("temperature", job.Temperature)
					}

					if job.TopP != 0 {
						promptJob.SetOption("top_p", job.TopP)
					}

					if job.MaxTokens != 0 {
						promptJob.SetOption("max_tokens", job.MaxTokens)
					}

					jobs = append(jobs, promptJob)

				}

				if job.Type == "api_tool" {
					toolJob := &tool.APITool{
						Endpoint:     job.Endpoint,
						HttpMethod:   job.Method,
						Params:       job.Params,
						OutputParses: job.OutputParses,
					}
					toolJob.SetName(job.Name)
					toolJob.SetDescription(job.Description)
					jobs = append(jobs, toolJob)
				}

				if job.Type == "search" {
					searchEngineInit, ok := SearchRetrievers[job.SearchEngine]
					if !ok {
						searchEngineInit = SearchRetrievers["tavily"]
					}

					searchJob := &retrieval.Retriever{
						QueryField:  job.QueryField,
						OutputField: job.OutputField,
					}

					searchJob.Retrieval = retrieval.NewWebSearchRetriever(searchEngineInit(searchJob))

					if limit, ok := job.SearchOptions["limit"]; ok {
						searchJob.SetOption("limit", limit)
					}

					if debug, ok := job.SearchOptions["debug"]; ok {
						searchJob.SetOption("debug", debug)
					}

					if topic, ok := job.SearchOptions["topic"]; ok {
						searchJob.SetOption("topic", topic)
					}

					if days, ok := job.SearchOptions["days"]; ok {
						searchJob.SetOption("days", days)
					}

					if days, ok := job.SearchOptions["search_depth"]; ok {
						searchJob.SetOption("search_depth", days)
					}

					if depthMode, ok := job.SearchOptions["depth_mode"]; ok {
						mode := "flat"
						if IsDepth, ok := depthMode.(bool); ok && IsDepth {
							mode = "depth"
						}
						// depth_mode: 1. 深度 Depth 2. 广度 Flat
						searchJob.SetOption("depth_mode", mode)
					}

					if block_size, ok := job.SearchOptions["block_size"]; ok {
						searchJob.SetOption("block_size", block_size)
					}

					searchJob.SetName(job.Name)
					searchJob.SetDescription(job.Description)

					jobs = append(jobs, searchJob)
				}
			}

			return jobs
		}
		levelMap[name] = 1
		if stream.Level != 0 {
			levelMap[name] = stream.Level
		}
		for _, job := range stream.Jobs {
			descriptionMap[job.Name] = job.Description
		}
	}

	return roadMap, levelMap, descriptionMap
}
