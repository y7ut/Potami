package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/y7ut/potami/internal/schema"
	"github.com/y7ut/potami/pkg/json"
)

var (
	applyStreamFile string
)

var currentContext *PotamiServiceContext
var client *resty.Client

var StreamCommand = &cobra.Command{
	Use:   "stream",
	Short: "Potami Stream",
}

var ApplyCommand = &cobra.Command{
	Use:    "apply",
	Short:  "Apply Stream",
	PreRun: InitContextAndClient,
	Run: func(c *cobra.Command, args []string) {
		var streamApplyConfigBytes []byte

		if c.Flag("input").Value.String() != "yaml" && c.Flag("input").Value.String() != "json" {
			log.Fatalf("invalid input type: %s", c.Flag("input").Value.String())
			return
		}
		if applyStreamFile != "" {
			configOfFile, err := os.ReadFile(applyStreamFile)
			if err != nil {
				log.Fatalf("failed to read stream file: %v", err)
			}
			streamApplyConfigBytes = configOfFile
		} else if len(args) > 0 {
			streamApplyConfigBytes = []byte(args[0])
		} else {

			fi, err := os.Stdin.Stat()
			if err != nil {
				log.Fatalf("failed to read stream from stdin: %v", err)
			}

			if fi.Mode()&os.ModeNamedPipe == 0 {
				log.Fatalf("stdin is not a pipe")
			}
			bytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				log.Fatalf("failed to read stream from stdin: %v", err)
			}
			streamApplyConfigBytes = bytes
		}

		if len(streamApplyConfigBytes) == 0 {
			log.Fatal("no stream config provided")
		}
		if strings.TrimSpace(string(streamApplyConfigBytes)) == "" {
			log.Fatal("stream config is empty")
		}

		contentType := "application/json"

		if c.Flag("input").Value.String() == "yaml" {
			contentType = "application/x-yaml"
		}

		type ApplyResponseErr struct {
			Error string `json:"error"`
		}

		var applyErr ApplyResponseErr

		resp, err := client.SetDisableWarn(true).R().
			SetHeader("Content-Type", contentType).
			SetHeader("Accept", contentType).
			SetBody(streamApplyConfigBytes).
			SetError(&applyErr).
			Post("/api/stream")

		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode() != 200 {
			log.Fatal(applyErr.Error)
		}

		fmt.Println(resp.String())
	},
}

var ListCommand = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List Stream",
	PreRun:  InitContextAndClient,
	Run: func(c *cobra.Command, args []string) {
		var res []schema.HumanFriendlyStreamConfig
		resp, err := client.SetDisableWarn(true).R().
			SetHeader("Content-Type", "application/json").
			SetResult(&res).
			Get("/api/stream")

		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode() != 200 {
			log.Fatalf("failed to get stream: %s", resp.String())
		}

		table := NewTable([]string{"NAME", "DESCRIPTION", "JOBS"})
		for _, v := range res {
			jobsDetail := make([]string, 0)
			for _, job := range v.Jobs {
				desc, ok := job["description"]
				if !ok {
					desc = "N/A"
				}
				jobsDetail = append(jobsDetail, desc)
			}
			jobsLine := strings.Join(jobsDetail, ">>")

			table.AddRow([]string{string(v.Name), string(v.Description), jobsLine})

		}

		table.Render()
	},
}

var InfoCommand = &cobra.Command{
	Use:    "info",
	Short:  "Info Stream",
	PreRun: InitContextAndClient,
	Run: func(c *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatalf("stream info requires at least 1 argument")
		}
		streamToFind := args[0]

		if c.Flag("output").Changed && c.Flag("output").Value.String() == "yaml" {

			resp, err := client.SetDisableWarn(true).R().
				SetHeader("Content-Type", "application/json").
				SetHeader("Accept", "application/x-yaml").
				Get("/api/stream/" + streamToFind)
			if err != nil {
				log.Fatal(err)
			}
			if resp.StatusCode() != 200 {
				log.Fatalf("failed to get stream: %s", resp.String())
			}
			fmt.Println(resp.String())
			return
		}

		var res schema.Stream
		resp, err := client.SetDisableWarn(true).R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetResult(&res).
			Get("/api/stream/" + streamToFind)

		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode() != 200 {
			log.Fatalf("failed to get stream: %s", resp.String())
		}
		if c.Flag("output").Changed {
			output, err := json.MarshalIndent(res, "", "  ")
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(output))
			return
		}
		table := NewTable([]string{"NAME", "DESCRIPTION", "TYPE", "PARAMS", "OUTPUTS"})
		for _, v := range res.Jobs {

			paramsLine := strings.Join(v.Params, ",")

			outputsLine := strings.Join(v.Output, ",")
			if v.Type == "api_tool" {
				for k, o := range v.OutputParses {
					outputsLine += fmt.Sprintf("%s:%s,", k, o)
				}
				outputsLine = strings.TrimSuffix(outputsLine, ",")
			}
			if v.Type == "search" {
				paramsLine += v.QueryField
				outputsLine += v.OutputField
			}
			table.AddRow([]string{v.Name, v.Description, v.Type, paramsLine, outputsLine})
		}

		table.Render()
	},
}

var RemoveCommand = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm"},
	Short:   "Remove Stream",
	PreRun:  InitContextAndClient,
	Run: func(c *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatalf("required stream name")
		}
		streamToRemove := args[0]

		resp, err := client.SetDisableWarn(true).R().
			SetHeader("Content-Type", "application/json").
			Delete("/api/stream/" + streamToRemove)

		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode() != 204 {
			log.Fatalf("failed to get stream: %s", resp.String())
		}

		fmt.Printf("stream %s removed successfully\n", streamToRemove)
	},
}

func init() {
	ApplyCommand.Flags().StringVarP(&applyStreamFile, "file", "f", "", "the stream config file")

	InfoCommand.Flags().StringP("context", "c", "", "the context used")
	InfoCommand.Flags().BoolP("detail", "d", false, "whether to show detail info")
	InfoCommand.Flags().StringP("output", "o", "yaml", "the output schema of info")
	ListCommand.Flags().StringP("context", "c", "", "the context used")
	ApplyCommand.Flags().StringP("input", "i", "yaml", "the input schema of apply")
	ApplyCommand.Flags().StringP("context", "c", "", "the context used")
	RemoveCommand.Flags().StringP("context", "c", "", "the context used")
	StreamCommand.AddCommand(ApplyCommand)
	StreamCommand.AddCommand(ListCommand)
	StreamCommand.AddCommand(InfoCommand)
	StreamCommand.AddCommand(RemoveCommand)

	RootCmd.AddCommand(StreamCommand)
}

func InitContextAndClient(cmd *cobra.Command, args []string) {
	if cmd.Flag("context").Value.String() != "" {
		contexts := getContexts()
		providedContext, ok := contexts[cmd.Flag("context").Value.String()]
		if !ok {
			log.Fatal("provided context not found")
		}
		currentContext = &providedContext
	} else {
		defaultContext, err := getCurrentContext()
		if err != nil {
			log.Fatal(err)
		}
		currentContext = defaultContext
	}

	client = resty.New().SetBaseURL(currentContext.Endpoint)
}
