package dispatch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/ppool"
)

var defaultRetryTimeGenerator = func(i int) time.Duration {
	return time.Duration(i*i*100) * time.Millisecond
}

type Kernal struct {
	IntervalTime  time.Duration
	UpstreamQueue *task.RetryAbleQueue
	Pool          *ppool.Pool[*task.Task]

	TaskTrackers []task.TaskTracker
	Done         chan struct{}

	StartOnce          *sync.Once
	RetryTimeGenerator func(i int) time.Duration
}

func NewDispatchKernal(intervalTime time.Duration, upstreamQueue *task.RetryAbleQueue, pool *ppool.Pool[*task.Task], trackers []task.TaskTracker) *Kernal {
	return &Kernal{
		IntervalTime:       intervalTime,
		UpstreamQueue:      upstreamQueue,
		Pool:               pool,
		TaskTrackers:       trackers,
		Done:               make(chan struct{}),
		StartOnce:          &sync.Once{},
		RetryTimeGenerator: defaultRetryTimeGenerator,
	}
}

func (w *Kernal) SetRetryTimeGenerator(f func(i int) time.Duration) {
	w.RetryTimeGenerator = f
}

func (w *Kernal) Start(ctx context.Context) {
	w.StartOnce.Do(func() {
		go w.start(ctx)
		logrus.Info("dispatch worker start...")
	})
}

func (w *Kernal) start(ctx context.Context) {
	defer func() {
		logrus.Info("dispatch worker stoping...")
		time.Sleep(2 * time.Second)
		logrus.Info("dispatch worker stopped")
	}()
	upstreamTask := w.UpstreamQueue.ReadTask(ctx)
	for {
		select {
		case r := <-upstreamTask:
			if r == nil {
				return
			}
			task, len := r()
			i := 1
			ok, err := w.Pool.Serve(task)
			if err != nil {
				return
			}

			for !ok {
				// 当前等待的一个，有可能已经pop出来了但是还没有进入reader的一个，再加上队列中剩余的数量
				// waitTaskCount := w.UpstreamQueue.SafeLen() + 2
				waitDuration := w.RetryTimeGenerator(i)
				logrus.Warning(fmt.Sprintf("[Dispatch]pool is busy, at least %d tasks is waiting, next retry time is: %dms", len+1, w.RetryTimeGenerator(i)/time.Millisecond))
				time.Sleep(waitDuration)
				i++
				//
				ok, err = w.Pool.Serve(task)
				if err != nil {
					return
				}
			}

			go task.CallBackListener()
			task.SetTrackers(w.TaskTrackers...)

			time.Sleep(w.IntervalTime)
		case <-ctx.Done():
			return
		case <-w.Done:
			return
		}
	}
}

func (w *Kernal) Stop() {
	close(w.Done)
}
