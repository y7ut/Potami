package conf

import (
	"context"
	"sync"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

var (
	milvusClient *milvusclient.Client
	milvusOnce   sync.Once
)

func MilvusClient() *milvusclient.Client {
	milvusOnce.Do(func() {
		var err error
		milvusClient, err = NewMilvus()
		if err != nil {
			panic(err)
		}
	})
	return milvusClient
}

func NewMilvus() (*milvusclient.Client, error) {
	return milvusclient.New(context.Background(), &milvusclient.ClientConfig{
		Address:  Milvus.Address,
		Username: Milvus.Username,
		Password: Milvus.Password,
	})
}
