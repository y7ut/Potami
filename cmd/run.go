package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/y7ut/potami/api"
	"github.com/y7ut/potami/internal/op"
	"github.com/y7ut/potami/internal/server"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run phaino server",
	Run: func(cmd *cobra.Command, args []string) {
		Run()
	},
}

func init() {
	RootCmd.AddCommand(RunCmd)
}

func Run() {
	op.Initialized()

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()
	op.Dispatcher.Start(ctx)
	// op.StartTaskKeeper(ctx)

	server.Initialized()
	server.Route(api.RegisterRouter)
	go server.Start()

	// 阻塞等待退出信号
	<-waitExitSign()
	server.Stop()
	op.TaskQueue.Close()
	op.Dispatcher.Stop()
	op.TaskPool.Stop()

}

func waitExitSign() <-chan os.Signal {
	c := make(chan os.Signal, 2)

	signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}
	// 监听信号
	if !signal.Ignored(syscall.SIGHUP) {
		signals = append(signals, syscall.SIGHUP)
	}

	signal.Notify(c, signals...)

	return c
}
