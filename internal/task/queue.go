package task

import (
	"container/heap"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/y7ut/potami/pkg/queue"
)

// EventLoopInterval 事件循环间隔(豪秒)
const (
	EventLoopInterval = 100
)

// priorityQueue 优先级队列
type RetryAbleQueue struct {
	pq *queue.PriorityQueue[*Task]

	close        bool // 已经结束
	sideEntry    chan *Task
	reader       chan func() (*Task, int) //
	done         chan struct{}
	readyToRead  sync.Once
	readyToRetry sync.Once

	MaxSize int
	sync.Mutex
}

// NewTaskQueue
func NewTaskQueue(size int) *RetryAbleQueue {

	tasks := make([]*Task, 0)
	priorityQueueInstance := queue.NewPriorityQueue(tasks)

	heap.Init(priorityQueueInstance)

	var rewindChannel = make(chan *Task, 200)
	var reader = make(chan func() (*Task, int))
	var done = make(chan struct{})
	queue := &RetryAbleQueue{
		pq:           priorityQueueInstance,
		sideEntry:    rewindChannel,
		reader:       reader,
		done:         done,
		close:        false,
		Mutex:        sync.Mutex{},
		readyToRead:  sync.Once{},
		readyToRetry: sync.Once{},

		MaxSize: size,
	}

	return queue
}

// 新增任务到优先级队列
func (q *RetryAbleQueue) PushTask(t *Task) (int, error) {
	var count int

	q.Lock()
	defer q.Unlock()

	if q.close {
		return count, errors.New("queue has closed")
	}

	count = q.pq.Len()
	if count >= q.MaxSize {
		return count, fmt.Errorf("queue is full with %d tasks", count)
	}

	t.SetSideEntry(q.sideEntry)
	heap.Push(q.pq, t)

	return count, nil
}

func (q *RetryAbleQueue) pushTaskUnsafe(t *Task) {

	q.Lock()
	defer q.Unlock()

	if q.close {
		return
	}

	t.SetSideEntry(q.sideEntry)
	heap.Push(q.pq, t)

}

// 更新任务
func (q *RetryAbleQueue) Update(t *Task, name string, level int) {
	t.Call = name
	t.SetLevel(level)

	q.Lock()
	heap.Fix(q.pq, int(t.Index()))
	q.Unlock()
}

// popTask 从优先级队列中推出元素
func (q *RetryAbleQueue) popTask() (*Task, int) {
	q.Lock()
	item := heap.Pop(q.pq).(*Task)
	l := q.pq.Len()
	q.Unlock()
	return item, l
}

// SafeLen 优先级队列长度
func (q *RetryAbleQueue) SafeLen() int {
	q.Lock()
	len := q.pq.Len()
	q.Unlock()
	return len
}

// ReadTask 获取一个 Channel 用于读取
func (q *RetryAbleQueue) ReadTask(ctx context.Context) <-chan func() (*Task, int) {
	// 启动回收通道的channel，用于2次进入优先级队列(只有)
	q.readyToRetry.Do(func() {
		go listeningRewindTask(ctx, q)
		logrus.Info("pq rewind channel has started")
	})
	q.readyToRead.Do(func() {
		// 开启一个协程用于读取Reading的通道
		go readingPushedTask(ctx, q)
		logrus.Info("pq reading channel has started")
	})
	// 返回通道
	return q.reader
}

// readingPushedTask 循环读取队列中的任务
func readingPushedTask(ctx context.Context, pq *RetryAbleQueue) {
	ticker := time.NewTicker(time.Duration(EventLoopInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logrus.Warning("pq reading channel cancel")
			return
		case <-pq.done:
			logrus.Info("pq reading channel close")
			return
		case <-ticker.C:
			// 检查队列是否有数据
			if pq.SafeLen() > 0 {
				select {
				case pq.reader <- pq.popTask:
				case <-ctx.Done():
					logrus.Warning("pq reading stopped by context during task push")
					return
				case <-pq.done:
					logrus.Info("pq reading stopped by done signal during task push")
					return
				}
			}
		}
	}
}

// listeningRewindTask 循环监听重试队列中的任务，重新评定优先级，然后重新投递到队列中
func listeningRewindTask(ctx context.Context, pq *RetryAbleQueue) {
	for {
		select {
		case <-pq.done:
			// queue 停止关闭监听
			logrus.Info("pq rewind channel closed")
			return
		case <-ctx.Done():
			// 外部上下文取消，不在继续读取
			logrus.Warning("pq rewind channel cancel")
		default:
			if pq.close {
				return
			}
			badTask := <-pq.sideEntry
			if badTask == nil {
				continue
			}
			logrus.WithField("task_id", badTask.ID).Warning(fmt.Sprintf("task[%s] rewind from side entry", badTask.Call))
			// 重新制定任务的调度数值
			badTask.ReLevel()
			pq.pushTaskUnsafe(badTask)
			time.Sleep(time.Duration(EventLoopInterval) * time.Millisecond)
		}
	}
}

func (pq *RetryAbleQueue) Close() {
	pq.close = true
	close(pq.done)
	close(pq.reader)
	logrus.Info("queue has closed")
}
