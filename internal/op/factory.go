package op

import (
	"sync"

	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/internal/job/llm"
	"github.com/y7ut/potami/internal/job/search"
	"github.com/y7ut/potami/internal/job/tool"
	"github.com/y7ut/potami/internal/schema"
	"github.com/y7ut/potami/internal/task"
)

const DEFAULT_TASK_LEVEL = 100

var (
	streamFactory    *task.StreamTaskFactory
	factorySyncMutex sync.Mutex
	searchEngines    = map[string]func() search.Engine{
		"tavily": func() search.Engine { return search.NewTavilySearch() },
		"google": func() search.Engine { return search.NewGoogleCustomSearch() },
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

					promptJob := &llm.Dialog{
						Intput:      job.Params,
						Output:      job.Output,
						Model:       job.LlmModel,
						System:      job.SystemPrompt,
						Temperature: job.Temperature,
						Template:    job.Template,
						TopP:        job.TopP,
					}
					promptJob.SetName(job.Description)
					jobs = append(jobs, promptJob)

				}

				if job.Type == "api_tool" {
					toolJob := &tool.APITool{
						Endpoint:     job.Endpoint,
						HttpMethod:   job.Method,
						Params:       job.Params,
						OutputParses: job.OutputParses,
					}
					toolJob.SetName(job.Description)
					jobs = append(jobs, toolJob)
				}

				if job.Type == "search" {
					engineInit, ok := searchEngines[job.SearchEngine]
					if !ok {
						engineInit = searchEngines["tavily"]
					}
					searchOptions := make([]search.SearchEngineOption, 0)
					if _, ok := job.SearchOptions["limit"]; ok {
						limit, ok := job.SearchOptions["limit"].(int)
						if !ok {
							limit = int(job.SearchOptions["limit"].(float64))
						}
						searchOptions = append(searchOptions, search.WithLimit(limit))
					}

					if _, ok := job.SearchOptions["debug"]; ok {
						if debug, ok := job.SearchOptions["debug"].(bool); ok {
							searchOptions = append(searchOptions, search.WithDebug(debug))
						}
					}

					if _, ok := job.SearchOptions["topic"]; ok {
						if topic, ok := job.SearchOptions["topic"].(string); ok {
							searchOptions = append(searchOptions, search.WithOption("topic", topic))
						}
					}

					if _, ok := job.SearchOptions["search_depth"]; ok {
						if searchDepth, ok := job.SearchOptions["search_depth"].(string); ok {
							searchOptions = append(searchOptions, search.WithOption("search_depth", searchDepth))
						}
					}

					if _, ok := job.SearchOptions["days"]; ok {
						days, ok := job.SearchOptions["days"].(int)
						if !ok {
							days = int(job.SearchOptions["days"].(float64))
						}
						searchOptions = append(searchOptions, search.WithOption("days", days))
					}

					depthMode := false
					if _, ok := job.SearchOptions["depth_mode"]; ok {
						depthMode = job.SearchOptions["depth_mode"].(bool)
					}

					blockSize := 3000
					if _, ok := job.SearchOptions["block_size"]; ok {
						blockSize, ok = job.SearchOptions["block_size"].(int)
						if !ok {
							blockSize = int(job.SearchOptions["block_size"].(float64))
						}
					}

					searchJob := &search.SearchService{
						Engine:      engineInit(),
						Options:     searchOptions,
						DepthMode:   depthMode,
						BlockSize:   blockSize,
						QueryField:  job.QueryField,
						OutputField: job.OutputField,
					}
					searchJob.SetName(job.Description)
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
