package search

import (
	"fmt"
	"strings"
	"time"
)

// Document 文档
type Document struct {
	Content     string    `json:"content"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"published_at"`
}

// DocumentsOutputParse 格式化并压缩输出
func DocumentsCompress(documents []Document, size int, depthMode bool) string {
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
				doc = substringByRune(doc, 0, resultBuilder.Len()-size)
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
			doc = substringByRune(doc, 0, size/len(documents))
			doc = fmt.Sprintf("%s\n", doc)
			resultBuilder.WriteString(doc)
		}
	}

	return resultBuilder.String()
}

// substringByRune 按字符获取字符串的部分段落
func substringByRune(s string, start, length int) string {
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
