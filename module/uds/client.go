package uds

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gpm/module/logger"
	"log"
	"net"
	"strings"
)

type UDSClient struct {
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

func Connect(name string) (*UDSClient, error) {
	conn, err := makeConn()
	if err != nil {
		return nil, err
	}

	client := &UDSClient{
		conn: conn,
	}

	JSON, err := json.Marshal(map[string]string{
		"type": "connect",
		"name": name,
	})
	if err != nil {
		return nil, err
	}
	_, err = conn.Write([]byte(append(JSON, '\n')))
	if err != nil {
		return nil, err
	}

	go func() {
		reader := bufio.NewReader(conn)
		for {
			message, err := reader.ReadString('\n')
			if err != nil {
				log.Println(err)
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

func (this *UDSClient) Command(command string) error {
	JSON, err := json.Marshal(map[string]string{
		"type":    "command",
		"command": command,
	})
	if err != nil {
		return err
	}

	_, err = this.conn.Write(append(JSON, '\n'))
	return err
}
