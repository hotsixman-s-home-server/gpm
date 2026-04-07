package cli

import (
	"fmt"
	"gpm/module/client"
	"gpm/module/logger"
	"gpm/module/types"
	"os"

	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart [name]",
	Short: "Restart process",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		restartMessage := types.RestartMessage{
			Type: "restart",
			Name: args[0],
		}

		conn, reader, err := client.MakeUDSConn()
		if err != nil {
			logger.Errorln(err)
			os.Exit(1)
		}

		resultMessage, err := client.Restart(conn, reader, restartMessage)
		if err != nil {
			logger.Errorln(err)
			os.Exit(1)
		}

		if resultMessage.Success {
			logger.Logln(fmt.Sprintf("Successfully restarted process \"%s\".", restartMessage.Name))
			os.Exit(0)
		} else {
			logger.Errorln(resultMessage.Error)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
