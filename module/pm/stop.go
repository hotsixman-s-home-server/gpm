package pm

import "gpm/module/types"

func (pm *PM) Stop(name string) error {
	process := pm.process[name]
	if process == nil {
		return &types.NoProcessError{Name: name}
	}
	if process.status == "running" {
		err := process.cmd.Process.Kill()
		if err != nil {
			return err
		}
	}
	delete(pm.process, name)
	return nil
}
