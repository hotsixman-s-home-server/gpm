package pm

import "gpm/module/types"

func (pm *PM) Delete(message types.DeleteMessage) error {
	process := pm.process[message.Name]
	if process == nil {
		return &types.NoProcessError{Name: message.Name}
	}

	err := pm.Stop(types.StopMessage{
		Type: "stop",
		Name: message.Name,
	})
	if err != nil {
		return err
	}

	pm.processMutex.Lock()
	defer pm.processMutex.Unlock()
	delete(pm.process, message.Name)
	for i, process := range pm.processArr {
		if process.name == message.Name {
			pm.processArr = append(pm.processArr[:i], pm.processArr[i+1:]...)
		}
	}

	return nil
}
