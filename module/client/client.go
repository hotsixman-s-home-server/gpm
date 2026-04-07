package client

import (
	"bufio"
	"geep/module/util"
	"net"
)

func MakeUDSConn() (conn net.Conn, bufReader *bufio.Reader, err error) {
	socketPath, err := util.GetUDSPath()
	if err != nil {
		return nil, nil, err
	}

	conn, err = net.Dial("unix", socketPath)
	if err != nil {
		return nil, nil, err
	}
	return conn, bufio.NewReader(conn), nil
}
