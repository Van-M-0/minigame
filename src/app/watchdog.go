package main

import (
	"exportor/defines"
	"exportor/proto"
	"sync"
	"mylog"
	"time"
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
	dog.lb = newLobby(dog)
	return dog
}

func (d *watchdog) clientConnect(client defines.ITcpClient) error {
	mylog.Infoln("client conencted")
	return nil
}

func (d *watchdog) clientDisconnect(client defines.ITcpClient) {
	mylog.Infoln("client disconnected")
	iuid := client.Get("uid")
	if iuid == nil {
		mylog.Infoln("client uid not extis")
		return
	}
	d.cliLock.Lock()
	defer d.cliLock.Unlock()
	uid := iuid.(uint32)
	if _, ok := d.clients[uid]; !ok {
		mylog.Infoln("client not eixts dis", uid)
		return
	}
	d.lb.onUserOffline(uid)
	mylog.Infoln("client disconnected")
}

func (d *watchdog) clientAuth(client defines.ITcpClient) error {
	mylog.Infoln("client auth")
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
	iuid := client.Get("uid")
	if iuid == nil {
		mylog.Infoln("client uid not extis")
		return
	}
	uid := iuid.(uint32)
	if _, ok := d.clients[uid]; !ok {
		mylog.Infoln("client not eixts msg ", uid)
		return
	}

	alive := client.Get("alive")
	if alive == nil {
		client.Set("alive", time.Now().Unix())
		alive = client.Get("alive")
	}

	if message.Cmd == proto.CmdHeartBeat {
		lt, _  := alive.(int64)
		if lt + 5 < time.Now().Unix() {
			client.Set("alive", time.Now().Unix())
			client.Send(message.Cmd, message.Msg)
		}
		return
	}

	d.lb.onUserMessage(uid, message)
}

func (d *watchdog) serverMessage(uid uint32, message *proto.Message) {
	d.lb.onServerMessage(uid, message)
}

func (d *watchdog) sendClientMessage(uid uint32, cmd uint32, data interface{}) {
	d.cliLock.Lock()
	cli, ok := d.clients[uid]
	if ok {
		cli.Send(cmd, data)
	} else {
		mylog.Infoln("client not exits send", uid)
	}
	d.cliLock.Unlock()
}

func (d *watchdog) bcClientMessage(uids []uint32, cmd uint32, data interface{}) {
	d.cliLock.Lock()
	for _, uid := range uids {
		if cli, ok := d.clients[uid]; ok {
			cli.Send(cmd, data)
		} else {
			mylog.Infoln("client not exits bc", uid)
		}
	}
	d.cliLock.Unlock()
}

func (d *watchdog) start() {
	d.lb.start()
}

func (d *watchdog) close() {
	d.lb.close()
}

