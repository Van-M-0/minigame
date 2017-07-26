package main

import (
	"exportor/defines"
	"network"
	"exportor/proto"
	"os"
	"fmt"
	"sync"
)

func startClient() {
	c := network.NewTcpClient(&defines.NetClientOption{
		Host: ":9890",
		ConnectCb: func(client defines.ITcpClient) error {
			return nil
		},
		CloseCb: func(client defines.ITcpClient) {

		},
		MsgCb: func(client defines.ITcpClient, message *proto.Message) {
		},
		AuthCb: func(defines.ITcpClient) error {
			return nil
		},
	})
	c.Connect()
	c.Send(proto.CmdUserLogin, &proto.UserLogin{
		Account: "hello",
	})
}

func startServer() {
	hp := newHttpServer()
	wd := newWatchdog()
	gw := network.NewTcpServer(&defines.NetServerOption{
		Host: ":9890",
		ConnectCb: wd.clientConnect,
		CloseCb: wd.clientDisconnect,
		MsgCb: wd.clientMessage,
		AuthCb: wd.clientAuth,
	})
	hp.start()
	wd.start()
	gw.Start()
}

func main() {
	p := os.Args[1]
	fmt.Println("start args ", p)

	if p == "client" {
		startClient()
	} else if p == "server" {
		startServer()
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)
	wg.Wait()
}