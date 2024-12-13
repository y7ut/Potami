package boardcast

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sirupsen/logrus"
	"gopkg.in/antage/eventsource.v1"
)

const (
	Heartbeat      = 10 // 实际心跳包发送间隔时间
	HeartbeatEvent = "heartbeat"
	FinishEvent    = "finished"
	StartEvent     = "start"
	UpdateEvent    = "update"
	DeadEvent      = "dead"
	WaitEvent      = "wait"
)

// Boardcast SSE广播
type Boardcast[T any] struct {
	RoundID         int  // RoundID 通知广播进行的轮次ID
	IsStop          bool // IsStop 广播是否结束
	Resource        T    // Resource 广播通知的资源
	LastBoardcastAt time.Time

	selfCheckFunc func(resource T) string // 资源自检的方法
	stopCh        chan struct{}           // stopCh

	eventSource eventsource.EventSource // SSE
	once        sync.Once               // once
}

// NewBoardCast 初始化一个资源广播
func NewBoardCast[T any](resource T, selfCheckFunc func(resource T) string) *Boardcast[T] {
	b := &Boardcast[T]{
		Resource: resource,
		eventSource: eventsource.New(
			&eventsource.Settings{
				Timeout:        2 * time.Second,
				CloseOnTimeout: true,
				IdleTimeout:    time.Duration(2*Heartbeat) * time.Second,
				Gzip:           true,
			},
			func(req *http.Request) [][]byte {
				return [][]byte{
					[]byte("X-Accel-Buffering: no"),
					[]byte("Access-Control-Allow-Origin: *"),
					[]byte("Cache-Control: no-cache"),
				}
			},
		),
		stopCh:        make(chan struct{}),
		once:          sync.Once{},
		selfCheckFunc: selfCheckFunc,
	}
	b.Start()
	return b
}

// Start 开启广播
func (t *Boardcast[T]) Start() {
	t.once.Do(func() {
		go func() {
			// 心跳
			logrus.Info("boardcast heartbeat start")
			heartbeatTicker := time.NewTicker(time.Second)
			defer func() {
				logrus.Debug("boardcast heartbeat stop")
				heartbeatTicker.Stop()
				t.eventSource.Close()
			}()
			for {
				select {
				case <-t.stopCh:
					return
				case <-heartbeatTicker.C:
					if time.Since(t.LastBoardcastAt) < time.Duration(Heartbeat)*time.Second {
						continue
					}
					event := t.selfCheckFunc(t.Resource)
					t.Send(event)
				}
			}
		}()
	})
}

// Listen 监听广播
func (t *Boardcast[T]) Listen(w gin.ResponseWriter, r *http.Request) error {
	if t.IsStop {
		return fmt.Errorf("boardcast has been closed")
	}
	// 把当前连接放到到es对象的链接列表
	if t.eventSource == nil {
		return fmt.Errorf("eventsource has been closed")
	}
	t.eventSource.ServeHTTP(w, r)
	logrus.Debug("new client connected!")
	return nil
}

// 发送消息到广播中
func (t *Boardcast[T]) Send(event string) {
	if t.IsStop {
		return
	}
	t.LastBoardcastAt = time.Now()
	t.RoundID++
	stream, err := json.Marshal(t.Resource)
	if err != nil {
		logrus.Errorf("boardcast error: %s", err)
	}
	logrus.Debugf("boardcast event: %s", event)
	t.eventSource.SendEventMessage(string(stream), event, strconv.Itoa(t.RoundID))
}

// Stop 停止推送消息到浏览器，关闭广播
func (t *Boardcast[T]) Stop() {
	if t.IsStop {
		return
	}
	t.IsStop = true
	close(t.stopCh)
}
