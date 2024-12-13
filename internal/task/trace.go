package task

import (
	"time"
)

type TraceRecord struct {
	TraceID  string         `json:"trace_id"`
	Name     string         `json:"name"`
	Duration int64          `json:"duration"`
	StartAt  string         `json:"start_at"`
	FinishAt string         `json:"finish_at"`
	Error    string         `json:"error"`
	Bill     float64        `json:"bill"`
	Inputs   []string       `json:"inputs"`
	Outputs  []string       `json:"outputs"`
	Options  map[string]any `json:"options"`
	Retrys   []*TraceRecord `json:"retrys"`
}

func GetTracesFromTask(t *Task) []*TraceRecord {
	traces := make([]*TraceRecord, 0)
	for current := t.JobsPipline.Front(); current != nil; current = current.Next() {
		j := current.Value.(Job)
		traces = append(traces, GetTracesFromJob(j))
	}
	return traces
}

func GetTracesFromJob(j Job) *TraceRecord {
	if j.GetTraceIDS() == nil || len(j.GetTraceIDS()) == 0 {
		return nil
	}
	mainTraceID := j.GetTraceIDS()[0]
	mainStartAt, _ := j.GetStartAt(mainTraceID)
	mainStartTime := ""
	if !mainStartAt.IsZero() {
		mainStartTime = mainStartAt.Format(time.RFC3339Nano)
	}
	mainFinishedAt, _ := j.GetFinishAt(mainTraceID)
	mainFinishTime := ""
	if !mainFinishedAt.IsZero() {
		mainFinishTime = mainFinishedAt.Format(time.RFC3339Nano)
	}
	var mainDuration int64
	if !mainStartAt.IsZero() && !mainFinishedAt.IsZero() {
		mainDuration = mainFinishedAt.Sub(mainStartAt).Milliseconds()
	}
	mainError, _ := j.GetError(mainTraceID)
	mainBill, _ := j.GetBill(mainTraceID)
	mainInputs, _ := j.GetInput(mainTraceID)
	mainOutputs, _ := j.GetOutput(mainTraceID)
	var retrys []*TraceRecord
	if len(j.GetTraceIDS()) > 1 {
		retrys = make([]*TraceRecord, 0)
		for _, traceID := range j.GetTraceIDS()[1:] {
			startAt, _ := j.GetStartAt(traceID)
			startTime := ""
			if !startAt.IsZero() {
				startTime = startAt.Format(time.RFC3339Nano)
			}
			finishAt, _ := j.GetFinishAt(traceID)
			finishTime := ""
			if !finishAt.IsZero() {
				finishTime = finishAt.Format(time.RFC3339Nano)
			}
			var duration int64
			if !startAt.IsZero() && !finishAt.IsZero() {
				duration = finishAt.Sub(startAt).Milliseconds()
			}
			retryError, _ := j.GetError(traceID)
			retryBill, _ := j.GetBill(traceID)
			retryInputs, _ := j.GetInput(traceID)
			retryOutputs, _ := j.GetOutput(traceID)
			retrys = append(retrys, &TraceRecord{
				TraceID:  traceID,
				Name:     j.GetName(),
				StartAt:  startTime,
				FinishAt: finishTime,
				Duration: duration,
				Error:    retryError,
				Bill:     retryBill,
				Inputs:   retryInputs,
				Outputs:  retryOutputs,
				Options:  j.GetOptions(),
				Retrys:   nil,
			})
		}
	}
	return &TraceRecord{
		TraceID:  mainTraceID,
		Name:     j.GetName(),
		StartAt:  mainStartTime,
		FinishAt: mainFinishTime,
		Duration: mainDuration,
		Error:    mainError,
		Bill:     mainBill,
		Inputs:   mainInputs,
		Outputs:  mainOutputs,
		Options:  j.GetOptions(),
		Retrys:   retrys,
	}
}
