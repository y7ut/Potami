package op

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/y7ut/potami/internal/db"
	"github.com/y7ut/potami/internal/schema"
	"github.com/y7ut/potami/internal/service/job"
	"github.com/y7ut/potami/internal/service/stream"
)

var (
	streams      map[string]*schema.Stream
	streamsOnce  sync.Once
	streamsMutex sync.RWMutex
)

// InitStreams initializes the streams
func InitStreams() {
	streamsOnce.Do(func() {
		streams = make(map[string]*schema.Stream)
		if err := loadStreamsFromDB(); err != nil {
			logrus.WithError(err).Fatal("Failed to initialize streams from database")
		}
	})
}

// loadStreamsFromDB loads streams from the database
func loadStreamsFromDB() error {
	streamsFromDB, err := fetchAndBuildStreams()
	if err != nil {
		return err
	}

	streamsMutex.Lock()
	defer streamsMutex.Unlock()
	for _, stream := range streamsFromDB {
		logrus.Infof("Initializing stream from DB: %s", stream.Name)
		streams[stream.Name] = stream
	}
	return nil
}

// fetchAndBuildStreams fetches all streams from the database and builds a map of jobs to streams
func fetchAndBuildStreams() (map[string]*schema.Stream, error) {
	ctx := context.Background()

	dbStreams, err := stream.List(ctx)
	if err != nil {
		return nil, err
	}

	dbJobs, err := job.List(ctx)
	if err != nil {
		return nil, err
	}

	jobsStreamRelation, err := buildJobsStreamRelation(dbJobs)
	if err != nil {
		return nil, err
	}

	streamsResult := make(map[string]*schema.Stream)
	for _, s := range dbStreams {
		streamsResult[s.Name] = &schema.Stream{
			Name:        s.Name,
			Description: s.Description.String,
			Jobs:        jobsStreamRelation[s.ID],
			Level:       int(s.Level.Int64),
		}
	}
	return streamsResult, nil
}

// buildJobsStreamRelation builds a map of jobs to streams
func buildJobsStreamRelation(dbJobs []*db.Job) (map[int64][]*schema.Job, error) {
	slices.SortFunc(dbJobs, func(a, b *db.Job) int {
		return int(a.Sorted - b.Sorted)
	})

	jobsStreamRelation := make(map[int64][]*schema.Job)
	for _, j := range dbJobs {
		currentJob, err := convertDBJobToSchemaJob(j)
		if err != nil {
			logrus.WithError(err).Errorf("Skipping job %s due to conversion error", j.Name)
			continue
		}
		jobsStreamRelation[j.StreamID] = append(jobsStreamRelation[j.StreamID], currentJob)
	}
	return jobsStreamRelation, nil
}

// convertDBJobToSchemaJob converts a DB job to a schema job
func convertDBJobToSchemaJob(job *db.Job) (*schema.Job, error) {
	outputParses := make(map[string]string)
	if job.OutputParses.Valid {
		if err := json.Unmarshal([]byte(job.OutputParses.String), &outputParses); err != nil {
			return nil, err
		}
	}

	searchOptions := make(map[string]interface{})
	if job.SearchOptions.Valid {
		if err := json.Unmarshal([]byte(job.SearchOptions.String), &searchOptions); err != nil {
			return nil, err
		}
	}

	var params []string
	if job.Params.Valid && job.Params.String != "" {
		params = strings.Split(job.Params.String, ",")
	}

	var outputs []string
	if job.Output.Valid && job.Output.String != "" {
		outputs = strings.Split(job.Output.String, ",")
	}

	return &schema.Job{
		Name:          job.Name,
		Type:          schema.JobType(job.Type),
		Description:   job.Description.String,
		Params:        params,
		LlmModel:      job.LlmModel.String,
		LLMProvider:   job.LlmProvider.String,
		Temperature:   job.Temperature.Float64,
		TopP:          job.TopP.Float64,
		MaxTokens:     int(job.MaxTokens.Int64),
		SystemPrompt:  job.SystemPrompt.String,
		Template:      job.Template.String,
		Endpoint:      job.Endpoint.String,
		Method:        job.Method.String,
		Output:        outputs,
		OutputParses:  outputParses,
		SearchEngine:  job.SearchEngine.String,
		SearchOptions: searchOptions,
		QueryField:    job.QueryField.String,
		OutputField:   job.OutputField.String,
		Corpus:        job.Corpus.String,
		ResourceType:  job.ResourceType.String,
		SplitRule:    job.SplitRule.String,
	}, nil
}

func GetStream(name string) (stream *schema.Stream, ok bool) {
	streamsMutex.RLock()
	defer streamsMutex.RUnlock()
	stream, ok = streams[name]
	return
}

func UpdateStream(stream *schema.Stream) {
	streamsMutex.Lock()
	defer streamsMutex.Unlock()
	streams[stream.Name] = stream
}

func RemoveStream(name string) {
	streamsMutex.Lock()
	defer streamsMutex.Unlock()
	delete(streams, name)
}

func GetStreamList() []*schema.HumanFriendlyStreamConfig {
	StreamHuamnFriendly := make([]*schema.HumanFriendlyStreamConfig, 0)

	streamsMutex.RLock()
	streamsCurrent := streams
	streamsMutex.RUnlock()

	for k, v := range streamsCurrent {
		generatorOutput := make(map[string]bool)
		jobs := make([]map[string]string, 0)
		output := make(map[string]bool, 0)
		for _, job := range v.Jobs {
			jobs = append(jobs, map[string]string{
				"name":        job.Name,
				"description": job.Description,
				"type":        string(job.Type),
			})
			if job.Params != nil {
				for _, p := range job.Params {
					if p != "" {
						output[p] = true
					}
				}
			}
			if job.QueryField != "" {
				output[job.QueryField] = true
			}
			if job.Output != nil {
				for _, p := range job.Output {
					if p != "" {
						output[p] = true
						generatorOutput[p] = true
					}
				}
			}
			if job.OutputField != "" {
				generatorOutput[job.OutputField] = true
			}
			if job.OutputParses != nil {
				for k := range job.OutputParses {
					if k != "" {
						output[k] = true
						generatorOutput[k] = true
					}
				}
			}
		}
		unqiueOutput := make([]string, 0)
		requireParams := make([]string, 0)
		for k := range output {
			if _, ok := generatorOutput[k]; !ok {
				requireParams = append(requireParams, k)
				continue
			}
			unqiueOutput = append(unqiueOutput, k)
		}
		StreamHuamnFriendly = append(StreamHuamnFriendly, &schema.HumanFriendlyStreamConfig{
			Name:           k,
			Description:    v.Description,
			Jobs:           jobs,
			RequiredParams: requireParams,
			Output:         unqiueOutput,
		})
	}
	slices.SortFunc(StreamHuamnFriendly, func(a, b *schema.HumanFriendlyStreamConfig) int {
		return strings.Compare(a.Name, b.Name)
	})
	return StreamHuamnFriendly
}
