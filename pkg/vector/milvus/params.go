package milvus

import (
	"encoding/base64"
	"fmt"
)

type MilvusParams struct {
	Dimensions      int64
	Collection      string
	Partition       string
	useBM25         bool
	useContextEmbed bool
}

func NewMilvusParams() *MilvusParams {
	return &MilvusParams{}
}

func (p *MilvusParams) Hash() string {
	paramsStr := fmt.Sprintf("%s_%s_%t_%t", p.Collection, p.Partition, p.useBM25, p.useContextEmbed)
	base64Str := base64.StdEncoding.EncodeToString([]byte(paramsStr))

	return base64Str
}

type MilvusParamsOption func(*MilvusParams)

func WithCollection(collection string) MilvusParamsOption {
	return func(m *MilvusParams) {
		m.Collection = collection
	}
}

func WithPartition(partition string) MilvusParamsOption {
	return func(m *MilvusParams) {
		m.Partition = partition
	}
}

func WithUseBM25(useBM25 bool) MilvusParamsOption {
	return func(m *MilvusParams) {
		m.useBM25 = useBM25
	}
}

func WithUseContextEmbed(useContextEmbed bool) MilvusParamsOption {
	return func(m *MilvusParams) {
		m.useContextEmbed = useContextEmbed
	}
}

func WithDimensions(dimensions int64) MilvusParamsOption {
	return func(m *MilvusParams) {
		m.Dimensions = dimensions
	}
}
