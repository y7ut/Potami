package document

import (
	"fmt"
	"strings"
)

var defaultCompressMethod = func(doc Document) string {
	return fmt.Sprintf("《%s》\n%s\n", doc.Name, doc.Text)
}

var defaultCompressConfig = &ComppressConfig{
	CompressMethod: defaultCompressMethod,
	Size:           1000,
	Depth:          true,
}

type ComppressConfig struct {
	CompressMethod func(Document) string
	Size           int
	Depth          bool
}

func WithCompressMethod(compressFunc func(Document) string) func(*ComppressConfig) {
	return func(config *ComppressConfig) {
		config.CompressMethod = compressFunc
	}
}

func WithSize(size int) func(*ComppressConfig) {
	return func(config *ComppressConfig) {
		config.Size = size
	}
}

func WithDepth(depth bool) func(*ComppressConfig) {
	return func(config *ComppressConfig) {
		config.Depth = depth
	}
}

func (d DocumentCollection) Compress(options ...func(*ComppressConfig)) string {
	config := defaultCompressConfig
	for _, option := range options {
		option(config)
	}
	return d.compress(config.CompressMethod, config.Size, config.Depth)
}

// Compress 格式化并压缩输出
// compressFunc 压缩函数, 返回压缩后的文本 size 压缩后的文本最大长度 depth 压缩时会深度优先
func (d DocumentCollection) compress(compressFunc func(Document) string, size int, depth bool) string {
	compressResult := make(map[string]int)
	length := 0
	for _, doc := range d {
		compressed := compressFunc(*doc)
		compressResult[compressed] = len(compressed)
		length += len(compressed)
	}

	var resultBuilder strings.Builder
	if depth {
		for doc := range compressResult {
			var breakdown bool
			if resultBuilder.Len()+len(doc) > size {
				doc = substringByRune(doc, 0, size-resultBuilder.Len())
				breakdown = true
			}
			doc = fmt.Sprintf("%s\n", doc)
			resultBuilder.WriteString(doc)
			if breakdown {
				break
			}
		}
	} else {
		for doc := range compressResult {
			doc = substringByRune(doc, 0, length/len(d))
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
