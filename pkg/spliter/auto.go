package spliter

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/y7ut/potami/internal/document"
)

type AutoSplitter struct {
	ChunkSize int
}

func (as *AutoSplitter) Split(ctx context.Context, resource *document.Resource) document.DocumentCollection {
	documents := make([]document.Document, 0)
	for _, text := range smartSplit(resource.Content, as.ChunkSize) {
		documents = append(documents, document.Document{
			Text: text,
			Name: resource.Name,
			Source: &document.Resource{
				Name:     resource.Name,
				Address:  resource.Address,
				MineType: resource.MineType,
			},
		})
	}

	return documents
}

// smartSplit 智能切分文本，保持语义完整性和UTF-8字符完整性
func smartSplit(text string, maxChunkSize int) []string {
	if len(text) <= maxChunkSize {
		return []string{text}
	}

	var chunks []string
	lines := strings.Split(text, "\n")

	var currentChunk strings.Builder

	for _, line := range lines {
		// 如果单行就超过最大长度，需要强制切分
		if len(line) > maxChunkSize {
			// 先保存当前块（如果有内容）
			if currentChunk.Len() > 0 {
				chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
				currentChunk.Reset()
			}

			// 对超长行进行安全切分
			for len(line) > 0 {
				chunkSize := maxChunkSize
				if len(line) < chunkSize {
					chunkSize = len(line)
				}

				// 尝试在句号、感叹号、问号等处切分
				cutPos := chunkSize
				if len(line) > chunkSize {
					// 在合理范围内寻找句子结束符
					searchStart := chunkSize * 3 / 4 // 从75%位置开始搜索
					runes := []rune(line)
					if searchStart < len(runes) && chunkSize < len(runes) {
						for i := searchStart; i < chunkSize && i < len(runes); i++ {
							if runes[i] == '。' || runes[i] == '！' || runes[i] == '？' || runes[i] == '\n' {
								cutPos = utf8.RuneCountInString(string(runes[:i+1]))
								break
							}
						}
					}
				}

				chunk := safeSubstring(line, cutPos)
				if chunk != "" {
					chunks = append(chunks, strings.TrimSpace(chunk))
				}
				line = line[len(chunk):]
			}
			continue
		}

		// 检查添加这一行是否会超过限制
		testContent := currentChunk.String()
		if testContent != "" {
			testContent += "\n"
		}
		testContent += line

		if len(testContent) > maxChunkSize {
			// 保存当前块
			if currentChunk.Len() > 0 {
				chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
				currentChunk.Reset()
			}
			// 开始新块
			currentChunk.WriteString(line)
		} else {
			// 添加到当前块
			if currentChunk.Len() > 0 {
				currentChunk.WriteString("\n")
			}
			currentChunk.WriteString(line)
		}
	}

	// 保存最后一个块
	if currentChunk.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
	}

	return chunks
}

// safeSubstring 安全地截取字符串，确保不会在UTF-8字符中间切断
func safeSubstring(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// 从maxLen位置向前查找，找到一个有效的UTF-8字符边界
	for i := maxLen; i > 0; i-- {
		if utf8.ValidString(s[:i]) {
			return s[:i]
		}
	}

	// 如果找不到有效边界，返回空字符串（这种情况很少见）
	return ""
}
