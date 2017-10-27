package main

import (
	"exportor/defines"
	"network"
	"exportor/proto"
	"os"
	"fmt"
	"sync"
	"msgpacker"
	"os/exec"
	"path/filepath"
	"mylog"
)

func startClient() {
	c := network.NewTcpClient(&defines.NetClientOption{
		Host: "119.23.239.108:9890",
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
						//Conf: []byte(`{"A":1}`),
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
	wd := newWatchdog()
	gw := network.NewTcpServer(&defines.NetServerOption{
		Host: ":9890",
		ConnectCb: wd.clientConnect,
		CloseCb: wd.clientDisconnect,
		MsgCb: wd.clientMessage,
		AuthCb: wd.clientAuth,
	})
	wd.start()
	gw.Start()
}

func main() {

	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		panic(fmt.Errorf("get exe path err %v", err).Error())
	}
	path, err := filepath.Abs(file)
	if err != nil {
		panic(fmt.Errorf("get file path err %v", err).Error())
	}
	dir, _ := filepath.Split(path)


	//logdir := workdir + "log" + ts + "/"
	logdir := dir + "log" + "/"
	logfile := logdir + "game.log"

	/*
	if err := os.Mkdir(logdir, os.ModePerm); err != nil {
		fmt.Println("create dir failed ", logdir)
	}
	*/

	if true {
		file, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		fmt.Println("open file ", file, logfile)
		if err == nil {
			mylog.SetOutput(file)
			mylog.SetLevel(mylog.DebugLevel)
			mylog.SetFormatter(new(mylog.GameFormatter))
		} else {
			fmt.Println("Failed to log to file, using default stderr", err)
			return
		}

		mylog.Infoln("start programa")
	}


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