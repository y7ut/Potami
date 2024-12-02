package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/y7ut/collection"
	"github.com/y7ut/potami/internal/op"
	"github.com/y7ut/potami/internal/task"
)

const (
	StatusUncomplete = iota
	StatusComplete
)

type TaskItem struct {
	Schedule    string    `json:"schedule"`
	Health      string    `json:"health"`
	Name        string    `json:"name"`
	TaskID      string    `json:"task_id"`
	ErrorStacks []string  `json:"error_stacks"`
	CreatedAt   time.Time `json:"created_at"`  // 创建时间
	StartAt     time.Time `json:"start_at"`    // 开启时间
	CompleteAt  time.Time `json:"complete_at"` // 完成时间
	CloseAt     time.Time `json:"close_at"`    // 关闭时间
}

func List(c *gin.Context) {

	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}

	ids := c.Query("ids")
	tasks_id := make([]string, 0)
	if ids != "" {
		tasks_id = append(tasks_id, strings.Split(ids, ",")...)
	}

	tasks := make([]*task.Task, 0)
	op.Tasks.Range(func(key string, value *task.Task) bool {
		if len(tasks_id) == 0 {
			tasks = append(tasks, value)
			return true
		}

		for _, id := range tasks_id {
			if value.ID == id {
				tasks = append(tasks, value)
				return true
			}
		}
		return true
	})

	order := c.Query("order")

	sortFunc := func(i, j *task.Task) bool {
		return i.CreatedAt.UnixMicro() > j.CreatedAt.UnixMicro()
	}
	if order == "complete_at" {
		sortFunc = func(i, j *task.Task) bool {
			return i.CompleteAt.UnixMicro() > j.CompleteAt.UnixMicro()
		}
	}

	collextions := collection.New(tasks).Sort(sortFunc)

	status := c.Query("status")
	if status != "" {
		StatusCode, _ := strconv.Atoi(status)
		switch StatusCode {
		case StatusComplete:
			collextions = collextions.Filter(func(task *task.Task) bool {
				return task.GetCompleteness() == 1
			})
		case StatusUncomplete:
			collextions = collextions.Filter(func(task *task.Task) bool {
				return task.GetCompleteness() != 1
			})
		}
	}

	res := make([]TaskItem, 0)
	for i, t := range collextions.All() {
		if i >= (page-1)*limit && i < limit*page {
			res = append(res, TaskItem{
				TaskID:      t.ID,
				Name:        t.Call,
				Schedule:    fmt.Sprintf("%.2f", t.GetCompleteness()),
				Health:      fmt.Sprintf("%.2f", t.Health()),
				ErrorStacks: t.ErrorStacks,
				CreatedAt:   t.CreatedAt,
				StartAt:     t.StartAt,
				CompleteAt:  t.CompleteAt,
				CloseAt:     t.CloseAt,
			})
		}
	}

	c.JSON(200, gin.H{
		"tasks": res,
		"total": collextions.Len(),
	})
}
