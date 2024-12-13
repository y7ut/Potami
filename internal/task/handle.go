package task

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/y7ut/ppool"
)

func excute(ctx context.Context, stream *Task, wp *ppool.Pool[*Task]) {

	resultChannel := make(chan error)

	if stream.ErrorReopenStage != nil {
		logrus.WithField("task_id", stream.ID).Debug("location error stage")

		stream.CurrentStage = stream.ErrorReopenStage
		stream.ErrorReopenStage = nil
	}

	if stream.CurrentStage == nil {
		panic("current stage is nil")
	}

	j := stream.CurrentStage.Value.(Job)

	if stream.Arrived == 0 {
		stream.StartAt = time.Now()
	}

	go func() {
		j.SetTraceID(uuid.New().String())
		j.TimeWatch()
		currentResult := j.Handle(ctx)
		resultChannel <- currentResult
		j.TimeWatch()
	}()

	// 等待结果，或超时（wp.timeout）
	select {
	case err := <-resultChannel:
		if err != nil {
			j.SetError(err)
			// 记录错误爆发
			if !stream.BreakOutWithError(err) {
				logrus.WithField("task_id", stream.ID).Error(fmt.Sprintf("Worker[%s] failed too many time", stream.Call))
				stream.Close()
				return
			}

			// 重来
			stream.Rewind()
			return
		}

		// 继续
		if roll, err := stream.Roll(); !roll || err != nil {
			stream.Close()
			return
		}

		ok, err := wp.Serve(stream)
		if err != nil {
			// Pool 出现问题已经关闭了
			return
		}

		if !ok {
			logrus.Warning(fmt.Sprintf("[POOL INNER]pool is busy, Worker[%s] will be wait", stream.Call))
			// 无资源继续执行，出现阻塞及时挂起等待之后再重试
			stream.BreakOut()
			// 重来
			stream.Rewind()
			logrus.Warning(fmt.Sprintf("[POOL INNER]pool is busy, Worker[%s] send rewind signal, will be rewind", stream.Call))
			return
		}

		return
	case <-ctx.Done():
		// 注意这里如果有一个job超时，接下来就会停止了，所以这里尽量将超时变成一个错误
		logrus.WithField("task_id", stream.ID).Error(fmt.Sprintf("task[%s] job handle timeout", stream.Call))
		// 记录错误爆发
		if !stream.BreakOutWithError(ctx.Err()) {
			logrus.WithField("task_id", stream.ID).Error(fmt.Sprintf("Worker[%s] inner failed too many time ", stream.Call))
			stream.Close()
			return
		}

		// 重来
		stream.Rewind()
		return
	}
}
