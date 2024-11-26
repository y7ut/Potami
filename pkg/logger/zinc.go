package logger

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

const defaultURI = "/es/_bulk"

type zincLogData struct {
	Time    time.Time     `json:"time"`
	Level   logrus.Level  `json:"level"`
	Message string        `json:"message"`
	Data    logrus.Fields `json:"data"`
}

type zincLogIndex struct {
	Index map[string]string `json:"index"`
}

type zincLogHook struct {
	host     string
	index    string
	user     string
	password string
}

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
		fmt.Println(err.Error())
	}

	return nil
}

func (h *zincLogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func NewZincLogHook(zincUrl string, index string) (*zincLogHook, error) {
	// zine_url like `zinc://user:password@127.0.0.1:4080`
	regex := `zinc://([^:]+):([^@]+)@([a-z0-9\\._-]+):([0-9]+)`
	re := regexp.MustCompile(regex)
	result := re.FindStringSubmatch(zincUrl)
	if len(result) == 5 {
		return &zincLogHook{
			host:     fmt.Sprintf("http://%s:%s", result[3], result[4]),
			index:    index,
			user:     result[1],
			password: result[2],
		}, nil
	}
	return nil, fmt.Errorf("zinc url error")

}
