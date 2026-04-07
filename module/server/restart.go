package server

import (
	"gpm/module/types"
	"gpm/module/util"
	"net"

	"github.com/mitchellh/mapstructure"
)

func (server *Server) restart(conn net.Conn, message map[string]any) error {
	var restartMessage types.RestartMessage
	resultMessage := types.RestartResultMessage{
		Type:    "restartResult",
		Success: false,
		Error:   "",
	}

	err := mapstructure.Decode(message, &restartMessage)
	if err != nil {
		server.mainLogger.Errorln(err)
		resultMessage.Error = err.Error()
		err = util.SendMessage(conn, resultMessage)
		if err != nil {
			return err
		}
	}

	err = server.pm.Restart(restartMessage)
	if err != nil {
		server.mainLogger.Errorln(err)
		resultMessage.Error = err.Error()
		err = util.SendMessage(conn, resultMessage)
		if err != nil {
			return err
		}
	}

	resultMessage.Success = true
	err = util.SendMessage(conn, resultMessage)
	if err != nil {
		return err
	}
	return nil
}
