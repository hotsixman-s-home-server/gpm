package client

import (
	"bufio"
	"gpm/module/types"
	"gpm/module/util"
	"net"
	"strings"
)

func Restart(conn net.Conn, reader *bufio.Reader, restartMessage types.RestartMessage) (message *types.RestartResultMessage, err error) {
	err = util.SendMessage(conn, restartMessage)
	if err != nil {
		return nil, err
	}

	messageJSON, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	message, err = util.ParseMessage[types.RestartResultMessage]([]byte(strings.TrimSpace(messageJSON)))
	if err != nil {
		return nil, err
	}

	return message, nil
}
