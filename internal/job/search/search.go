package search

import (
	"context"
	"fmt"

	"github.com/y7ut/potami/internal/task"
)

// SearchService API搜索
type SearchService struct {
	task.JobHelper

	Engine    Engine
	Options   []SearchEngineOption
	BlockSize int
	DepthMode bool

	QueryField  string
	OutputField string
}

func (ss *SearchService) Handle(ctx context.Context) error {
	query, err := perpareQuery(ss)
	if err != nil {
		ss.Logger().WithError(err).Error("search error")
		return err
	}
	queryString := query

	result, err := ss.Engine.Search(ctx, queryString, ss.Options...)
	if err != nil {
		ss.Logger().WithError(err).Error("search error")
		return err
	}
	output := DocumentsCompress(result, ss.BlockSize, ss.DepthMode)
	ss.SetAttribute(ss.OutputField, output)
	mode := "depth"
	if ss.DepthMode {
		mode = "flat"
	}
	ss.Logger().WithField(ss.OutputField, output).Debugf("search finish with block size: %d, result size: %d and mode: %v", ss.BlockSize, len(result), mode)
	return nil
}

func perpareQuery(t *SearchService) (string, error) {
	query, ok := t.GetAttribute(t.QueryField)
	if !ok {
		return "", fmt.Errorf("search error: query field %s not found", t.QueryField)
	}
	queryString, ok := query.(string)
	if !ok {
		return "", fmt.Errorf("search error: query field %s not string", t.QueryField)
	}

	return queryString, nil
}
