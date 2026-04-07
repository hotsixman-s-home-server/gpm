package cli

import (
	"fmt"
	"geep/module/client"
	"geep/module/logger"
	"geep/module/types"
	"os"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [cmd]",
	Short: "Start a new process",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			logger.Errorln("Invalid \"name\" flag.")
			os.Exit(1)
		}

		cwd, err := cmd.Flags().GetString("cwd")
		if err != nil {
			logger.Errorln("Invalid \"cwd\" flag.")
			os.Exit(1)
		}
		if cwd == "" {
			cwd, err = os.Getwd()
			if err != nil {
				logger.Errorln("Cannot get cwd.")
				os.Exit(1)
			}
		}

		env, err := cmd.Flags().GetStringToString("env")
		if err != nil {
			logger.Errorln("Invalid \"env\" flag.")
			os.Exit(1)
		}
		if env == nil {
			env = make(map[string]string)
		}

		maxRecoverCount, err := cmd.Flags().GetInt("max-recover")
		if err != nil {
			logger.Errorln(err)
			os.Exit(1)
		}

		maxLogfileSize, err := cmd.Flags().GetInt("max-log")
		if err != nil {
			logger.Errorln(err)
			os.Exit(1)
		}

		startMessage := types.StartMessage{
			Type:            "start",
			Name:            name,
			Run:             args[0],
			Args:            args[1:],
			Cwd:             cwd,
			Env:             env,
			MaxRecoverCount: maxRecoverCount,
			MaxLogfileSize:  maxLogfileSize,
		}

		conn, reader, err := client.MakeUDSConn()
		if err != nil {
			logger.Errorln(err)
			os.Exit(1)
		}

		resultMessage, err := client.Start(conn, reader, startMessage)
		if err != nil {
			logger.Errorln(err)
			os.Exit(1)
		}

		if resultMessage.Success {
			logger.Logln(fmt.Sprintf("Successfully started process \"%s\".", startMessage.Name))
			os.Exit(0)
		} else {
			logger.Errorln(resultMessage.Error)
			os.Exit(1)
		}
	},
}

func init() {
	startCmd.Flags().String("name", "", "Set the name of the process.")
	startCmd.MarkFlagRequired("name")
	startCmd.Flags().String("cwd", "", "Working directory of the starting process.")
	startCmd.Flags().StringToString("env", nil, "Set envoriment values for the starting process.")
	startCmd.Flags().Int("max-recover", 10, "Max recover count.")
	startCmd.Flags().Int("max-log", 1024*100, "Max logfile size(KB).")
	rootCmd.AddCommand(startCmd)
}
