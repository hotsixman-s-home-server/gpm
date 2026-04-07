package client

import (
	"bufio"
	"errors"
	"geep/module/logger"
	"geep/module/types"
	"geep/module/util"
	"io"
	"net"
)

type Client struct {
	conn   net.Conn
	reader *bufio.Reader
}

func NewClient(name string, conn net.Conn, reader *bufio.Reader, lineCount int, messageChan chan types.LogMessage, closeChan chan bool) (*Client, error) {
	client := &Client{
		conn:   conn,
		reader: reader,
	}

	// Send connection request message
	message := types.ConnectRequestMessage{
		Type:  "connect",
		Name:  name,
		Lines: lineCount,
	}
	err := util.SendMessage(conn, message)
	if err != nil {
		return nil, err
	}

	// @todo Need to check ConnectResponseMessage

	go func() {
		// close
		defer func() {
			closeChan <- true
			close(closeChan)
		}()
		// log
		for {
			messageJSON, err := reader.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					logger.Logln("Connection Closed.")
				} else {
					logger.Errorln(err)
				}
				return
			}

			message, err := util.ParseMessage[types.LogMessage]([]byte(messageJSON))
			if err != nil {
				logger.Errorln(err)
				continue
			}

			messageChan <- *message
		}
	}()

	return client, nil
}

func (client *Client) Command(command string) error {
	message := types.CommandMessage{
		Type:    "command",
		Command: command,
	}

	return util.SendMessage(client.conn, message)
}
