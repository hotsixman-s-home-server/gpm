package server

import (
	"bufio"
	"errors"
	"geep/module/types"
	"geep/module/util"
	"io"
	"net"
	"strings"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
)

func (server *Server) connect(conn net.Conn, reader *bufio.Reader, message map[string]any) error {
	var connectRequestMessage types.ConnectRequestMessage
	err := mapstructure.Decode(message, &connectRequestMessage)
	if err != nil {
		return err
	}

	id := ""
	for {
		id = uuid.New().String()
		if server.client[id] == nil {
			break
		}
	}
	client := &ServerSideClient{
		conn:   conn,
		name:   connectRequestMessage.Name,
		reader: reader,
	}
	server.client[id] = client

	err = sendOldLogs(conn, server.pm, connectRequestMessage)
	if err != nil {
		return err
	}

	for {
		JSON, err := client.reader.ReadString('\n')
		if err != nil {
			server.mutex.Lock()
			delete(server.client, id)
			server.mutex.Unlock()
			if errors.Is(err, io.EOF) {
				return nil
			} else {
				return err
			}
		}

		message, err := util.ParseMessage[types.CommandMessage]([]byte(strings.TrimSpace(JSON)))
		if err != nil {
			return err
		}

		if message.Command == "" {
			continue
		}
		if server.pm != nil {
			err := server.pm.Input(client.name, message.Command)
			if err != nil {
				util.SendMessage(conn, &types.LogMessage{
					Type:    "error",
					Message: err.Error(),
				})
			}
		}
	}
}

func sendOldLogs(conn net.Conn, pm types.PMInterface, message types.ConnectRequestMessage) error {
	logs, errors, err := pm.Tail(message.Name, message.Lines)
	if err != nil {
		return err
	}
	for _, log := range logs {
		err := util.SendMessage(conn, &types.LogMessage{
			Type:    "rawlog",
			Message: log,
		})
		if err != nil {
			return err
		}
	}
	for _, errorMsg := range errors {
		err := util.SendMessage(conn, &types.LogMessage{
			Type:    "rawerror",
			Message: errorMsg,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
