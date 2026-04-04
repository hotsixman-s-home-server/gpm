package cli

import (
	"gpm/module/logger"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new process",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Logln(args)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
