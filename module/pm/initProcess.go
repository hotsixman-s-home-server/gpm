package pm

import (
	"bufio"
	"fmt"
	"geep/module/logger"
	"geep/module/types"
	"os"
	"os/exec"

	processUtil "github.com/shirou/gopsutil/v3/process"
)

func (pm *PM) initProcess(startMessage types.StartMessage, process *PMProcess) error {
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
	logger, err := logger.CreateLogger(startMessage.Name, true, pm.server, startMessage.MaxLogfileSize)
	if err != nil {
		pm.mainLogger.Errorln(err)
		return err
	}

	err = cmd.Start()
	if err != nil {
		pm.mainLogger.Errorln(err)
		return err
	}

	util, err := processUtil.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		pm.mainLogger.Errorln(err)
		cmd.Process.Kill()
		return err
	}

	pm.mainLogger.Logln("Process started:", startMessage.Name)

	if process == nil {
		process = &PMProcess{
			name:           startMessage.Name,
			status:         "running",
			cmd:            cmd,
			stdin:          stdin,
			stdout:         stdout,
			stderr:         stderr,
			logger:         logger,
			startMessage:   startMessage,
			util:           util,
			recoveredCount: 0,
			autoClean:      true,
		}
		pm.process[startMessage.Name] = process
		pm.processArr = append(pm.processArr, process)
		process.logger.Logln("Process started.")
	} else {
		process.status = "running"
		process.cmd = cmd
		process.stdin = stdin
		process.stdout = stdout
		process.stderr = stderr
		process.util = util
		process.autoClean = true
		process.logger.Logln("Process restarted.")
	}

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
		recoverFlag := false
		err := process.cmd.Wait()

		if !process.autoClean {
			return
		}

		pm.processMutex.Lock()
		defer pm.processMutex.Unlock()
		if err == nil || process.status == "stop" {
			process.logger.Logln(fmt.Sprintf("Process exited. Error: %v", err))
			process.status = "stop"

		} else {
			process.logger.Errorln(fmt.Sprintf("Process exited. Error: %v", err))
			process.status = "error"
			if process.recoveredCount < process.startMessage.MaxRecoverCount {
				process.recoveredCount++
				recoverFlag = true
			}
		}

		process.clean()

		if recoverFlag {
			pm.mainLogger.Errorln("Recovering process:", process.name)
			process.logger.Errorln("Recovering process...")
			go pm.initProcess(startMessage, process)
		}
	}()

	return nil
}
