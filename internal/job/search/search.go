package search

import (
	"context"
	"fmt"

	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/pkg/param"
)

// SearchService API搜索
type SearchService struct {
	task.JobHelper

	Engine Engine

	QueryField  string
	OutputField string
}

func (ss *SearchService) Handle(ctx context.Context) error {

	p, err := perpareParam(ss)
	if err != nil {
		ss.Logger().WithError(err).Error("search error")
		return err
	}

	result, err := ss.Engine.Search(ctx, p)
	if err != nil {
		ss.Logger().WithError(err).Error("search error")
		return err
	}

	output := DocumentsCompress(result, p.Size, p.DepthMode)

	ss.SetAttribute(ss.OutputField, output)
	ss.Logger().WithField(ss.OutputField, output).Debugf("search finish with block size: %d, result size: %d and mode: %v", p.Size, len(result), p.DepthMode)
	return nil
}

// perpareParam
func perpareParam(ss *SearchService) (*SearchParam, error) {
	searchParam := &SearchParam{}
	query, ok := ss.GetAttribute(ss.QueryField)
	if !ok {
		return nil, fmt.Errorf("search error: query field %s not found", ss.QueryField)
	}
	queryString, ok := query.(string)
	if !ok {
		return nil, fmt.Errorf("search error: query field %s not string", ss.QueryField)
	}
	searchParam.Query = queryString

	if err := param.Assign(&searchParam.Debug, ss.GetOptionWithDefault("debug", false)); err != nil {
		return nil, fmt.Errorf("search debug type error, error: %v", err)
	}

	if err := param.Assign(&searchParam.Limit, ss.GetOptionWithDefault("limit", 10)); err != nil {
		return nil, fmt.Errorf("search limit type error, error: %v", err)
	}

	if err := param.Assign(&searchParam.Size, ss.GetOptionWithDefault("block_size", 5)); err != nil {
		return nil, fmt.Errorf("search block size type error, error: %v", err)
	}

	if err := param.Assign(&searchParam.DepthMode, ss.GetOptionWithDefault("depth_mode", "depth")); err != nil {
		return nil, fmt.Errorf("search depth mode type error, error: %v", err)
	}

	if err := param.Assign(&searchParam.Topic, ss.GetOptionWithDefault("topic", TopicGeneral)); err != nil {
		return nil, fmt.Errorf("search topic type error, error: %v", err)
	}

	if err := param.Assign(&searchParam.Days, ss.GetOptionWithDefault("days", 3)); err != nil {
		return nil, fmt.Errorf("search days type error, error: %v", err)
	}

	if err := param.Assign(&searchParam.SearchDepth, ss.GetOptionWithDefault("search_depth", DepthBasic)); err != nil {
		return nil, fmt.Errorf("search search depth type error, error: %v", err)
	}

	return searchParam, nil
}
