package main

import (
	"exportor/defines"
	"exportor/proto"
	"fmt"
	"sync"
)

type connection struct {
	defines.ITcpClient
}

type watchdog struct {
	cliLock 	sync.RWMutex
	clients 	map[uint32]*connection
	cliIdGen 	uint32
	lb 			*lobby
}

func newWatchdog() *watchdog {
	dog := &watchdog{}
	dog.clients = make(map[uint32]*connection)
	dog.lb = newLobby()
	return dog
}

func (d *watchdog) clientConnect(client defines.ITcpClient) error {
	fmt.Println("client conencted")
	return nil
}

func (d *watchdog) clientDisconnect(client defines.ITcpClient) {
	fmt.Println("client disconnected")
}

func (d *watchdog) clientAuth(client defines.ITcpClient) error {
	fmt.Println("client auth")
	d.cliLock.Lock()
	d.cliIdGen++
	id := d.cliIdGen + 1
	client.Set("uid", id)
	d.clients[id] = &connection{
		ITcpClient:	client,
	}
	d.cliLock.Unlock()
	return nil
}

func (d *watchdog) clientMessage(client defines.ITcpClient, message *proto.Message) {
	fmt.Println("client message", message)
	iuid := client.Get("uid")
	if iuid == nil {
		fmt.Println("client uid not extis")
		return
	}
	uid := iuid.(uint32)
	if _, ok := d.clients[uid]; !ok {
		fmt.Println("client not eixts")
		return
	}
	d.lb.onUserMessage(uid, message)
}

func (d *watchdog) serverMessage(uid uint32, message *proto.Message) {
	d.lb.onServerMessage(uid, message)
}

func (d *watchdog) sendClientMessage(uid uint32, message *proto.Message) {
	cli, ok := d.clients[uid]
	if !ok {
		fmt.Println("client not exits", uid)
		return
	}
	cli.Send(message.Cmd, message.Msg)
}

func (d *watchdog) start() {
	d.lb.start()
}

func (d *watchdog) close() {
	d.lb.close()
}

