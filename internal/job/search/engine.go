package search

import "context"

// SearchEngine 搜索引擎
type Engine interface {
	Search(ctx context.Context, query string, options ...SearchEngineOption) ([]Document, error)
}

// SearchEngineOptionsManager 搜索引擎选项管理
type SearchEngineOptionsManager struct {
	Limit   int
	Debug   bool
	options map[string]interface{}
}

func (seom *SearchEngineOptionsManager) Get(name string, defaultValue any) any {
	if v, ok := seom.options[name]; ok {
		return v
	}
	return defaultValue
}

func DefaultOptionsManager() *SearchEngineOptionsManager {
	return &SearchEngineOptionsManager{
		Limit:   5,
		options: make(map[string]interface{}),
	}
}

func WithLimit(limit int) SearchEngineOption {
	return func(seom *SearchEngineOptionsManager) {
		seom.Limit = limit
	}
}

func WithDebug(debug bool) SearchEngineOption {
	return func(seom *SearchEngineOptionsManager) {
		seom.Debug = debug
	}
}

func WithOption(name string, value any) SearchEngineOption {
	return func(seom *SearchEngineOptionsManager) {
		seom.options[name] = value
	}
}

// SearchEngineOption 搜索引擎选项
type SearchEngineOption func(seom *SearchEngineOptionsManager)
