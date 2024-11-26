package task

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// MaxEventListenTime 最大事件监听时间
const MaxEventListenTime = 60

type TaskTracker interface {
	Notice(ctx context.Context, status float64, t *Task)
}

// 重新计算权重level
func (stream *Task) SetTrackers(listener ...TaskTracker) {
	// 开启进度追踪
	if stream.Tracking {
		stream.trackingOnce.Do(func() {
			logrus.WithField("task_id", stream.ID).Info(fmt.Sprintf("start tracking task: %s", stream.Call))
			go tracking(stream, listener...)
		})
	}
}

// Tracking 开启一个task 对于其生命周期的事件通知接受者
func tracking(stream *Task, trackers ...TaskTracker) {
	// 事件监听开启
	ctx, canel := context.WithTimeout(context.Background(), time.Duration(stream.JobsPipline.Len()*MaxEventListenTime)*time.Second)
	defer canel()

	for {
		select {
		case <-stream.Done():
			logrus.WithField("task_id", stream.ID).Debug("stream done....")
			for _, t := range trackers {
				t.Notice(ctx, 1, stream)
			}
			return
		case arrviedState := <-stream.CompleteSate():
			// 事件监听
			for _, t := range trackers {
				t.Notice(ctx, arrviedState, stream)
			}
			if arrviedState >= 1 || stream.Arrived == stream.JobsPipline.Len() {
				// 任务完成
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
