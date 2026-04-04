package cli

import (
	"bufio"
	"gpm/module/logger"
	"gpm/module/uds"
	"os"

	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to process",
	Run: func(cmd *cobra.Command, args []string) {
		name, err := cmd.Flags().GetString("name")
		if err != nil || name == "" {
			logger.Errorln(err)
			os.Exit(1)
		}

		client, err := uds.Connect(name)
		if err != nil {
			logger.Errorln(err)
			os.Exit(1)
		}

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			command := scanner.Text()
			client.Command(command)
		}
	},
}

func init() {
	connectCmd.Flags().StringP("name", "", "", "Name of process")
	rootCmd.AddCommand(connectCmd)
}
