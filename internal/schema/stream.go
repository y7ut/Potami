package schema

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

type Stream struct {
	Name        string `mapstructure:"name" json:"name" yaml:"name"`
	Description string `mapstructure:"description" json:"description" yaml:"description"`
	Jobs        []*Job `mapstructure:"jobs" json:"jobs" yaml:"jobs"`
	Level       int    `mapstructure:"level,omitempty" json:"level,omitempty" yaml:"level,omitempty"`
}

type HumanFriendlyStreamConfig struct {
	Name           string              `json:"name,omitempty"`
	Description    string              `json:"description,omitempty"`
	Jobs           []map[string]string `json:"jobs,omitempty"`
	RequiredParams []string            `json:"required_params,omitempty"`
	Output         []string            `json:"output,omitempty"`
}

// 自定义验证器，用于验证 Stream 和 Job 的特定规则
var validate = validator.New()

func ValidateStream(stream Stream) error {
	jobNames := make(map[string]bool)
	paramsFlow := make(map[string]bool)

	for _, job := range stream.Jobs {
		// 检查 Job 名称唯一性
		if jobNames[job.Name] {
			return fmt.Errorf("job 名称 %s 重复", job.Name)
		}
		jobNames[job.Name] = true

		// 确保 prompt 类型 job 至少有一个 output
		if job.Type == "prompt" && len(job.Output) == 0 {
			return fmt.Errorf("job %s 类型为 prompt 时必须包含至少一个 output", job.Name)
		}

		if job.Type == "api_tool" {
			for parses := range job.OutputParses {
				paramsFlow[parses] = true
			}
		}

		// 检查 output 参数在 template 中是否有对应的 XML 标签
		if err := checkOutputXMLTags(job, paramsFlow); err != nil {
			return err
		}

		// 检查 params 是否符合流转要求
		if err := checkParamsFlow(job, paramsFlow); err != nil {
			return err
		}

		// 确保 api_tool 类型 job 才能包含 output_parses
		if job.Type != "api_tool" && len(job.OutputParses) > 0 {
			return fmt.Errorf("job %s 类型为 %s 时不应包含 output_parses", job.Name, job.Type)
		}

		if job.Type == "search" {
			if _, ok := paramsFlow[job.QueryField]; !ok {
				return fmt.Errorf("search %s 的 query_field %s 不在 params 中", job.Name, job.QueryField)
			}
			paramsFlow[job.OutputField] = true
		}

	}

	// 验证其他结构体字段是否符合基本约束
	return validate.Struct(stream)
}

// 检查 Output 是否存在于 Template 的 XML 标签中
func checkOutputXMLTags(job *Job, paramsFlow map[string]bool) error {
	for _, output := range job.Output {

		tagPattern := fmt.Sprintf("(?s)<%s>(.*?)</%s>", output, output)
		matched, err := regexp.MatchString(tagPattern, job.Template+job.SystemPrompt)
		if err != nil {
			return fmt.Errorf("模板匹配出错: %v", err)
		}
		if !matched {
			return fmt.Errorf("job %s 的 output %s 在 template 中缺少对应的 XML 标签", job.Name, output)
		}

		paramsFlow[output] = true
	}
	return nil
}

// 检查 Params 流转是否合法
func checkParamsFlow(job *Job, paramsFlow map[string]bool) error {
	// 正则匹配 template 中的所有占位符
	re := regexp.MustCompile(`{{\.(\w+)}}`)
	matches := re.FindAllStringSubmatch(job.Template+job.SystemPrompt, -1)
	templateParams := make(map[string]bool)

	for _, match := range matches {
		templateParams[match[1]] = true
	}

	// 检查 job.Params 中的参数是否在上一个 job 的 params 或 output 中
	for _, param := range job.Params {
		if !templateParams[param] && job.Type == "prompt" {
			return fmt.Errorf("job %s 的参数 %s 在 template 中未出现", job.Name, param)
		}
		paramsFlow[param] = true
	}

	return nil
}
