package cmd

import (
	"fmt"
	"os"

	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/internal/logger"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:     "potami",
	Short:   "To shine, be bright or resplendent",
	Version: "0.0.1",
}

func init() {
	cobra.OnInitialize(conf.InitConfig)
	cobra.OnInitialize(logger.InitGlobalLogger)
	// Add other Initialize
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
