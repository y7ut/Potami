package task

import (
	"github.com/y7ut/potami/pkg/json"
)

// MarshalJSON 任务信息
func (task *Task) MarshalJSON() ([]byte, error) {
	taskInfo := make(map[string]interface{})
	taskInfo["uuid"] = task.ID
	taskInfo["name"] = task.Call
	taskInfo["level"] = task.level
	taskInfo["length"] = task.JobsPipline.Len()
	taskInfo["allow_retry_count"] = task.maxHit + 1 // 最大错误次数，包含第一次
	taskInfo["start_at"] = task.StartAt
	taskInfo["created_at"] = task.CreatedAt
	taskInfo["complete_at"] = task.CompleteAt
	taskInfo["close_at"] = task.CloseAt
	taskInfo["meta_data"] = task.MetaData
	taskInfo["error_stacks"] = task.ErrorStacks
	taskInfo["traces"] = GetTracesFromTask(task)
	return json.Marshal(taskInfo)
}
