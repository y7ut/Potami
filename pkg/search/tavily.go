package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/internal/document"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/pkg/param"
)

const (
	TopicGeneral         = "general"
	TopicNews            = "news"
	DepthBasic           = "basic"
	DepthAdvanced        = "advanced"
	DefaultDays          = 7
	TavilySearchEndpoint = "https://api.tavily.com/search"
)

var _ SearchEngine = (*TavilySearch)(nil)

type TavilySearch struct {
	MaxResults int `json:"max_results"`

	IncludeImages     bool `json:"include_images"`
	IncludeImageDesc  bool `json:"include_image_descriptions"`
	IncludeAnswer     bool `json:"include_answer"`
	IncludeRawContent bool `json:"include_raw_content"`
	Debug             bool

	Query string `json:"query"`

	ApiKey      string `json:"api_key"`
	Topic       string `json:"topic"`
	SearchDepth string `json:"search_depth"`
	Days        int    `json:"days"`

	IncludeDomains []string `json:"include_domains"`
	ExcludeDomains []string `json:"exclude_domains"`

	options task.WithOption `json:"-"`
}

type TavilySearchImage struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

type TavilySearchResult struct {
	Title         string  `json:"title"`
	URL           string  `json:"url"`
	Content       string  `json:"content"`
	Score         float64 `json:"score"`
	RawContent    *string `json:"raw_content"`
	PublishedDate *string `json:"published_date"`
}

type TavilySearchResponse struct {
	Query             string               `json:"query"`
	FollowUpQuestions *string              `json:"follow_up_questions"`
	Answer            *string              `json:"answer"`
	Images            []TavilySearchImage  `json:"images"`
	Results           []TavilySearchResult `json:"results"`
	ResponseTime      float64              `json:"response_time"`
}

func NewTavilySearch(options task.WithOption) *TavilySearch {
	ts := newTavilySearch(conf.Tavily.GetKey(), conf.Tavily.Debug, conf.Tavily.IncludeDomains, conf.Tavily.ExcludeDomains)
	ts.options = options
	return ts
}

// NewTavilySearch
func newTavilySearch(apiKey string, debug bool, includeDomain []string, excludeDomain []string) *TavilySearch {
	return &TavilySearch{
		MaxResults:        5,
		ApiKey:            apiKey,
		Topic:             TopicGeneral,
		Days:              DefaultDays,
		SearchDepth:       DepthBasic,
		IncludeImages:     false,
		IncludeImageDesc:  false,
		IncludeAnswer:     false,
		IncludeRawContent: true,
		Debug:             debug,
		IncludeDomains:    includeDomain,
		ExcludeDomains:    excludeDomain,
	}
}

// Search
func (t *TavilySearch) Search(ctx context.Context, query string) (document.DocumentCollection, error) {
	if err := t.applyParams(); err != nil {
		return nil, err
	}
	t.Query = query

	var body io.Reader
	reqbody, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("tavily params marshal error: %v", err)
	}
	body = strings.NewReader(string(reqbody))

	if t.Debug {
		fmt.Printf("use tavily api key: %s\n", t.ApiKey)
		fmt.Printf("Tavily api input: %s\n", string(reqbody))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TavilySearchEndpoint, body)
	if err != nil {
		return nil, fmt.Errorf("tavily api request error: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tavily API request error: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Tavily API response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tavily API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	if t.Debug {
		fmt.Printf("Tavily API output: %s\n", string(respBody))
	}
	var tsResponse TavilySearchResponse
	if err := json.Unmarshal(respBody, &tsResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Tavily API response: %v", err)
	}

	// 整理返回结果
	return t.formatResults(tsResponse), nil
}

// applyParams
// Available params:
// - debug: bool
// - limit: int
// - topic: string
// - search_depth: string
// - days: int
func (t *TavilySearch) applyParams() error {

	if err := param.Assign(&t.Debug, t.options.GetOptionWithDefault("debug", false)); err != nil {
		return err
	}
	if err := param.Assign(&t.MaxResults, t.options.GetOptionWithDefault("limit", 5)); err != nil {
		return err
	}

	if err := param.Assign(&t.Topic, t.options.GetOptionWithDefault("topic", TopicGeneral)); err != nil {
		return err
	}
	if t.Topic != TopicGeneral && t.Topic != TopicNews {
		return fmt.Errorf("tavily topic error: %s is not a valid topic", t.Topic)
	}

	if err := param.Assign(&t.SearchDepth, t.options.GetOptionWithDefault("search_depth", DepthBasic)); err != nil {
		return err
	}
	if t.SearchDepth != DepthBasic && t.SearchDepth != DepthAdvanced {
		return fmt.Errorf("tavily search depth error: %s is not a valid search depth", t.SearchDepth)
	}

	if err := param.Assign(&t.Days, t.options.GetOptionWithDefault("days", DefaultDays)); err != nil {
		return err
	}
	if t.Days < 1 || t.Days > 30 {
		return fmt.Errorf("tavily days error: %d is not a valid days, days must between 1 and 30", t.Days)
	}

	var includeDomains string
	includeDomainsUnsafe, _ := t.options.GetOption("include_domains")
	if includeDomainsUnsafe != nil {
		if err := param.Assign(&includeDomains, includeDomainsUnsafe); err != nil {
			return err
		}
		if includeDomains != "" {
			t.IncludeDomains = append(t.IncludeDomains, strings.Split(includeDomains, ",")...)
		}
	}

	var excludeDomains string
	excludeDomainsUnsafe, _ := t.options.GetOption("exclude_domains")
	if excludeDomainsUnsafe != nil {
		if err := param.Assign(&excludeDomains, excludeDomainsUnsafe); err != nil {
			return err
		}
		if excludeDomains != "" {
			t.ExcludeDomains = append(t.ExcludeDomains, strings.Split(excludeDomains, ",")...)
		}
	}

	return nil
}

// formatResults 格式化搜索结果
func (t *TavilySearch) formatResults(response TavilySearchResponse) document.DocumentCollection {
	documents := make([]document.Document, 0)
	layout := "Mon, 02 Jan 2006 15:04:05 MST"
	for _, result := range response.Results {
		content := result.Content
		if result.RawContent != nil {
			content = *result.RawContent
		}
		content = strings.TrimSpace(content)
		content = strings.Replace(content, "\n", " ", -1)

		doc := document.Document{
			Text: content,
			Name: result.Title,
			Source: &document.Resource{
				Name:     result.Title,
				Address:  result.URL,
				MineType: "text/html",
			},
			MetaData: make(map[string]string),
		}
		if result.PublishedDate != nil {
			publishedAt, err := time.Parse(layout, *result.PublishedDate)
			if err == nil {
				doc.MetaData["published_date"] = publishedAt.Format("2006年 01月 02日")
			}
		}
		documents = append(documents, doc)
	}

	return documents
}
