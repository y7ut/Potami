package conf

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// logConf 日志配置
type logConf struct {
	Path      string `mapstructure:"path"`
	Name      string `mapstructure:"name"`
	Level     string `mapstructure:"level"`
	ZineURL   string `mapstructure:"zinc_url"`
	ZineIndex string `mapstructure:"zinc_index"`
}

type redisConf struct {
	Address          []string      `mapstructure:"address"`
	Username         string        `mapstructure:"username"`
	Password         string        `mapstructure:"password"`
	Select           int           `mapstructure:"select"`
	DisableCache     bool          `mapstructure:"disable_cache"`
	ConnWriteTimeout time.Duration `mapstructure:"conn_write_timeout"`
}

// task 任务
type taskConf struct {
	Tracking   bool           `mapstructure:"tracking"`
	RetryCount int            `mapstructure:"retry_count"` // 重试次数
	LevelMap   map[string]int `mapstructure:"level_map"`   // 优先级配置
}

// goroutinePool 协程池
type goroutinePoolConf struct {
	MaxWorkers      int `mapstructure:"max_worker"`        // 最大协程数（并非数量）
	Timeout         int `mapstructure:"time_out"`          // 协程池中所执行的单个任务的 Timeout
	MaxIdleDuration int `mapstructure:"max_idle_duration"` // 协程释放资源的空闲时间
}

type httpServer struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type openai struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
}

type fengchao struct {
	APIKey    string `mapstructure:"api_key"`
	APISecret string `mapstructure:"api_secret"`
	BaseURL   string `mapstructure:"base_url"`
}

type db struct {
	Type string `mapstructure:"type"`
	Path string `mapstructure:"path"`
}

type tavily struct {
	Days    int      `mapstructure:"days"`
	APIKey  string   `mapstructure:"api_key"`
	APIKeys []string `mapstructure:"api_keys"`
	Debug   bool     `mapstructure:"debug"`

	IncludeDomains []string `mapstructure:"include_domains"`
	ExcludeDomains []string `mapstructure:"exclude_domains"`
}

func (t *tavily) GetKey() string {
	if t.APIKey != "" || len(t.APIKeys) == 0 {
		return t.APIKey
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return t.APIKeys[r.Intn(len(t.APIKeys))]
}

type googleCustomSearch struct {
	APIKey string `mapstructure:"api_key"`
	CX     string `mapstructure:"cx"`
}

func (hs *httpServer) Address() string {
	return hs.Host + ":" + strconv.Itoa(hs.Port)
}

// TransformLevel 转换日志级别
func (conf *logConf) TransformLevel() logrus.Level {
	switch conf.Level {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.DebugLevel
	}
}
