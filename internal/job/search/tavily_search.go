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
func (t *TavilySearch) Search(ctx context.Context, query string, options ...SearchEngineOption) ([]Document, error) {
	if err := t.applyOptions(options...); err != nil {
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

func (t *TavilySearch) applyOptions(options ...SearchEngineOption) error {
	om := DefaultOptionsManager()
	for _, applyOption := range options {
		applyOption(om)
	}

	if om.Debug {
		t.Debug = om.Debug
	}

	t.MaxResults = om.Limit

	t.Topic = om.Get("topic", TopicGeneral).(string)

	if t.Topic != TopicGeneral && t.Topic != TopicNews {
		return fmt.Errorf("tavily topic error: %s", t.Topic)
	}

	t.SearchDepth = om.Get("search_depth", DepthBasic).(string)
	if t.SearchDepth != DepthBasic && t.SearchDepth != DepthAdvanced {
		return fmt.Errorf("tavily search depth error: %s", t.SearchDepth)
	}

	t.Days = om.Get("days", 7).(int)
	if t.Days < 1 || t.Days > 30 {
		return fmt.Errorf("tavily days error: %d", t.Days)
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

// DocumentsOutputParse 格式化输出
func DocumentsOutputParse(documents []Document, size int, depthMode bool) string {
	ducumentDict := make(map[string]int)
	length := 0
	for _, doc := range documents {
		content := fmt.Sprintf("%s\n%s\n", doc.Title, doc.Content)
		if !doc.PublishedAt.IsZero() {
			content = fmt.Sprintf("%s\n%s %s\n", doc.Title, doc.PublishedAt.Format("2006年 01月 02日："), doc.Content)
		}
		ducumentDict[content] = len(content)
		length += len(content)
	}

	var resultBuilder strings.Builder
	if depthMode {
		for doc := range ducumentDict {
			var breakdown bool
			if resultBuilder.Len() > size {
				doc = SubstringByRune(doc, 0, resultBuilder.Len()-size)
				breakdown = true
			}
			doc = fmt.Sprintf("%s\n", doc)
			resultBuilder.WriteString(doc)
			if breakdown {
				resultBuilder.WriteString("...")
				break
			}
		}
	} else {
		for doc := range ducumentDict {
			doc = SubstringByRune(doc, 0, size/len(documents))
			doc = fmt.Sprintf("%s\n", doc)
			resultBuilder.WriteString(doc)
		}
	}

	return resultBuilder.String()
}

// SubstringByRune 按字符获取字符串的部分段落
func SubstringByRune(s string, start, length int) string {
	runes := []rune(s) // 将字符串转换为字符切片
	if start < 0 || start >= len(runes) {
		return "" // 起始位置无效，返回空字符串
	}

	end := start + length
	if end > len(runes) {
		end = len(runes) // 如果超出范围，取最大长度
	}

	return string(runes[start:end]) // 截取并转换回字符串
}
