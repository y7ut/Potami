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
)

const (
	TopicGeneral         = "general"
	TopicNews            = "news"
	DepthBasic           = "basic"
	DepthAdvanced        = "advanced"
	TavilySearchEndpoint = "https://api.tavily.com/search"
)

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

func NewTavilySearch() *TavilySearch {
	return newTavilySearch(conf.Tavily.GetKey(), conf.Tavily.Days, conf.Tavily.Debug, conf.Tavily.IncludeDomains, conf.Tavily.ExcludeDomains)
}

// NewTavilySearch
func newTavilySearch(apiKey string, days int, debug bool, includeDomain []string, excludeDomain []string) *TavilySearch {
	return &TavilySearch{
		MaxResults:        5,
		ApiKey:            apiKey,
		Topic:             TopicGeneral,
		Days:              days,
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
func (t *TavilySearch) Search(ctx context.Context, param *SearchParam) ([]Document, error) {
	if err := t.applyParams(param); err != nil {
		return nil, err
	}

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
func (t *TavilySearch) applyParams(param *SearchParam) error {

	if param.Debug {
		t.Debug = param.Debug
	}

	t.MaxResults = param.Limit
	t.Query = param.Query

	t.Topic = param.Topic
	if t.Topic != TopicGeneral && t.Topic != TopicNews {
		return fmt.Errorf("tavily topic error: %s is not a valid topic", t.Topic)
	}

	t.SearchDepth = param.SearchDepth
	if t.SearchDepth != DepthBasic && t.SearchDepth != DepthAdvanced {
		return fmt.Errorf("tavily search depth error: %s is not a valid search depth", t.SearchDepth)
	}

	t.Days = param.Days
	if t.Days < 1 || t.Days > 30 {
		return fmt.Errorf("tavily days error: %d is not a valid days, days must between 1 and 30", t.Days)
	}

	return nil
}

// formatResults 格式化搜索结果
func (t *TavilySearch) formatResults(response TavilySearchResponse) []Document {
	documents := make([]Document, 0)
	layout := "Mon, 02 Jan 2006 15:04:05 MST"
	for _, result := range response.Results {
		content := result.Content
		if result.RawContent != nil {
			content = *result.RawContent
		}
		content = strings.TrimSpace(content)
		content = strings.Replace(content, "\n", " ", -1)

		doc := Document{
			Content:     content,
			Title:       result.Title,
			URL:         result.URL,
			PublishedAt: time.Time{},
		}
		if result.PublishedDate != nil {
			publishedAt, err := time.Parse(layout, *result.PublishedDate)
			if err == nil {
				doc.PublishedAt = publishedAt
			}
		}
		documents = append(documents, doc)
	}

	return documents
}
