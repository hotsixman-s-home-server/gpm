package uds

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gpm/module/logger"
	"gpm/module/types"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type UDSServer struct {
	log      *logger.Logger
	listener net.Listener
	clients  map[string]*serverSideClient
	mutex    *sync.Mutex
	pm       types.PMInterface
}

type serverSideClient struct {
	conn   net.Conn
	name   string
	reader *bufio.Reader
	writer *bufio.Writer
}

func Listen(log *logger.Logger) (*UDSServer, error) {
	socketPath := GetSocketPath()

	if _, err := os.Stat(socketPath); err == nil {
		os.Remove(socketPath)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on uds: %v", err)
	}

	server := &UDSServer{
		log:      log,
		listener: listener,
		clients:  make(map[string]*serverSideClient),
		mutex:    &sync.Mutex{},
		pm:       nil,
	}

	server.accept()

	return server, nil
}

func (this *UDSServer) SetPM(pm types.PMInterface) {
	this.pm = pm
}

func (this *UDSServer) Broadcast(name string, JSON []byte) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	for _, cli := range this.clients {
		if cli.name == name {
			go func() {
				cli.conn.Write(append(JSON, '\n'))
			}()
		}
	}
}

func (this *UDSServer) accept() {
	go func() {
		for {
			conn, err := this.listener.Accept()
			if err != nil {
				continue
			}

			go this.handleClient(conn)
		}
	}()
}

func (this *UDSServer) checkClient(conn net.Conn) (*bufio.Reader, map[string]any, error) {
	reader := bufio.NewReader(conn)

	JSON, err := reader.ReadString('\n')
	if err != nil {
		return nil, nil, err
	}
	JSON = strings.TrimSpace(JSON)

	var message map[string]any
	err = json.Unmarshal([]byte(JSON), &message)
	if err != nil {
		return nil, nil, err
	}

	return reader, message, nil
}

func (this *UDSServer) handleClient(conn net.Conn) {
	defer conn.Close()
	reader, message, err := this.checkClient(conn)
	if err != nil {
		this.log.Errorln(err)
		return
	}

	if _, ok := message["type"].(string); !ok {
		this.log.Errorln("Invalid message:", message)
		return
	}

	switch message["type"] {
	case "connect":
		{
			name, nameOk := message["name"].(string)
			if !nameOk {
				return
			}

			id := ""
			for {
				id = uuid.New().String()
				if this.clients[id] == nil {
					break
				}
			}
			client := &serverSideClient{
				conn:   conn,
				name:   name,
				reader: reader,
				writer: bufio.NewWriter(conn),
			}
			this.clients[id] = client

			for {
				JSON, err := client.reader.ReadString('\n')
				if err != nil {
					this.mutex.Lock()
					delete(this.clients, id)
					this.mutex.Unlock()
					return
				}

				var data map[string]string
				err = json.Unmarshal([]byte(strings.TrimSpace(JSON)), &data)
				if err != nil {
					continue
				}

				if data["type"] == "command" {
					if data["command"] == "" {
						continue
					}
					if this.pm != nil {
						this.pm.Input(client.name, data["command"])
					}
				}
			}
		}
	case "start":
		{
			this.start(conn, message)
		}
	}
}
