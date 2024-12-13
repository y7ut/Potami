package logger

import (
	"encoding/json"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

const defaultURI = "/es/_bulk"

// zincLogData
type zincLogData struct {
	Time    time.Time     `json:"time"`
	Level   logrus.Level  `json:"level"`
	Message string        `json:"message"`
	Data    logrus.Fields `json:"data"`
}

// zincLogIndex
type zincLogIndex struct {
	Index map[string]string `json:"index"`
}

// zincLogHook
type zincLogHook struct {
	host     string
	index    string
	user     string
	password string
}

// Fire Implements logrus.Hook
// TODO: Async use channel push log
func (h *zincLogHook) Fire(entry *logrus.Entry) error {
	index := &zincLogIndex{
		Index: map[string]string{
			"_index": h.index,
		},
	}
	indexBytes, _ := json.Marshal(index)

	data := &zincLogData{
		Time:    entry.Time,
		Level:   entry.Level,
		Message: entry.Message,
		Data:    entry.Data,
	}
	dataBytes, _ := json.Marshal(data)

	logStr := string(indexBytes) + "\n" + string(dataBytes) + "\n"
	client := resty.New()
	url := h.host + defaultURI
	if _, err := client.SetDisableWarn(true).R().
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(h.user, h.password).
		SetBody(logStr).
		Post(url); err != nil {
		return err
	}

	return nil
}

// Levels 需要覆盖的日志级别
func (h *zincLogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// NewZincLogHook 创建Zinc日志Hook
func NewZincLogHook(address, user, password, index string) *zincLogHook {
	return &zincLogHook{
		host:     address,
		index:    index,
		user:     user,
		password: password,
	}
}
