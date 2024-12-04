package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/y7ut/potami/api"
	"github.com/y7ut/potami/internal/boardcast"
	"github.com/y7ut/potami/pkg/json"
)

// event bytes prefix 用于判断事件数据行的类型
var (
	// StartEventPrefix  start event 生成启动的事件的字节前缀
	StartEventPrefix = []byte("event: start")
	// AddEventPrefix  add event 生成数据的事件的字节前缀
	AddEventPrefix = []byte("event: update")
	// FinishEventPrefix  finish event 生成结束的事件的字节前缀
	FinishEventPrefix = []byte("event: finished")
	// DeadEventPrefix  error event 生成中断的事件的字节前缀
	DeadEventPrefix = []byte("event: dead")
	// ErrorHeartPrefix  error event 生成错误的事件的字节前缀
	ErrorHeartPrefix = []byte("event: heartbeat")
	// DataPrefix  data event 生成数据的事件的字节前缀
	DataPrefix = []byte("data: ")
	// IDPrefix  data event 生成数据的事件的字节前缀
	IDPrefix = []byte("id: ")

	completeParams = make(map[string]string)
)

// switchEventType 用于判断事件数据行的类型
func switchEventType(line []byte) string {
	if bytes.HasPrefix(line, StartEventPrefix) {
		return boardcast.StartEvent
	} else if bytes.HasPrefix(line, AddEventPrefix) {
		return boardcast.UpdateEvent
	} else if bytes.HasPrefix(line, FinishEventPrefix) {
		return boardcast.FinishEvent
	} else if bytes.HasPrefix(line, DeadEventPrefix) {
		return boardcast.DeadEvent
	} else if bytes.HasPrefix(line, ErrorHeartPrefix) {
		return boardcast.HeartbeatEvent
	}
	return ""
}

var CompleteCommand = &cobra.Command{
	Use:    "complete",
	Short:  "complete task with stream",
	PreRun: InitContextAndClient,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 1 {
			log.Fatalf("required stream name to complete")
		}
		streamsName := args[0]

		var params = make(map[string]interface{})
		if completeParams != nil {
			for k, v := range completeParams {
				params[k] = v
			}
		} else if len(args) < 2 {
			fi, err := os.Stdin.Stat()
			if err != nil {
				log.Fatalf("failed to read params from stdin: %v", err)
			}

			if fi.Mode()&os.ModeNamedPipe == 0 {
				log.Fatalf("stdin is not a pipe")
			}
			bytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				log.Fatalf("failed to read params from stdin: %v", err)
			}
			err = json.Unmarshal(bytes, &params)
			if err != nil {
				log.Fatalf("failed to parse params: %v", err)
			}
		} else {
			paramsBytes := []byte(args[1])
			err := json.Unmarshal(paramsBytes, &params)
			if err != nil {
				log.Fatalf("failed to parse params: %v", err)
			}
		}

		CompletionTask := &api.StreamTask{
			Params: params,
			Name:   streamsName + time.Now().Format("_2006-01-02 15:04:05"),
			Mode:   "stream",
		}

		resp, err := client.SetDisableWarn(true).R().
			SetHeader("Content-Type", "application/json").
			SetBody(CompletionTask).
			SetDoNotParseResponse(true).
			Post("/api/complete/" + streamsName)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode() != 200 {
			buffer := bufio.NewReader(resp.RawResponse.Body)
			defer resp.RawResponse.Body.Close()
			errorStr, err := buffer.ReadString('\n')
			if err != nil && err != io.EOF {
				log.Fatalf("failed to complete stream: %s", err)
			}
			log.Fatalf("failed to complete stream: %s", errorStr)
		}
		completionTaskOutputParses(resp.RawResponse)
	},
}

func init() {
	CompleteCommand.Flags().StringP("context", "c", "", "the context used")
	CompleteCommand.Flags().StringToStringVarP(&completeParams, "params", "p", nil, "-s k1=v1 -s k2=v2")
	StreamCommand.AddCommand(CompleteCommand)
}

func completionTaskOutputParses(resp *http.Response) {
	reader := bufio.NewReader(resp.Body)
	defer resp.Body.Close()
	tmpbuffer := make([]byte, 0)
	currentDescription := ""
	currentOutput := make(map[string]interface{})
	numberOfOutput := 1
	finish := false
	taskID := ""
	for {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}

		if isPrefix {
			tmpbuffer = append(tmpbuffer, line...)
			continue
		}

		if len(tmpbuffer) > 0 {
			line = append(tmpbuffer, line...)
			tmpbuffer = tmpbuffer[:0]
		}

		line = bytes.Trim(line, "")
		if len(line) == 0 {
			if finish {
				fmt.Println("已完成")
				return
			}
			continue
		}

		switch switchEventType(line) {
		case boardcast.StartEvent:
			continue
		case boardcast.WaitEvent:
			continue
		case boardcast.FinishEvent:
			finish = true
			continue
		case boardcast.DeadEvent:
			// 准备捕获错误, 下一次读取会进行错误处理
			finish = true
			continue
		case boardcast.UpdateEvent:
			continue
		}

		if bytes.HasPrefix(line, IDPrefix) {
			continue
		}

		single := false
		if !bytes.HasPrefix(line, DataPrefix) {
			single = true
		}

		line = bytes.TrimPrefix(line, DataPrefix)
		if len(line) == 0 {
			if finish {
				return
			}
			continue
		}
		var msg map[string]interface{}

		err = json.Unmarshal(line, &msg)
		if err != nil {
			continue
		}
		task := msg
		if single {
			task = msg["task"].(map[string]interface{})
		}
		if taskID == "" {
			taskID = task["uuid"].(string)
			fmt.Printf("任务ID: %s\n", taskID)
			fmt.Println(strings.Repeat("-", TableBoxWidth*2) + "\n")
		}
		metadata, ok := task["meta_data"]
		if !ok {
			continue
		}
		metadataMap, ok := metadata.(map[string]interface{})
		if !ok {
			continue
		}
		description, ok := task["current_description"]
		if !ok {
			continue
		}

		var newsParamsCount int
		for k, v := range metadataMap {
			if currentOutput[k] != v {
				currentOutput[k] = v
				fmt.Printf("[%d] %s:\n%v\n\n", numberOfOutput, k, v)
				numberOfOutput++
				newsParamsCount++
			}
		}
		if newsParamsCount > 0 {
			fmt.Println(strings.Repeat("-", TableBoxWidth*2) + "\n")
		}

		if currentDescription != description.(string) {
			currentDescription = description.(string)
			fmt.Printf("当前阶段: %s\n\n", currentDescription)
			fmt.Println(strings.Repeat("-", TableBoxWidth*2) + "\n")
		}

	}
}
