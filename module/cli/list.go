package cli

import (
	"fmt"
	"geep/module/client"
	"geep/module/logger"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
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

		// Define styles
		re := lipgloss.NewRenderer(os.Stdout)
		headerStyle := re.NewStyle().Foreground(lipgloss.Color("5")).Bold(true).Align(lipgloss.Center)
		cellStyle := re.NewStyle().Padding(0, 1)
		borderStyle := re.NewStyle().Foreground(lipgloss.Color("240"))

		// Prepare rows
		var rows [][]string
		for _, elem := range resultMessage.List {
			rows = append(rows, []string{
				elem.Name,
				elem.Status,
				fmt.Sprintf("%d", elem.Recovered),
				fmt.Sprintf("%.2f%%", elem.CPUPercent),
				fmt.Sprintf("%.2f MB", elem.Mem),
			})
		}

		// Create and render table
		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(borderStyle).
			Headers("NAME", "STATUS", "RECOVERED", "CPU", "MEMORY").
			Rows(rows...).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == 0 {
					return headerStyle
				}
				return cellStyle
			})

		fmt.Println(t)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
