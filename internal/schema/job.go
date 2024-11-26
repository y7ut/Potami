package schema

type Job struct {
	Name        string   `mapstructure:"name" json:"name" yaml:"name"`                                                   // job 的名称
	Type        string   `mapstructure:"type" json:"type" yaml:"type" validate:"required,oneof=prompt api_tool search" ` // job 的类型
	Description string   `mapstructure:"description" json:"description" yaml:"description"`                              // 通用 job 的描述
	Params      []string `mapstructure:"params" json:"params" yaml:"params"`                                             // 通用 job 中的参数

	LlmModel     string   `mapstructure:"llm_model,omitempty" json:"llm_model,omitempty" yaml:"llm_model,omitempty"`             // 仅在 prompt 类型的 job 中使用
	Temperature  float64  `mapstructure:"temperature,omitempty" json:"temperature,omitempty" yaml:"temperature,omitempty"`       // 仅在 prompt 类型的 job 中使用
	TopP         float64  `mapstructure:"top_p,omitempty" json:"top_p,omitempty" yaml:"top_p,omitempty"`                         // 仅在 prompt 类型的 job 中使用
	MaxTokens    int      `mapstructure:"max_tokens,omitempty" json:"max_tokens,omitempty" yaml:"max_tokens,omitempty"`          // 仅在 prompt 类型的 job 中使用
	SystemPrompt string   `mapstructure:"system_prompt,omitempty" json:"system_prompt,omitempty" yaml:"system_prompt,omitempty"` // 仅在 prompt 类型的 job 中使用
	Template     string   `mapstructure:"template,omitempty" json:"template,omitempty" yaml:"template,omitempty"`                // 仅在 prompt 类型的 job 中使用
	Output       []string `mapstructure:"output,omitempty" json:"output,omitempty" yaml:"output,omitempty"`                      // prompt 类型 job 的输出

	SearchEngine  string                 `mapstructure:"search_engine,omitempty" json:"search_engine,omitempty" yaml:"search_engine,omitempty"`    // 仅在 search 类型的 job 中使用
	SearchLimit   int                    `mapstructure:"search_limit,omitempty" json:"search_limit,omitempty" yaml:"search_limit,omitempty"`       // 仅在 search 类型的 job 中使用
	BlockSize     int                    `mapstructure:"block_size,omitempty" json:"block_size,omitempty" yaml:"block_size,omitempty"`             // 仅在 search 类型的 job 中使用
	SearchOptions map[string]interface{} `mapstructure:"search_options,omitempty" json:"search_options,omitempty" yaml:"search_options,omitempty"` // 仅在 search 类型的 job 中使用
	QueryField    string                 `mapstructure:"query_field,omitempty" json:"query_field,omitempty" yaml:"query_field,omitempty"`          // 仅在 search 类型的 job 中使用
	OutputField   string                 `mapstructure:"output_field,omitempty" json:"output_field,omitempty" yaml:"output_field,omitempty"`       // 仅在 search 类型的 job 中使用

	Endpoint string `mapstructure:"endpoint,omitempty" json:"endpoint,omitempty" yaml:"endpoint,omitempty"` // 仅在 api_tool 类型的 job 中使用
	Method   string `mapstructure:"method,omitempty" json:"method,omitempty" yaml:"method,omitempty"`       // 仅在 api_tool 类型的 job 中使用

	OutputParses map[string]string `mapstructure:"output_parses,omitempty" json:"output_parses,omitempty" yaml:"output_parses,omitempty"` // 仅在 api_tool 类型 job 中使用
}
