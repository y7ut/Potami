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
	ParentID string         `json:"parent_id,omitempty"`
}

func GetTracesFromTask(t *Task) []*TraceRecord {
	traces := make([]*TraceRecord, 0)
	for current := t.JobsPipline.Front(); current != nil; current = current.Next() {
		j := current.Value.(Job)
		if trace := GetTracesFromJob(j); trace != nil {
			traces = append(traces, trace...)
		}
	}
	return traces
}

func GetTracesFromJob(j Job) []*TraceRecord {
	if j.GetTraceIDS() == nil || len(j.GetTraceIDS()) == 0 {
		return nil
	}
	traces := make([]*TraceRecord, 0)

	for i, traceID := range j.GetTraceIDS() {
		startAt, _ := j.GetStartAt(traceID)
		finishAt, _ := j.GetFinishAt(traceID)
		startTime := ""
		if !startAt.IsZero() {
			startTime = startAt.Format(time.RFC3339Nano)
		}
		finishTime := ""
		if !finishAt.IsZero() {
			finishTime = finishAt.Format(time.RFC3339Nano)
		}
		var druation int64
		if !startAt.IsZero() && !finishAt.IsZero() {
			druation = finishAt.Sub(startAt).Milliseconds()
		}
		bill, _ := j.GetBill(traceID)
		inputs, _ := j.GetInput(traceID)
		outputs, _ := j.GetOutput(traceID)
		traceError, _ := j.GetError(traceID)
		t := &TraceRecord{
			TraceID:  traceID,
			Name:     j.GetDescription(),
			StartAt:  startTime,
			FinishAt: finishTime,
			Duration: druation,
			Bill:     bill,
			Inputs:   inputs,
			Outputs:  outputs,
			Options:  j.GetOptions(),
			Error:    traceError,
		}
		if i > 0 {
			t.ParentID = traces[i-1].TraceID
		}
		traces = append(traces, t)

	}
	return traces
}
