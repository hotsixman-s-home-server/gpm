package uds

import (
	"encoding/json"
	"gpm/module/types"
	"net"

	"github.com/mitchellh/mapstructure"
)

// Actions
func (this *UDSServer) start(conn net.Conn, message map[string]any) {
	var startMessage types.StartMessage
	resultMessage := types.StartResultMessage{
		Type:    "startResult",
		Success: false,
		Error:   "",
	}

	err := mapstructure.Decode(message, &startMessage)
	if err != nil {
		this.log.Errorln(err)
		resultMessage.Error = err.Error()
		JSON, err := json.Marshal(resultMessage)
		if err != nil {
			return
		}
		conn.Write(append(JSON, '\n'))
	}

	err = this.pm.Start(startMessage)
	if err != nil {
		this.log.Errorln(err)
		resultMessage.Error = err.Error()
		JSON, err := json.Marshal(resultMessage)
		if err != nil {
			return
		}
		conn.Write(append(JSON, '\n'))
	}

	resultMessage.Success = true
	JSON, err := json.Marshal(resultMessage)
	if err != nil {
		return
	}
	conn.Write(append(JSON, '\n'))
}
