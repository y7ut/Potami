package job

import (
	"context"

	"github.com/y7ut/potami/internal/task"
)

var _ task.Job = (*Job)(nil)

type Job struct {
	task.JobHelper
}

func (j *Job) Handle(context.Context) error {
	return nil
}

func NewBlankJob() *Job {
	return &Job{}
}