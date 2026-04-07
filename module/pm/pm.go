package pm

import (
	"gpm/module/logger"
	"gpm/module/types"
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
	status       PMProcessStatus
	cmd          *exec.Cmd
	stdin        io.WriteCloser
	stdout       io.ReadCloser
	stderr       io.ReadCloser
	logger       *logger.Logger
	startMessage types.StartMessage
	util         *processUtil.Process
	restartCount int
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

func (pm *PM) Input(name string, message string) {
	if pm.process[name] == nil {
		return
	}

	pm.process[name].stdin.Write([]byte(message))
}
