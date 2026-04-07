package cli

import (
	"fmt"
	"gpm/module/client"
	"gpm/module/logger"
	"os"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List processes",
	Run: func(cmd *cobra.Command, args []string) {
		conn, reader, err := client.MakeUDSConn()
		if err != nil {
			logger.Errorln(err)
			os.Exit(1)
		}

		resultMessage, err := client.List(conn, reader)
		if err != nil {
			logger.Errorln(err)
			os.Exit(1)
		}

		fmt.Printf("| %-15s | %-10s | %-10s | %10s | %12s |\n", "Name", "Status", "Recovered", "CPU", "Memory")
		fmt.Println("---------------------------------------------------------")
		for _, elem := range resultMessage.List {
			fmt.Printf("| %-15s | %-10s | %-10d | %9.2f%% | %9.2f MB |\n", elem.Name, elem.Status, elem.Recovered, elem.CPUPercent, elem.Mem)
			fmt.Println("---------------------------------------------------------")
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
