package main

import (
	"exportor/proto"
	"msgpacker"
	"fmt"
)

type lbRequest struct {
	req 	func()
}

type user struct {

}

type lobby struct {
	reqChn 		chan *lbRequest
}

func newLobby() *lobby {
	lb := &lobby{}
	lb.reqChn = make(chan *lbRequest, 1024)
	return lb
}

func (lb *lobby) start() {
	lb.handleReq()
}

func (lb *lobby) close() {

}

func (lb *lobby) handleReq() {
	go func() {
		for {
			select {
			case r := <- lb.reqChn:
				r.req()
			}
		}
	}()
}

func (lb *lobby) onUserMessage(uid uint32, message *proto.Message) {
	lb.reqChn <- &lbRequest{
		req: func() {
			if message.Cmd == proto.CmdUserLogin {
				lb.onUserLogin(uid, message.Msg)
			}
		},
	}
}

func (lb *lobby) onServerMessage(uid uint32, message *proto.Message) {
	lb.reqChn <- &lbRequest{
		req: func() {
			if message.Cmd == proto.CmdUserLogin {
				lb.onUserLogin(uid, message.Msg)
			}
		},
	}
}

func (lb *lobby) onDbMessage(fn func()) {
	lb.reqChn <- &lbRequest{
		req: fn,
	}
}

func (lb *lobby) onUserLogin(uid uint32, data []byte) {
	var req proto.UserLogin
	msgpacker.UnMarshal(data, &req)

	fmt.Println("handle user login ", req)
	dbLobbyUserLogin(req.Account, func(accounts *T_Accounts, users *T_Users ,err int) {
		lb.onDbMessage(func() {
			fmt.Println("get acc info ", req.Account, accounts, users)
		})
	})
}

func (lb *lobby) onUserWechatLogin(uid uint32, data []byte) {

}

