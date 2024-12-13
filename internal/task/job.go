package task

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// 可记录账单的
type BillingRecord interface {
	Billing(float64)
	GetBills() map[string]float64
	GetBill(traceId string) (float64, error)
}

// 可记录时间的
type TimerRecord interface {
	TimeWatch()
	GetStartAt(traceId string) (time.Time, error)
	GetFinishAt(traceId string) (time.Time, error)
	GetStartAts() map[string]time.Time
	GetFinishAts() map[string]time.Time
}

// 可记录错误的
type ErrorRecord interface {
	GetErrors() map[string]string
	GetError(traceId string) (string, error)
	SetError(error)
}

// IORecord 可记录输入输出的
type IORecord interface {
	Inputs() map[string][]string
	GetInput(traceId string) ([]string, error)
	Outputs() map[string][]string
	GetOutput(traceId string) ([]string, error)
}

// WithAttribute 可携带属性的
type WithAttribute interface {
	GetAttributes(keys ...string) map[string]interface{}
	GetAttribute(key string) (interface{}, bool)
	SetAttribute(key string, value interface{})
	SetAttributes(values map[string]interface{})
}

type WithOption interface {
	GetOptionWithDefault(key string, defaultValue ...interface{}) interface{}
	GetOption(key string) (interface{}, bool)
	GetOptions() map[string]interface{}
	SetOption(key string, value interface{})
}

// Tracer 可堆栈记录的
type Tracer interface {
	GetTraceIDS() []string
	GetCurrentTraceID() string
	SetTraceID(string)
	BillingRecord
	TimerRecord
	ErrorRecord
	IORecord
}

// WithName 可命名的
type WithName interface {
	GetName() string
	SetName(string)
}

// WithTask 可携带Task
type WithTask interface {
	GetTask() *Task
	SetTask(*Task)
	WithAttribute
}

// Task中的工作
type Job interface {
	Handle(context.Context) error
	WithTask
	WithName
	WithOption
	Tracer
}

// JobHelper trait
type JobHelper struct {
	traces           []string
	name             string
	InputAttributes  map[string]map[string]bool
	OutputAttributes map[string]map[string]bool
	startAts         map[string]time.Time
	finishAts        map[string]time.Time
	bills            map[string]float64
	Errors           map[string]string
	Task             *Task
	Option           map[string]interface{} // 任务配置
}

func (j *JobHelper) GetTraceIDS() []string {
	return j.traces
}

func (j *JobHelper) SetTraceID(traceID string) {
	j.traces = append(j.traces, traceID)
}

func (j *JobHelper) GetCurrentTraceID() string {
	if len(j.traces) == 0 {
		j.SetTraceID(uuid.New().String())
	}
	return j.traces[len(j.traces)-1]
}

func (j *JobHelper) SetName(name string) {
	j.name = name
}

func (j *JobHelper) GetName() string {
	return j.name
}

func (j *JobHelper) GetTask() *Task {
	return j.Task
}

func (j *JobHelper) SetTask(t *Task) {
	j.Task = t
}

func (j *JobHelper) Logger() *logrus.Entry {
	return logrus.WithField("task_id", j.GetTask().ID)
}

func (j *JobHelper) GetAttributes(keys ...string) map[string]interface{} {
	if len(keys) == 0 {
		return j.GetTask().MetaData
	}
	attributes := make(map[string]interface{}, 0)

	j.perpareInputAttributes()
	for _, key := range keys {
		attribute, ok := j.GetAttribute(key)
		if ok {
			attributes[key] = attribute
		}

		j.InputAttributes[j.GetCurrentTraceID()][key] = true
	}
	return attributes
}

func (j *JobHelper) GetAttribute(key string) (interface{}, bool) {
	attribute, ok := j.GetTask().MetaData[key]
	j.perpareInputAttributes()
	j.InputAttributes[j.GetCurrentTraceID()][key] = true
	return attribute, ok
}

func (j *JobHelper) perpareInputAttributes() {
	if j.InputAttributes == nil {
		j.InputAttributes = make(map[string]map[string]bool, 0)
	}
	if j.InputAttributes[j.GetCurrentTraceID()] == nil {
		j.InputAttributes[j.GetCurrentTraceID()] = make(map[string]bool)
	}
}

func (j *JobHelper) perpareOutputAttributes() {
	if j.OutputAttributes == nil {
		j.OutputAttributes = make(map[string]map[string]bool, 0)
	}
	if j.OutputAttributes[j.GetCurrentTraceID()] == nil {
		j.OutputAttributes[j.GetCurrentTraceID()] = make(map[string]bool)
	}
}

func (j *JobHelper) SetAttribute(key string, value interface{}) {
	j.GetTask().MetaData[key] = value
	j.perpareOutputAttributes()
	j.OutputAttributes[j.GetCurrentTraceID()][key] = true
}

func (j *JobHelper) SetAttributes(values map[string]interface{}) {
	j.perpareOutputAttributes()
	for k, v := range values {
		j.SetAttribute(k, v)
		j.OutputAttributes[j.GetCurrentTraceID()][k] = true
	}
}

func (j *JobHelper) Inputs() map[string][]string {
	inputs := make(map[string][]string, 0)
	for k, v := range j.InputAttributes {
		inputs[k] = make([]string, 0)
		for k1 := range v {
			inputs[k] = append(inputs[k], k1)
		}
	}
	return inputs
}

func (j *JobHelper) GetInput(traceId string) ([]string, error) {
	input, ok := j.Inputs()[traceId]
	if !ok {
		return nil, fmt.Errorf("trace %s not found", traceId)
	}
	return input, nil
}

func (j *JobHelper) Outputs() map[string][]string {
	outputs := make(map[string][]string, 0)
	for k, v := range j.OutputAttributes {
		outputs[k] = make([]string, 0)
		for k1 := range v {
			outputs[k] = append(outputs[k], k1)
		}
	}
	return outputs
}

func (j *JobHelper) GetOutput(traceId string) ([]string, error) {
	output, ok := j.Outputs()[traceId]
	if !ok {
		return nil, fmt.Errorf("trace %s not found", traceId)
	}
	return output, nil
}

func (j *JobHelper) TimeWatch() {
	if j.startAts == nil {
		j.startAts = make(map[string]time.Time, 0)
	}
	if j.finishAts == nil {
		j.finishAts = make(map[string]time.Time, 0)
	}

	tid := j.GetCurrentTraceID()

	if startAt, ok := j.startAts[tid]; !ok || startAt.IsZero() {
		j.startAts[tid] = time.Now()
		return
	}

	if finishedAt, ok := j.finishAts[tid]; !ok || finishedAt.IsZero() {
		j.finishAts[tid] = time.Now()
	}
}

func (j *JobHelper) GetStartAts() map[string]time.Time {
	return j.startAts
}

func (j *JobHelper) GetStartAt(traceId string) (time.Time, error) {
	if j.startAts == nil {
		return time.Time{}, nil
	}
	startAt, ok := j.startAts[traceId]
	if !ok {
		return time.Time{}, fmt.Errorf("trace %s not found", traceId)
	}
	return startAt, nil
}

func (j *JobHelper) GetFinishAts() map[string]time.Time {
	return j.finishAts
}

func (j *JobHelper) GetFinishAt(traceId string) (time.Time, error) {
	if j.finishAts == nil {
		return time.Time{}, nil
	}
	finishedAt, ok := j.finishAts[traceId]
	if !ok {
		return time.Time{}, fmt.Errorf("trace %s not found", traceId)
	}
	return finishedAt, nil
}

// func (j *JobHelper) ResetTime() {
// 	if j.startAts == nil {
// 		j.startAts = make(map[string]time.Time, 0)
// 	}
// 	if j.finishAts == nil {
// 		j.finishAts = make(map[string]time.Time, 0)
// 	}

// 	j.startAts[j.GetCurrentTraceID()] = time.Time{}
// 	j.finishAts[j.GetCurrentTraceID()] = time.Time{}
// }

func (j *JobHelper) Billing(bill float64) {
	if j.bills == nil {
		j.bills = make(map[string]float64, 0)
	}
	j.bills[j.GetCurrentTraceID()] += bill
}

func (j *JobHelper) GetBills() map[string]float64 {
	return j.bills
}

func (j *JobHelper) GetBill(traceId string) (float64, error) {
	if j.bills == nil {
		return 0, nil
	}
	bill, ok := j.bills[traceId]
	if !ok {
		return 0, fmt.Errorf("trace %s not found", traceId)
	}
	return bill, nil
}

func (j *JobHelper) SetError(err error) {
	if j.Errors == nil {
		j.Errors = make(map[string]string, 0)
	}
	j.Errors[j.GetCurrentTraceID()] = err.Error()
}

func (j *JobHelper) GetErrors() map[string]string {
	return j.Errors
}

func (j *JobHelper) GetError(traceId string) (string, error) {
	if j.Errors == nil {
		return "", nil
	}
	err, ok := j.Errors[traceId]
	if !ok {
		return "", fmt.Errorf("trace %s not found", traceId)
	}
	return err, nil
}

func (j *JobHelper) GetOptionWithDefault(key string, defaultValue ...interface{}) interface{} {
	option, ok := j.GetOption(key)
	if !ok && len(defaultValue) > 0 {
		j.SetOption(key, defaultValue[0])
		return defaultValue[0]
	}
	return option
}

func (j *JobHelper) GetOption(key string) (interface{}, bool) {
	if j.Option == nil {
		return nil, false
	}
	option, ok := j.Option[key]
	return option, ok
}

func (j *JobHelper) GetOptions() map[string]interface{} {
	if j.Option == nil {
		j.Option = make(map[string]interface{}, 0)
	}
	return j.Option
}

func (j *JobHelper) SetOption(key string, value interface{}) {
	if j.Option == nil {
		j.Option = make(map[string]interface{}, 0)
	}
	j.Option[key] = value
}
