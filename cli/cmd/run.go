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
	"github.com/y7ut/potami/internal/task"
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

type TaskInfo struct {
	Traces   []*task.TraceRecord    `json:"traces"`
	MetaData map[string]interface{} `json:"meta_data"`
	UUID     string                 `json:"uuid"`
}

func completionTaskOutputParses(resp *http.Response) {
	reader := bufio.NewReader(resp.Body)
	defer resp.Body.Close()
	tmpbuffer := make([]byte, 0)

	finish := false
	taskID := ""
	traceStates := make(map[string]bool, 0)
	MetaData := make(map[string]string, 0)
	displayTrace := func(traceRecords []*task.TraceRecord) []*task.TraceRecord {
		retrys := make([]*task.TraceRecord, 0)
		for _, traceRecord := range traceRecords {
			if traceRecord == nil {
				continue
			}
			if _, ok := traceStates[traceRecord.TraceID]; ok {
				continue
			}
			if traceRecord.FinishAt != "" || traceRecord.Error != "" {
				fmt.Printf("Trace: %s\n", traceRecord.TraceID)
				fmt.Printf("Name: %s\n", traceRecord.Name)
				fmt.Printf("Duration: %d ms\n", traceRecord.Duration)
				fmt.Printf("Bill: $%f\n", traceRecord.Bill)

				fmt.Printf("Options:")
				for k, v := range traceRecord.Options {
					fmt.Printf(" %s: %v ", k, v)
				}
				fmt.Print("\n")
				fmt.Println(strings.Repeat("-", TableBoxWidth*2) + "\n")

				fmt.Println(" - Inputs:")
				for _, input := range traceRecord.Inputs {
					fmt.Printf("[%s]\n\n", input)
					attributeOfInput, ok := MetaData[input]

					if ok {
						fmt.Printf("%s\n", attributeOfInput)
					} else {
						fmt.Printf("N/A\n")
					}
				}
				fmt.Println(strings.Repeat("-", TableBoxWidth) + "\n")
				if traceRecord.Error != "" {
					fmt.Printf(" - Error: \n %s\n", traceRecord.Error)
					fmt.Println(strings.Repeat("-", TableBoxWidth*2) + "\n")
					traceStates[traceRecord.TraceID] = true
					if traceRecord.Retrys != nil {
						retrys = append(retrys, traceRecord.Retrys...)
					}
					continue
				}
				fmt.Println(" - Outputs:")
				for _, output := range traceRecord.Outputs {
					fmt.Printf("[%s]\n\n", output)
					attributeOfOutput, ok := MetaData[output]
					if ok {
						fmt.Printf("%s\n", attributeOfOutput)
					} else {
						fmt.Printf("N/A\n")
					}
				}
				fmt.Println(strings.Repeat("-", TableBoxWidth*2) + "\n")
				traceStates[traceRecord.TraceID] = true
			}
		}
		return retrys
	}
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
		var t TaskInfo
		if single {
			var msg struct {
				Task TaskInfo `json:"task"`
			}
			err = json.Unmarshal(line, &msg)
			if err != nil {
				continue
			}
			t = msg.Task
		} else {
			var msg TaskInfo
			err = json.Unmarshal(line, &msg)
			if err != nil {
				continue
			}
			t = msg
		}

		if taskID == "" {
			taskID = t.UUID
			fmt.Printf("任务ID: %s\n", taskID)
			fmt.Println(strings.Repeat("-", TableBoxWidth*2) + "\n")
		}

		for k, v := range t.MetaData {
			MetaData[k] = editStringSlim(v.(string), 1000)
		}

		retrys := displayTrace(t.Traces)
		if len(retrys) > 0 {
			fmt.Println("Retrys:")
			fmt.Println(strings.Repeat("-", TableBoxWidth*2) + "\n")
			displayTrace(retrys)
		}
	}
}
