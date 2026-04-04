package pm

import (
	"bufio"
	"gpm/module/logger"
	"gpm/module/types"
	"os/exec"
)

type PM struct {
	process map[string]*Process
}

type Process struct {
	name   string
	cmd    *exec.Cmd
	stdin  *bufio.Writer
	stdout *bufio.Reader
	stderr *bufio.Reader
	logger *logger.Logger
}

func (pm *PM) NewProcess(name string, udsServer types.UDSServerInterface, commands ...string) error {
	cmd := exec.Command(commands[0], commands[1:]...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	logger, err := logger.CreateLogger(name, true, udsServer)
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	process := &Process{
		name:   name,
		cmd:    cmd,
		stdin:  bufio.NewWriter(stdin),
		stdout: bufio.NewReader(stdout),
		stderr: bufio.NewReader(stderr),
		logger: logger,
	}
	pm.process[name] = process

	go func() {
		for {
			message, err := process.stdout.ReadString('\n')
			if err != nil {
				return
			}

			process.logger.Logln(message)
		}
	}()
	go func() {
		for {
			message, err := process.stderr.ReadString('\n')
			if err != nil {
				return
			}

			process.logger.Errorln(message)
		}
	}()

	return nil
}

func (pm *PM) Input(name string, message string) {
	if pm.process[name] == nil {
		return
	}

	pm.process[name].stdin.Write([]byte(message))
}
