package network

import (
	"exportor/defines"
	"net"
	"fmt"
)

type tcpServer struct {
	*netContext
	opt 			*defines.NetServerOption
	closeChn 		chan int
}

func newServer(opt *defines.NetServerOption) *tcpServer {
	server := &tcpServer{
		netContext: newNetContext(),
		opt: opt,
		closeChn: make(chan int),
	}
	return server
}

func (server *tcpServer) Start() error {
	l, err := net.Listen("tcp", server.opt.Host)
	if err != nil {
		return err
	}
	defer func() {
		l.Close()
	}()

		for {
			fmt.Println("server start ", server.opt.Host)
			conn, err := l.Accept()
			fmt.Println("server start ", server.opt.Host, conn, err)
			if err != nil {
				continue
			}

			go func() {
				server.handleClient(conn)
			}()
		}

	return nil
}

func (server *tcpServer) Stop() error {
	return nil
}

func (server *tcpServer) handleClient(conn net.Conn) {
	fmt.Println("handle client ", conn)
	client := newTcpClient(&defines.NetClientOption{
	})
	client.configureConn(conn)

	defer func() {
		client.Close()
		server.opt.CloseCb(client)
	}()

	server.opt.ConnectCb(client)
	if server.opt.AuthCb(client) != nil {
		return
	}
	go client.sendLoop()

	for {
		m, err := client.readMessage()
		if err != nil {
			fmt.Println("decode msg error", err)
			return
		}
		server.opt.MsgCb(client, m)
	}
}

