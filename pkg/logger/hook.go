package logger

import (
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

// UIDHook 为每条日志增加一个唯一标识
type UIDHook struct {
	key string
}

// Levels 需要覆盖的日志级别
func (h UIDHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire Implements logrus.Hook
func (h UIDHook) Fire(entry *logrus.Entry) error {
	logid := uuid.NewV4().String()[:8]
	entry.Data[h.key] = logid
	return nil
}

// NewUIDHook 初始化一个唯一标识, 可以设置其键值名称
func NewUIDHook(key ...string) *UIDHook {
	hook := &UIDHook{
		key: "log_id",
	}
	if len(key) > 0 {
		hook.key = key[0]
	}
	return hook
}
