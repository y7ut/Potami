package conf

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/viper"
)

const App = "potami"
const ConfigType = "yaml"

var (
	// Log 日志配置配置
	Log logConf
	// Redis redis 配置
	RedisConf redisConf
	// Task 任务队列配置
	Task taskConf
	// GoroutinePool 协程池配置
	GoroutinePool goroutinePoolConf
	// HttpServer http server 配置
	HttpServer httpServer
	// OPENAI openai 配置
	OpenAI openai
	// DB 数据库配置
	DB db
	// Tavily tavily 配置
	Tavily tavily
	// GoogleCustomSearch google custom search 配置
	GoogleCustomSearch googleCustomSearch
)

func InitConfig() {

	// Find current directory
	pwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	viper := viper.New()
	viper.SetConfigName(App)        // name of config file (without extension)
	viper.SetConfigType(ConfigType) // REQUIRED if the config file does not have the extension in the name

	viper.AddConfigPath(path.Join(pwd, "config")) // optionally look for config in the project directory or config directory

	// Read config
	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {            // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	configs := map[string]any{
		"log":                  &Log,
		"task":                 &Task,
		"pool":                 &GoroutinePool,
		"http":                 &HttpServer,
		"openai":               &OpenAI,
		"db":                   &DB,
		"redis":                &RedisConf,
		"tavily":               &Tavily,
		"google_custom_search": &GoogleCustomSearch,
	}

	for k, v := range configs {
		err := viper.UnmarshalKey(k, v)
		if err != nil {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}
}
