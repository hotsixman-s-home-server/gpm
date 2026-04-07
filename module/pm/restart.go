package pm

import "gpm/module/types"

func (pm *PM) Restart(message types.RestartMessage) error {
	process := pm.process[message.Name]
	if process == nil {
		return &types.NoProcessError{Name: message.Name}
	}

	process.autoClean = false
	err := pm.Stop(types.StopMessage{Type: "stop", Name: message.Name})
	if err != nil {
		return err
	}

	err = pm.initProcess(process.startMessage, process)
	if err != nil {
		return err
	}

	return nil
}
