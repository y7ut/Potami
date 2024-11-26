package conf

import (
	"sync"

	fengchaogo "github.com/ijiwei/fengchao-go"
)

var fengchaoClient *fengchaogo.FengChao
var fengchaoOnce sync.Once

// GetFengchaoClient 获取fengchao客户端
func FengchaoClient() *fengchaogo.FengChao {
	fengchaoOnce.Do(func() {
		if fengchaoClient != nil {
			return
		}
		fengchaoClient = fengchaogo.NewFengChao(FengChao.APIKey, FengChao.APISecret, FengChao.BaseURL)
	})

	return fengchaoClient
}
