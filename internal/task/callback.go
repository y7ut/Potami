package task

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/y7ut/potami/pkg/json"
)

const (
	CallBackHttpRequestHeaderAgent = "potami"
	CallBackHttpRequestMethod      = http.MethodGet
	CallBackHttpRequestContentType = "application/json"
	CallBackHttpRequestTimeout     = 5 * time.Second
)

type CallBackConfig struct {
	Url  string `json:"url"`
	Args string `json:"args"`

	MetaDataBuilder func(matedata map[string]interface{}, args []string) map[string]interface{}
}

type CallBackParams struct {
	TaskID          string   `json:"task_id" binding:"required"`
	Call            string   `json:"call" binding:"required"`
	HealthPercent   *float64 `json:"health" binding:"required"`
	CompletePercent *float64 `json:"complete" binding:"required"`

	CreatedAt  time.Time `json:"created_at"`  // 创建时间
	StartAt    time.Time `json:"start_at"`    // 开启时间
	CompleteAt time.Time `json:"complete_at"` // 完成时间
	CloseAt    time.Time `json:"close_at"`    // 关闭时间

	Errors []error `json:"errors,omitempty"`

	Data map[string]interface{} `json:"data"`
}

// sendCallBackRequest
func sendCallBackRequest(stream *Task, callbackParamsBuilder func(matedata map[string]interface{}, args []string) map[string]interface{}) error {
	callbackConfig := stream.CallBack
	if callbackConfig == nil {
		return nil
	}

	var callbackData map[string]interface{}
	if callbackParamsBuilder != nil {
		callbackData = callbackParamsBuilder(stream.MetaData, strings.Split(callbackConfig.Args, ","))
	}

	hp := stream.Health()
	cp := stream.GetCompleteness()
	callbackRequest := &CallBackParams{
		TaskID:          stream.ID,
		Call:            stream.Call,
		HealthPercent:   &hp,
		CompletePercent: &cp,
		CreatedAt:       stream.CreatedAt,
		StartAt:         stream.StartAt,
		CompleteAt:      stream.CompleteAt,
		CloseAt:         stream.CloseAt,
		Errors:          stream.ErrorStacks,
		Data:            callbackData,
	}

	// request
	var req *http.Request
	var err error
	reqbody, err := json.Marshal(callbackRequest)
	if err != nil {
		return fmt.Errorf("call back request error: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if req, err = http.NewRequestWithContext(ctx, CallBackHttpRequestMethod, callbackConfig.Url, strings.NewReader(string(reqbody))); err != nil {
		return fmt.Errorf("call back request error: %v", err)
	}
	req.Header.Set("User-Agent", CallBackHttpRequestHeaderAgent)
	req.Header.Set("Content-Type", CallBackHttpRequestContentType)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(reqbody)))

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return fmt.Errorf("call back request error: %v", err)
	}

	if resp != nil {
		if resp.StatusCode != 200 {
			defer resp.Body.Close()
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read callback response body: %v", err)
			}
			return fmt.Errorf("call back error: %s", string(bodyBytes))
		}
	}
	return nil
}
