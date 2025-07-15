package document

import (
	"encoding/json"
	"fmt"
)

type Document struct {
	Score float64
	ID    string
	Name  string
	Text  string

	Context      string
	ContextEmbed []float64

	Embed []float64

	MetaData map[string]string
	Source   *Resource
}

type DocumentCollection []Document

func (d Document) String() string {
	// 格式化输出
	metadata, _ := json.MarshalIndent(d.MetaData, "", "  ")
	return fmt.Sprintf("ID: %s \nName: %s \nText: %s \nScore: %f \nMetaData: %v\n", d.ID, d.Name, d.Text, d.Score, string(metadata))
}
