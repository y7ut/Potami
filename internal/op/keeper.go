package op

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/internal/task"

	"github.com/go-resty/resty/v2"
)

const (
	DefaultCrontabInterval = 10 * time.Second
	ZincSearchBlukUri      = "/api/_bulk"
)

var (
	TaskKeeperInstance *TaskKeeper
)

func StartTaskKeeper(ctx context.Context) {
	if TaskKeeperInstance != nil {
		TaskKeeperInstance.start(ctx)
		return
	}
	var err error
	TaskKeeperInstance, err = newTaskKeeper(conf.ZincConf.Address, conf.ZincConf.User, conf.ZincConf.Pass, conf.ZincConf.DataIndex, DefaultCrontabInterval)
	if err != nil {
		logrus.Fatalf("task keeper init error: %v", err)
	}
	TaskKeeperInstance.start(ctx)
}

type TaskKeeper struct {
	host  string
	user  string
	pass  string
	index string

	crontabInterval time.Duration

	once sync.Once
}

func newTaskKeeper(zincAddress, user, pass, index string, crontabInterval time.Duration) (*TaskKeeper, error) {
	return &TaskKeeper{
		host:            zincAddress,
		user:            user,
		pass:            pass,
		index:           index,
		crontabInterval: crontabInterval,
	}, nil
}

func (tk *TaskKeeper) start(ctx context.Context) {
	tk.once.Do(func() {
		go func() {
			ticker := time.NewTicker(tk.crontabInterval)
			for {
				select {
				case <-ctx.Done():
					logrus.Warning("task keeper stopped by context")
					return
				case <-ticker.C:
					tasks := checkinCompletedTask()
					if len(tasks) == 0 {
						continue
					}
					tk.keep(tasks)
				}
			}
		}()
	})
}

func checkinCompletedTask() []*task.Task {
	completedTask := make([]*task.Task, 0)

	Tasks.Range(func(key string, value *task.Task) bool {
		if !value.CompleteAt.IsZero() || !value.CloseAt.IsZero() {
			completedTask = append(completedTask, value)
		}
		return true
	})

	return completedTask
}

// keep 保留已完成的任务到 Data Index
func (tk *TaskKeeper) keep(tasks []*task.Task) error {

	indexConfig := &struct {
		Index map[string]string `json:"index"`
	}{
		Index: map[string]string{
			"_index": "data",
		},
	}

	indexBytes, err := json.Marshal(indexConfig)
	if err != nil {
		return fmt.Errorf("task keeper marshal error: %v", err)
	}
	dataBytes, err := json.Marshal(tasks)
	if err != nil {
		return fmt.Errorf("task keeper marshal error: %v", err)
	}

	client := resty.New()
	if _, err := client.SetDisableWarn(true).R().
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(tk.user, tk.pass).
		SetBody(fmt.Sprintf("%s\n%s\n", string(indexBytes), string(dataBytes))).
		Post(fmt.Sprintf("%s%s", tk.host, ZincSearchBlukUri)); err != nil {
		return err
	}
	logrus.Debugf("task keeper keep %d tasks", len(tasks))
	return nil
}
