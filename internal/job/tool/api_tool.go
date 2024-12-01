package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/y7ut/potami/internal/parser"
	"github.com/y7ut/potami/internal/task"
)

type APITool struct {
	task.JobHelper

	Endpoint   string
	HttpMethod string

	Params       []string
	OutputParses map[string]string
}

func (t *APITool) Handle(ctx context.Context) error {
	result, err := t.call(ctx)
	if err != nil {
		t.Logger().WithFields(t.GetAttributes()).WithError(err).Error("tool[api] error")
		return err
	}

	attributes, err := parser.JsonPathOutputParser([]byte(result), t.OutputParses)
	if err != nil {
		t.Logger().WithFields(t.GetAttributes()).WithError(err).Error("tool[api] output parse error")
		return err
	}
	t.SetAttributes(attributes)

	t.Logger().WithFields(t.GetAttributes()).Debug("tool[api] complete")
	return nil
}

func (t *APITool) call(ctx context.Context) ([]byte, error) {
	if t.HttpMethod == "" {
		t.HttpMethod = http.MethodGet
	}

	EndpointUrl := t.Endpoint

	generateRequestsFuncSet := map[string]func() (*http.Request, error){
		http.MethodGet: func() (*http.Request, error) {
			for _, param := range t.Params {
				v, ok := t.GetAttribute(param)
				if !ok {
					continue
				}
				if strings.Contains(EndpointUrl, "?") {
					EndpointUrl += "&" + param + "=" + fmt.Sprintf("%v", v)
				} else {
					EndpointUrl += "?" + param + "=" + fmt.Sprintf("%v", v)
				}
			}
			return http.NewRequestWithContext(ctx, http.MethodGet, EndpointUrl, nil)
		},
		http.MethodPost: func() (*http.Request, error) {
			var body io.Reader
			reqbody, err := json.Marshal(t.GetAttributes(t.Params...))
			if err != nil {
				return nil, fmt.Errorf("tool api request error: %v", err)
			}
			body = strings.NewReader(string(reqbody))
			return http.NewRequestWithContext(ctx, http.MethodPost, EndpointUrl, body)
		},
	}
	generateRequest, ok := generateRequestsFuncSet[t.HttpMethod]
	if !ok {
		return nil, fmt.Errorf("tool api request error: %s not support", t.HttpMethod)
	}
	req, err := generateRequest()
	if err != nil {
		return nil, fmt.Errorf("tool api request error: %v", err)
	}

	req.Header.Set("User-Agent", "potami")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tool api request error: %v", err)
	}

	if resp != nil {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("tool api error: %v", err)
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("tool api error: %s", string(body))
		}
		return body, nil
	}

	return nil, fmt.Errorf("tool api error: resp is nil")
}
