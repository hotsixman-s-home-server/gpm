package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var geepdir string // 값을 담을 변수 선언

var rootCmd = &cobra.Command{
	Use:   "geep",
	Short: "GEEP is a process manager for Go applications",
	Long:  `GEEP (Go + Keep) allows you to manage background processes with ease.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if geepdir != "" {
			os.Setenv("GEEP_DIR", geepdir)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&geepdir, "geepdir", "", "Set .geep dir")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
