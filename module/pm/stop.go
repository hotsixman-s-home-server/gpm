package pm

import (
	"gpm/module/types"
)

func (pm *PM) Stop(message types.StopMessage) error {
	pm.processMutex.Lock()
	defer pm.processMutex.Unlock()

	process := pm.process[message.Name]
	if process == nil {
		return &types.NoProcessError{Name: message.Name}
	}
	if process.status == "running" {
		process.autoClean = false
		err := process.cmd.Process.Kill()
		if err != nil {
			pm.mainLogger.Logln("Cannot stop process: ", message.Name)
			return err
		}
		process.status = "stop"
		process.clean()
		process.logger.Logln("Process stopped.")
		pm.mainLogger.Logln("Process stopped:", message.Name)
	}
	return nil
}
