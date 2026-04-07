package pm

import (
	"geep/module/logger"
	"geep/module/types"
	"io"
	"os/exec"
	"sync"

	processUtil "github.com/shirou/gopsutil/v3/process"
)

type PM struct {
	process      map[string]*PMProcess
	processArr   []*PMProcess
	mainLogger   *logger.Logger
	server       types.ServerInterface
	processMutex *sync.Mutex
}

type PMProcessStatus string

type PMProcess struct {
	name string
	// 'running'|'stop'|'error'
	status         PMProcessStatus
	cmd            *exec.Cmd
	stdin          io.WriteCloser
	stdout         io.ReadCloser
	stderr         io.ReadCloser
	logger         *logger.Logger
	startMessage   types.StartMessage
	util           *processUtil.Process
	recoveredCount int
	autoClean      bool
}

func NewPM(mainLogger *logger.Logger) *PM {
	pm := &PM{
		process:      make(map[string]*PMProcess),
		mainLogger:   mainLogger,
		server:       nil,
		processMutex: &sync.Mutex{},
	}

	return pm
}

func (pm *PM) SetServer(server types.ServerInterface) {
	pm.server = server
}

func (pm *PM) Input(name string, message string) error {
	process := pm.process[name]
	if process == nil {
		return &types.NoProcessError{Name: name}
	}
	if process.status != "running" {
		return &types.ProcessNotRunningError{Name: name}
	}

	pm.process[name].stdin.Write(append([]byte(message), '\n'))
	return nil
}

func (pm *PM) Tail(name string, lineCount int) ([]string, []string, error) {
	process := pm.process[name]
	if process == nil {
		return []string{}, []string{}, nil
	}

	logs, err := process.logger.TailLogs(lineCount)
	if err != nil {
		return []string{}, []string{}, nil
	}

	errors, err := process.logger.TailErrors(lineCount)
	if err != nil {
		return []string{}, []string{}, nil
	}

	return logs, errors, nil
}

func (process *PMProcess) clean() {
	process.stdin.Close()
	process.stdout.Close()
	process.stderr.Close()
	process.stdin = nil
	process.stdout = nil
	process.stderr = nil
	process.cmd = nil
	process.util = nil
}
