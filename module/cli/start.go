package cli

import (
	"fmt"
	"gpm/module/client"
	"gpm/module/logger"
	"gpm/module/types"
	"os"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [name]",
	Short: "Start a new process",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		run, err := cmd.Flags().GetString("run")
		if err != nil {
			logger.Errorln("Invalid \"run\" flag.")
			os.Exit(1)
		}

		processArgs, err := cmd.Flags().GetStringSlice("args")
		if err != nil {
			logger.Errorln("Invalid \"args\" flag.")
			os.Exit(1)
		}
		if processArgs == nil {
			processArgs = make([]string, 0)
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

		startMessage := types.StartMessage{
			Type:            "start",
			Name:            args[0],
			Run:             run,
			Args:            processArgs,
			Cwd:             cwd,
			Env:             env,
			MaxRecoverCount: maxRecoverCount,
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
	startCmd.Flags().String("run", "", "Command to execute.")
	startCmd.MarkFlagRequired("run")
	startCmd.Flags().String("cwd", "", "Working directory of the starting process.")
	startCmd.Flags().StringSlice("args", []string{}, "Extra arguments to start the process.")
	startCmd.Flags().StringToString("env", nil, "Set envoriment values for the starting process.")
	startCmd.Flags().Int("max-recover", 10, "Max recover count.")
	rootCmd.AddCommand(startCmd)
}
