package task

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/y7ut/ppool"
)

type Task struct {
	Tracking bool // 是否开启事件监听

	index    int     // heap index
	maxHit   int     // 最大重试次数
	errorHit int     // 错误次数
	level    float64 // 优先任务等级
	Arrived  int     // 进度

	CreatedAt  time.Time // 创建时间
	StartAt    time.Time // 开启时间
	CompleteAt time.Time // 完成时间
	CloseAt    time.Time // 关闭时间

	ID   string // 所执行的流程ID
	Call string // 任务名称

	ErrorStacks []string // 错误堆栈

	MetaData map[string]interface{} // 任务携带的数据
	Option   map[string]interface{} // 任务配置

	JobsPipline                    *list.List    // 任务管道
	CurrentStage, ErrorReopenStage *list.Element // 当前阶段, 错误重试阶段

	completenesNotifyChannel chan float64 // 任务完成度通知广播
	rewindChannel            chan *Task   // rewind channel

	trackingOnce sync.Once // 开启事件监听的同步锁
	callBackOnce sync.Once // 开启回调监听的同步锁

	CallBack *CallBackConfig // 是否开启回调

	DstWorkPool *ppool.Pool[*Task] // 目标工作池

	done chan struct{}

	sync.Mutex
}

// Done
func (task *Task) Done() <-chan struct{} {
	return task.done
}

// Close
func (task *Task) Close() {
	task.CloseAt = time.Now()
	close(task.done)
}

// Weight implement PriorityItem
func (task *Task) Weight() float64 {
	return task.level
}

// Index implement PriorityItem
func (task *Task) Index() int {
	return task.index
}

// SetIndex implement PriorityItem
func (task *Task) SetIndex(index int) {
	task.index = index
}

// SetLevel
func (task *Task) SetLevel(level int) {
	task.level = float64(level)
}

// SetSideEntry 为任务设置一个重试的 Chan
func (task *Task) SetSideEntry(c chan *Task) {
	task.rewindChannel = c
}

// Rool 推进到下一阶段
func (task *Task) Roll() (bool, error) {

	defer task.UpdateCompletePercent()

	// 向下执行
	task.Lock()
	defer task.Unlock()

	next := task.CurrentStage.Next()
	task.Arrived++
	if next == nil {
		// 结束了
		task.CompleteAt = time.Now()
		return false, nil
	}
	// 后面还有任务
	task.CurrentStage = next
	return true, nil
}

// BreakOut 将当前阶段截断为错误阶段
func (task *Task) BreakOut() {
	task.Lock()
	task.ErrorReopenStage = task.CurrentStage
	task.Unlock()
}

// BreakOutWithError 将当前阶段截断为错误阶段, 并记录错误，最后判断是否超过最大重试次数
func (task *Task) BreakOutWithError(err error) bool {

	task.Lock()
	task.ErrorStacks = append(task.ErrorStacks, fmt.Sprintf("#%d-%d:%+v", task.Arrived, task.errorHit, err.Error()))
	task.ErrorReopenStage = task.CurrentStage
	oldhit := task.errorHit
	task.errorHit = oldhit + 1
	task.Unlock()

	return oldhit < task.maxHit
}

// Rewind 重新开始
func (task *Task) Rewind() {
	logrus.WithField("task_id", task.ID).Warning(fmt.Sprintf("worker[%s] rewind try, %d chances to try again", task.Call, task.maxHit-task.errorHit))
	task.rewindChannel <- task
}

// BoardCast 获取任务的广播
func (task *Task) CompleteSate() <-chan float64 {
	return task.completenesNotifyChannel
}

// 获取任务进度
func (task *Task) GetCompleteness() float64 {
	completeness := float64(task.Arrived) / float64(task.JobsPipline.Len())
	return completeness
}

// UpdateCompletePercent 更新任务进度到广播中
func (task *Task) UpdateCompletePercent() {
	task.Lock()
	complete := task.GetCompleteness()
	task.Unlock()
	select {
	case task.completenesNotifyChannel <- complete:
		return
	default:
		go func() {
			task.completenesNotifyChannel <- complete
		}()
	}
}

// ReLevel 重新计算权重level
func (task *Task) ReLevel() {
	task.Lock()
	task.level = task.level + float64(task.errorHit*500)
	task.Unlock()
}

// Health
func (task *Task) Health() float64 {
	return 1 - float64(task.errorHit)/float64(task.maxHit+1)
}

func (task *Task) CallBackListener() { // 回调
	task.callBackOnce.Do(func() {
		logrus.WithField("task_id", task.ID).Debug("callback listener start")
		<-task.Done()

		if err := task.SendCallBack(); err != nil {
			logrus.WithField("task_id", task.ID).Warningf("callback failed, cause by : %s", err)
		}
		logrus.WithField("task_id", task.ID).Debug("callback listener finished")
	})
}

// SendCallBack
func (task *Task) SendCallBack() error {
	if task.CallBack == nil {
		return nil
	}
	// 回调
	if err := sendCallBackRequest(task, task.CallBack.MetaDataBuilder); err != nil {
		return sendCallBackRequest(task, task.CallBack.MetaDataBuilder)
	}
	return nil
}

func (task *Task) Work(ctx context.Context) {
	excute(ctx, task, task.DstWorkPool)
}
