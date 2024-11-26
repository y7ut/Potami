package cmd

import (
	"github.com/y7ut/potami/internal/db/migrate"
	"github.com/spf13/cobra"
)

var migrateMethod string

var MigrateCommand = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate Heimdallr DB",
	Run: func(c *cobra.Command, args []string) {
		switch migrateMethod {
		case "up":
			err := migrate.MigrateUp()
			if err != nil {
				panic(err)
			}
		case "down":
			err := migrate.MigrateDown()
			if err != nil {
				panic(err)
			}
		default:
			panic("unknown migrate method")
		}
	},
}

func init() {
	RootCmd.AddCommand(MigrateCommand)
	MigrateCommand.Flags().StringVarP(&migrateMethod, "method", "m", "up", "migrate method: up or down")
}
