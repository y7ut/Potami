package op

import (
	"sync"
	"time"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/pkg/vector/milvus"
)

var (
	milvusConnectionPool *milvus.MilvusConnectionPool

	milvusOnce sync.Once
)

func MilvusConnectionPool() *milvus.MilvusConnectionPool {
	milvusOnce.Do(func() {
		milvusConnectionPool = milvus.NewMilvusConnectionPool(
			5,
			300*time.Second,
			60*time.Second,
			&milvusclient.ClientConfig{
				Address:  conf.Milvus.Address,
				Username: conf.Milvus.Username,
				Password: conf.Milvus.Password,
			},
		)
		milvusConnectionPool.Start()
	})
	return milvusConnectionPool
}

// StopMilvusConnectionPool
func StopMilvusConnectionPool() {
	MilvusConnectionPool().Stop()

}
