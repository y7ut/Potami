package op

import (
	"log"
	"time"

	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/ppool"
	"github.com/y7ut/ppool/option"
)

var TaskPool *ppool.Pool[*task.Task]

// preparePool 初始化协程池
func preparePool() *ppool.Pool[*task.Task] {
	pool, err := ppool.CreatePool[*task.Task](
		option.WithMaxWorkCount(conf.GoroutinePool.MaxWorkers),
		option.WithTimeout(time.Duration(conf.GoroutinePool.Timeout)*time.Second),
		option.WithMaxIdleWorkerDuration(time.Duration(conf.GoroutinePool.MaxIdleDuration)*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}
	return pool
}
