package task

import (
	"context"
)

// Task中的工作
type Job interface {
	Handle(context.Context) error
	SetName(string)
	GetName() string
	GetTask() *Task
	SetTask(*Task)
}

// JobHelper trait
type JobHelper struct {
	name string
	*Task
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

func (j *JobHelper) GetAttributes(keys ...string) map[string]interface{} {
	if len(keys) == 0 {
		return j.GetTask().MetaData
	}
	attributes := make(map[string]interface{}, 0)
	for _, key := range keys {
		attribute, ok := j.GetAttribute(key)
		if ok {
			attributes[key] = attribute
		}
	}
	return attributes
}

func (j *JobHelper) GetAttribute(key string) (interface{}, bool) {
	attribute, ok := j.GetTask().MetaData[key]
	return attribute, ok
}

func (j *JobHelper) SetAttribute(key string, value interface{}) {
	j.GetTask().MetaData[key] = value
}

func (j *JobHelper) SetAttributes(values map[string]interface{}) {
	for k, v := range values {
		j.SetAttribute(k, v)
	}
}

func (j *JobHelper) GetOption(key string, defaultValue ...interface{}) interface{} {
	option, ok := j.GetTask().Option[key]
	if !ok && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return option
}

func (j *JobHelper) SetOption(key string, value interface{}) {
	j.GetTask().Option[key] = value
}
