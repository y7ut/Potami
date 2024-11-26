package conf

import (
	"log"
	"sync"

	"github.com/redis/rueidis"
)

var (
	redisClient rueidis.Client
	onceRedis   sync.Once
)

func GetRedisClient() rueidis.Client {
	onceRedis.Do(func() {
		client, err := rueidis.NewClient(rueidis.ClientOption{
			InitAddress:  RedisConf.Address,
			Password:     RedisConf.Password,
			SelectDB:     RedisConf.Select,
			DisableCache: RedisConf.DisableCache,
		})
		if err != nil {
			log.Fatalf("create a redis client failed: %s", err)
		}
		redisClient = client
	})
	return redisClient
}
