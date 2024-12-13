package search

// SearchParam 搜索引擎选项管理
type SearchParam struct {
	Limit     int
	Size      int
	DepthMode string
	Debug     bool
	Query     string
	TavilySearchParam
}

type TavilySearchParam struct {
	Topic       string
	Days        int
	SearchDepth string
}
