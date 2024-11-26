package task

import (
	"container/list"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/y7ut/ppool"
)

// 初始化一个工厂
func NewFactory(dstPool *ppool.Pool[*Task], retry int, tracking bool, roadMap map[string]func() []Job, levelMap map[string]int, descriptionMap map[string]string) *StreamTaskFactory {
	return &StreamTaskFactory{
		RoadMap:        roadMap,
		LevelMap:       levelMap,
		num:            atomic.Uint64{},
		RetryNum:       retry,
		TrackingTask:   tracking,
		DescriptionMap: descriptionMap,
		dstPool:        dstPool,
	}
}

// 流工厂
type StreamTaskFactory struct {
	RetryNum       int
	TrackingTask   bool
	RoadMap        map[string]func() []Job
	LevelMap       map[string]int
	DescriptionMap map[string]string
	num            atomic.Uint64
	dstPool        *ppool.Pool[*Task]
}

// 工厂方法
func (f *StreamTaskFactory) createStreamTask(call string, level int, jobs []Job, id ...uuid.UUID) (*Task, error) {
	f.num.Add(1)

	var newTaskID uuid.UUID
	if len(id) > 0 {
		newTaskID = id[0]
	} else {
		newTaskID = uuid.NewV4()
	}

	stream := &Task{
		Tracking:                 f.TrackingTask,
		ID:                       newTaskID.String(),
		Call:                     call,
		level:                    float64(level),
		errorHit:                 0,
		maxHit:                   f.RetryNum,
		completenesNotifyChannel: make(chan float64),
		done:                     make(chan struct{}),
		CreatedAt:                time.Now(),
		DstWorkPool:              f.dstPool,
	}

	jobslist := list.New()
	for _, j := range jobs {
		j.SetTask(stream)
		jobslist.PushBack(j)
	}

	stream.JobsPipline = jobslist
	stream.CurrentStage = stream.JobsPipline.Front()
	stream.MetaData = make(map[string]interface{})
	stream.Option = make(map[string]interface{})
	// task.UpdateCompletePercent()
	logrus.WithField("task_id", stream.ID).Info(fmt.Sprintf("create stream: %s ", call))
	return stream, nil
}

// 创建工厂的内置任务
func (f *StreamTaskFactory) Create(name string, call string) (*Task, error) {
	name = strings.ToLower(name)
	road := f.RoadMap[name]
	level := f.LevelMap[name]
	if road == nil || level == 0 {
		return nil, fmt.Errorf("没有找到该预定义任务[%s]", name)
	}
	return f.createStreamTask(call, level, road())
}

// 创建工厂的内置任务
func (f *StreamTaskFactory) CreateWithUuid(name string, call string, id ...uuid.UUID) (*Task, error) {
	name = strings.ToLower(name)
	road := f.RoadMap[name]
	level := f.LevelMap[name]
	if road == nil || level == 0 {
		return nil, fmt.Errorf("没有找到该预定义任务[%s]", name)
	}
	return f.createStreamTask(call, level, road(), id...)
}
