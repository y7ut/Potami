package tracker

import (
	"context"
	"fmt"

	"github.com/redis/rueidis"
	"github.com/sirupsen/logrus"
	"github.com/y7ut/potami/internal/task"
)

const TRACKER_KEY = "POTAMI:T:%s"

type RedisTracker struct {
	redis rueidis.Client
}

func NewRedisTracker(redis rueidis.Client) *RedisTracker {
	return &RedisTracker{redis: redis}
}

// Tracker RedisTracker
func (rt *RedisTracker) Notice(ctx context.Context, complete float64, t *task.Task) {
	if rt.redis == nil {
		return
	}
	AddCommand := rt.redis.B().
		Sadd().
		Key(fmt.Sprintf(TRACKER_KEY, t.ID)).
		Member(fmt.Sprintf("%.2f", complete*100)).
		Build()
	Expirecommand := rt.redis.B().
		Expire().
		Key(fmt.Sprintf(TRACKER_KEY, t.ID)).
		Seconds(3600).
		Build()

	logrus.WithField("task_id", t.ID).Debug(fmt.Sprintf("%s [redis]", fmt.Sprintf("Worker[%s]的工作进度已经到了------- %.2f%%", t.Call, complete*100)))
	res := rt.redis.DoMulti(ctx, AddCommand, Expirecommand)
	for _, r := range res {
		if r.Error() != nil {
			logrus.WithField("task_id", t.ID).Error(fmt.Sprintf("redis tracker error: %s", r.Error()))
		}
	}
}
