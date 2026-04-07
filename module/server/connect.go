package server

import (
	"bufio"
	"errors"
	"gpm/module/types"
	"gpm/module/util"
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
