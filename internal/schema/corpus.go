package schema

// Corpus 知识库
type Corpus struct {
	Name               string `mapstructure:"name" json:"name" yaml:"name"`
	Description        string `mapstructure:"description" json:"description" yaml:"description"`
	CollectionName     string `mapstructure:"collection_name" json:"collection_name" yaml:"collection_name"`
	RerankFusionMethod string `mapstructure:"rerank_fusion_method" json:"rerank_fusion_method" yaml:"rerank_fusion_method"`

	EmbeddingProvider string `mapstructure:"embedding_provider" json:"embedding_provider" yaml:"embedding_provider"`
	EmbeddingModel    string `mapstructure:"embedding_model" json:"embedding_model" yaml:"embedding_model"`

	EmbeddingOptions map[string]interface{} `mapstructure:"embedding_options" json:"embedding_options" yaml:"embedding_options"`

	VectorStore      string `mapstructure:"vector_store" json:"vector_store" yaml:"vector_store"`
	VectorDimension  int64  `mapstructure:"vector_dimension" json:"vector_dimension" yaml:"vector_dimension"`
	VectorMetricType string `mapstructure:"vector_metric_type" json:"vector_metric_type" yaml:"vector_metric_type"`

	UseBm25Index    bool `mapstructure:"use_bm25_index" json:"use_bm25_index" yaml:"use_bm25_index"`          // 是否使用BM25权重索引
	UseContextEmbed bool `mapstructure:"use_context_embed" json:"use_context_embed" yaml:"use_context_embed"` // 是否使用上下文嵌入

	// use_context_embed 为 true 时使用的上下文嵌入 prompt, 如果为空则使用默认的 prompt
	// prompt 中可以使用 {{document}} 表示整个文档，{{chunk}} 表示当前 chunk, 默认会开启 prompt cache
	ContextEmbedPrompt string `mapstructure:"context_embed_prompt" json:"context_embed_prompt" yaml:"context_embed_prompt"`
}
