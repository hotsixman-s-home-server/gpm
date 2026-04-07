package server

import (
	"bufio"
	"gpm/module/types"
	"gpm/module/util"
	"net"
	"strings"
	"sync"
)

type Server struct {
	mainLogger types.LoggerInterface
	listener   net.Listener
	client     map[string]*ServerSideClient
	mutex      *sync.Mutex
	pm         types.PMInterface
}

type ServerSideClient struct {
	conn   net.Conn
	name   string
	reader *bufio.Reader
}

func (server *Server) SetPM(pm types.PMInterface) {
	server.pm = pm
}

func (server *Server) SetLogger(logger types.LoggerInterface) {
	server.mainLogger = logger
}

func (server Server) Broadcast(name string, JSON []byte) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	for _, client := range server.client {
		if client.name == name {
			go func() {
				client.conn.Write(append(JSON, '\n'))
			}()
		}
	}
}

func (server *Server) accept() {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			continue
		}

		go func() {
			err := server.handleClient(conn)
			if err != nil {
				server.mainLogger.Errorln(err)
			}
		}()
	}
}

func (server *Server) handleClient(conn net.Conn) error {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	messageJSON, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	messagePointer, err := util.ParseMessage[map[string]any]([]byte(strings.TrimSpace(messageJSON)))
	if err != nil {
		return err
	}
	message := *messagePointer

	messageType, ok := message["type"].(string)
	if !ok {
		return &types.InvalidMessage{JSON: messageJSON}
	}

	switch messageType {
	case "connect":
		{
			err := server.connect(conn, reader, message)
			if err != nil {
				server.mainLogger.Errorln(err)
			}
		}
	case "start":
		{
			err := server.start(conn, message)
			if err != nil {
				server.mainLogger.Errorln(err)
			}
		}
	case "stop":
		{
			err := server.stop(conn, message)
			if err != nil {
				server.mainLogger.Errorln(err)
			}
		}
	case "delete":
		{
			err := server.delete(conn, message)
			if err != nil {
				server.mainLogger.Errorln(err)
			}
		}
	case "list":
		{
			err := server.list(conn, message)
			if err != nil {
				server.mainLogger.Errorln(err)
			}
		}
	case "restart":
		{
			err := server.restart(conn, message)
			if err != nil {
				server.mainLogger.Errorln(err)
			}
		}
	}

	return nil
}
