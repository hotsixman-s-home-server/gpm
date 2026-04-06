package uds

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"gpm/module/logger"
	"gpm/module/types"
	"io"
	"net"
	"strings"
)

type UDSConnectionClient struct {
	conn net.Conn
}

func makeConn() (net.Conn, error) {
	socketPath := GetSocketPath()
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %v", err)
	}
	return conn, nil
}

// Connect
func Connect(name string, closeChan chan bool) (*UDSConnectionClient, error) {
	conn, err := makeConn()
	if err != nil {
		return nil, err
	}

	client := &UDSConnectionClient{
		conn: conn,
	}

	// send message
	message := types.ConnectRequestMessage{
		Type: "connect",
		Name: name,
	}
	JSON, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	_, err = conn.Write([]byte(append(JSON, '\n')))
	if err != nil {
		return nil, err
	}

	// Log & Close
	go func() {
		defer func() {
			closeChan <- true
			close(closeChan)
		}()
		reader := bufio.NewReader(conn)
		for {
			message, err := reader.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					logger.Errorln("Connection Closed.")
				} else {
					logger.Errorln(err)
				}
				return
			}

			var data map[string]string
			err = json.Unmarshal([]byte(strings.TrimSpace(message)), &data)
			if err != nil {
				logger.Errorln(err)
				continue
			}

			switch data["type"] {
			case "log":
				if data["message"] != "" {
					logger.Logln(data["message"])
				}
			case "error":
				if data["message"] != "" {
					logger.Errorln(data["message"])
				}
			}
		}
	}()

	return client, nil
}

func (this *UDSConnectionClient) Command(command string) error {
	message := types.CommandMessage{
		Type:    "command",
		Command: command,
	}

	JSON, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = this.conn.Write(append(JSON, '\n'))
	return err
}

// Start
func Start(startMessage types.StartMessage) (message *types.StartResultMessage, err error) {
	conn, err := makeConn()
	if err != nil {
		return nil, err
	}

	JSON, err := json.Marshal(startMessage)
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(append(JSON, '\n'))
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)
	messageJSON, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	message = &types.StartResultMessage{}
	err = json.Unmarshal([]byte(strings.TrimSpace(messageJSON)), message)
	if err != nil {
		logger.Errorln(err)
		return nil, &types.InvalidMessage{
			JSON: messageJSON,
		}
	}

	return message, nil
}
