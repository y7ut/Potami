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
	streams          map[string]*schema.Stream
	streamsSyncMutex sync.Mutex
)

func InitStreams() {
	if streams != nil {
		return
	}
	streamsFromDB := initStreamsFromDB()

	streamsSyncMutex.Lock()
	defer streamsSyncMutex.Unlock()
	if streams != nil {
		return
	}
	streams = make(map[string]*schema.Stream)
	for _, stream := range streamsFromDB {
		logrus.Infof("init stream from db name %s", stream.Name)
		streams[stream.Name] = stream
	}

}

func initStreamsFromDB() map[string]*schema.Stream {
	streamsFromDB := make(map[string]*schema.Stream)
	dbStreams, err := stream.List(context.Background())
	if err != nil {
		logrus.Fatalf("get db streams config error: %s", err)
	}

	dbJobs, err := job.List(context.Background())
	if err != nil {
		logrus.Fatalf("get db jobs config error: %s", err)
	}
	slices.SortFunc(dbJobs, func(a *db.Job, b *db.Job) int {
		return int(a.Sorted - b.Sorted)
	})
	jobsStreamRelation := make(map[int64][]*schema.Job)
	for _, job := range dbJobs {
		outputParses := make(map[string]string)
		if job.OutputParses.Valid {
			err := json.Unmarshal([]byte(job.OutputParses.String), &outputParses)
			if err != nil {
				logrus.Fatalf("get db jobs config error: %s", err)
			}
		}
		searchOptions := make(map[string]interface{})
		if job.SearchOptions.Valid {
			err := json.Unmarshal([]byte(job.SearchOptions.String), &searchOptions)
			if err != nil {
				logrus.Fatalf("get db jobs config error: %s", err)
			}
		}

		currentJob := &schema.Job{
			Name:          job.Name,
			Type:          job.Type,
			Description:   job.Description.String,
			Params:        strings.Split(job.Params.String, ","),
			LlmModel:      job.LlmModel.String,
			Temperature:   job.Temperature.Float64,
			TopP:          job.TopP.Float64,
			MaxTokens:     int(job.MaxTokens.Int64),
			SystemPrompt:  job.SystemPrompt.String,
			Template:      job.Template.String,
			Endpoint:      job.Endpoint.String,
			Method:        job.Method.String,
			Output:        strings.Split(job.Output.String, ","),
			OutputParses:  outputParses,
			SearchEngine:  job.SearchEngine.String,
			SearchOptions: searchOptions,
			QueryField:    job.QueryField.String,
			OutputField:   job.OutputField.String,
		}
		jobsStreamRelation[job.StreamID] = append(jobsStreamRelation[job.StreamID], currentJob)
	}
	for _, stream := range dbStreams {
		streamsFromDB[stream.Name] = &schema.Stream{
			Name:        stream.Name,
			Description: stream.Description.String,
			Jobs:        jobsStreamRelation[stream.ID],
			Level:       int(stream.Level.Int64),
		}
	}
	return streamsFromDB
}

func GetStream(name string) (stream *schema.Stream, ok bool) {
	streamsSyncMutex.Lock()
	defer streamsSyncMutex.Unlock()
	stream, ok = streams[name]
	return
}

func UpdateStream(stream *schema.Stream) {
	streamsSyncMutex.Lock()
	defer streamsSyncMutex.Unlock()
	streams[stream.Name] = stream
}

func RemoveStream(name string) {
	streamsSyncMutex.Lock()
	defer streamsSyncMutex.Unlock()
	delete(streams, name)
}

func GetStreamList() []*schema.HumanFriendlyStreamConfig {
	StreamHuamnFriendly := make([]*schema.HumanFriendlyStreamConfig, 0)

	streamsSyncMutex.Lock()
	streamsCurrent := streams
	streamsSyncMutex.Unlock()

	for k, v := range streamsCurrent {
		generatorOutput := make(map[string]bool)
		jobs := make([]map[string]string, 0)
		output := make(map[string]bool, 0)
		for _, job := range v.Jobs {
			jobs = append(jobs, map[string]string{
				"name":        job.Name,
				"description": job.Description,
				"type":        job.Type,
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
