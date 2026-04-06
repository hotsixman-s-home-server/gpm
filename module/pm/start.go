package pm

import (
	"bufio"
	"fmt"
	"gpm/module/logger"
	"gpm/module/types"
	"os"
	"os/exec"
)

func (pm *PM) Start(startMessage types.StartMessage) error {
	formerProcess := pm.process[startMessage.Name]
	if formerProcess == nil {
		return pm.NewProcess(startMessage)
	} else if formerProcess.status == "stop" || formerProcess.status == "error" {
		err := pm.Stop(startMessage.Name)
		if err != nil {
			return err
		}
		return pm.NewProcess(startMessage)
	} else {
		return &types.ProcessRunningError{Name: startMessage.Name}
	}
}

func (pm *PM) NewProcess(startMessage types.StartMessage) error {
	pm.processMutex.Lock()
	defer pm.processMutex.Unlock()

	cmd := exec.Command(startMessage.Run, startMessage.Args...)
	cmd.Dir = startMessage.Cwd
	env := os.Environ()
	for k, v := range startMessage.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	stdin, err := cmd.StdinPipe()
	if err != nil {
		pm.mainLogger.Errorln(err)
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		pm.mainLogger.Errorln(err)
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		pm.mainLogger.Errorln(err)
		return err
	}
	logger, err := logger.CreateLogger(startMessage.Name, true, pm.server)
	if err != nil {
		pm.mainLogger.Errorln(err)
		return err
	}

	err = cmd.Start()
	if err != nil {
		pm.mainLogger.Errorln(err)
		return err
	}
	pm.mainLogger.Logln("Process started:", startMessage.Name)

	process := &PMProcess{
		name:         startMessage.Name,
		status:       "start",
		cmd:          cmd,
		stdin:        stdin,
		stdout:       stdout,
		stderr:       stderr,
		logger:       logger,
		startMessage: startMessage,
	}
	pm.process[startMessage.Name] = process

	go func() {
		scanner := bufio.NewScanner(process.stdout)
		for scanner.Scan() {
			process.logger.Logln(scanner.Text())
		}
	}()
	go func() {
		scanner := bufio.NewScanner(process.stderr)
		for scanner.Scan() {
			process.logger.Errorln(scanner.Text())
		}
	}()

	go func() {
		err := process.cmd.Wait()
		if err == nil {
			process.logger.Logln(fmt.Sprintf("Process exited. Error: %v", err))
			process.status = "stop"

		} else {
			process.logger.Errorln(fmt.Sprintf("Process exited. Error: %v", err))
			process.status = "error"
		}

		process.stdin.Close()
		process.stdout.Close()
		process.stderr.Close()
	}()

	return nil
}
