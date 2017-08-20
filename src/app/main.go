package main

import (
	"exportor/defines"
	"network"
	"exportor/proto"
	"os"
	"fmt"
	"sync"
	"msgpacker"
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
			if message.Cmd == proto.CmdUserLogin {
				var res proto.UserLoginRet
				msgpacker.UnMarshal(message.Msg, &res)
				fmt.Println("client login res ", res)
				if res.ErrCode == "ok" {
					client.Send(proto.CmdUserCreateRoom, &proto.UserCreateRoom{
						Kind: 1,
						Conf: []byte(`{"A":1}`),
					})
				}
			} else if message.Cmd == proto.CmdUserCreateRoom {
				var res proto.UserCreateRoomRet
				msgpacker.UnMarshal(message.Msg, &res)
				fmt.Println("client create room ret ", res)
				if res.ErrCode == "ok" {
					client.Send(proto.CmdUserEnterRoom, &proto.UserEnterRoom{
						RoomId: 799523,
						Test:`{"Test":112}`,
					})
				}
			} else if message.Cmd == proto.CmdUserEnterRoom {
				var res proto.UserEnterRoomRet
				msgpacker.UnMarshal(message.Msg, &res)
				fmt.Println("client enter room ret ", res)
				if res.ErrCode == "ok" {
					client.Send(proto.CmdUserGameMessage, &proto.UserMessage{
						Cmd: 1,
					})
				}
			} else if message.Cmd == proto.CmdUserGameMessage {
				var res proto.UserMessageRet
				msgpacker.UnMarshal(message.Msg, &res)
				fmt.Println("game message ", res, res.Msg)
				if res.Cmd == proto.CmdEgUserReady {

				}
			}
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

	firstCard := 53
	firstColor := (firstCard & 0xF0) >> 4
	fmt.Println(firstCard, firstColor, firstColor << 4 + 0x03)
	minval := firstColor << 4 + 0x03
	fmt.Println(minval)

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