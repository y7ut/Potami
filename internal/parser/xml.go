package parser

import (
	"fmt"
	"regexp"
	"strings"
)

const outputParseRegxTemplate = "(?s)<%s>(.*?)</%s>"

// XMLOutPutParse 解析XML输出, parses 为空则解析全部, 默认属性名为output
func XMLOutPutParser(content string, parses ...string) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	if len(parses) == 0 {
		values["output"] = content
		return values, nil
	}
	for _, p := range parses {
		expr := fmt.Sprintf(outputParseRegxTemplate, p, p)
		re, err := regexp.Compile(expr)
		if err != nil {
			return nil, fmt.Errorf("output parse error: %v", err)
		}
		match := re.FindStringSubmatch(content)
		if len(match) < 1 {
			continue
		}
		// 将匹配到的内容替换掉 {{标签}} 的格式
		values[p] = strings.Trim(match[1], "\n")
	}
	return values, nil
}
