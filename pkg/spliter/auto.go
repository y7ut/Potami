package spliter

import (
	"context"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/y7ut/potami/internal/document"
)

type AutoSplitter struct {
	ChunkSize int // 基于UTF-8字符数的chunk大小
}

func (as *AutoSplitter) Split(ctx context.Context, resource *document.Resource) document.DocumentCollection {
	documents := make([]*document.Document, 0)
	for _, text := range smartSplit(resource.Content, as.ChunkSize) {
		documents = append(documents, &document.Document{
			Text:   text,
			Name:   resource.Name,
			Source: resource,
		})
	}

	return documents
}

// smartSplit 智能切分文本，基于UTF-8字符数，保持语义完整性
func smartSplit(text string, maxChunkSize int) []string {
	// 如果文本字符数小于等于最大chunk大小，直接返回
	if utf8.RuneCountInString(text) <= maxChunkSize {
		return []string{text}
	}

	var chunks []string
	lines := strings.Split(text, "\n")
	var currentChunk strings.Builder

	for _, line := range lines {
		lineRuneCount := utf8.RuneCountInString(line)
		
		// 如果单行字符数就超过最大长度，需要强制切分
		if lineRuneCount > maxChunkSize {
			// 先保存当前块（如果有内容）
			if currentChunk.Len() > 0 {
				chunkText := strings.TrimSpace(currentChunk.String())
				if chunkText != "" {
					chunks = append(chunks, chunkText)
				}
				currentChunk.Reset()
			}

			// 对超长行进行智能切分
			chunks = append(chunks, splitLongLine(line, maxChunkSize)...)
			continue
		}

		// 计算添加这一行后的总字符数
		currentContent := currentChunk.String()
		var testRuneCount int
		if currentContent != "" {
			testRuneCount = utf8.RuneCountInString(currentContent) + 1 + lineRuneCount // +1 for \n
		} else {
			testRuneCount = lineRuneCount
		}

		if testRuneCount > maxChunkSize {
			// 保存当前块
			if currentChunk.Len() > 0 {
				chunkText := strings.TrimSpace(currentChunk.String())
				if chunkText != "" {
					chunks = append(chunks, chunkText)
				}
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
		chunkText := strings.TrimSpace(currentChunk.String())
		if chunkText != "" {
			chunks = append(chunks, chunkText)
		}
	}

	return chunks
}

// splitLongLine 切分超长行，尝试在合适的位置切分
func splitLongLine(line string, maxChunkSize int) []string {
	var chunks []string
	runes := []rune(line)
	
	for len(runes) > 0 {
		if len(runes) <= maxChunkSize {
			// 剩余字符数不超过限制，直接添加
			chunks = append(chunks, string(runes))
			break
		}

		// 寻找最佳切分位置
		cutPos := findBestCutPosition(runes, maxChunkSize)
		
		// 提取chunk并移除已处理的部分
		chunk := string(runes[:cutPos])
		chunks = append(chunks, strings.TrimSpace(chunk))
		runes = runes[cutPos:]
		
		// 跳过开头的空白字符
		for len(runes) > 0 && unicode.IsSpace(runes[0]) {
			runes = runes[1:]
		}
	}
	
	return chunks
}

// findBestCutPosition 在maxChunkSize范围内寻找最佳切分位置
func findBestCutPosition(runes []rune, maxChunkSize int) int {
	if len(runes) <= maxChunkSize {
		return len(runes)
	}

	// 句子结束符，优先级从高到低
	sentenceEnders := []rune{'。', '！', '？', '.', '!', '?'}
	// 次要分隔符
	secondaryDelimiters := []rune{'；', ';', '，', ',', '：', ':'}
	// 空白字符
	whitespaces := []rune{' ', '\t', '\n', '\r'}

	// 搜索范围：从75%位置开始到maxChunkSize
	searchStart := maxChunkSize * 3 / 4
	if searchStart < 0 {
		searchStart = 0
	}

	// 1. 优先在句子结束符处切分
	for i := maxChunkSize - 1; i >= searchStart; i-- {
		for _, ender := range sentenceEnders {
			if runes[i] == ender {
				return i + 1
			}
		}
	}

	// 2. 在次要分隔符处切分
	for i := maxChunkSize - 1; i >= searchStart; i-- {
		for _, delimiter := range secondaryDelimiters {
			if runes[i] == delimiter {
				return i + 1
			}
		}
	}

	// 3. 在空白字符处切分
	for i := maxChunkSize - 1; i >= searchStart; i-- {
		for _, ws := range whitespaces {
			if runes[i] == ws {
				return i
			}
		}
	}

	// 4. 如果找不到合适的切分点，在maxChunkSize处强制切分
	return maxChunkSize
}
