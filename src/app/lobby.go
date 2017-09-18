package main

import (
	"exportor/proto"
	"msgpacker"
	"fmt"
	"time"
)

type lbRequest struct {
	req 	func()
}

type userInfo struct {
	Cid 		uint32		`json:"-"`
	Account 	string
	UserId 		int
	UserName 	string
	Sex 		uint8
	RoomCard 	int
	RoomId 		int
	Coins 		int
	offline 	time.Time	`json:"-"`
}

type lobby struct {
	dog 		*watchdog
	reqChn 		chan *lbRequest
	users 		map[int]*userInfo
	uidUsers 	map[uint32]*userInfo
	roomMgr 	*roomMgr
}

func newLobby(dog *watchdog) *lobby {
	lb := &lobby{}
	lb.reqChn = make(chan *lbRequest, 1024)
	lb.users = make(map[int]*userInfo)
	lb.uidUsers = make(map[uint32]*userInfo)
	lb.dog = dog
	lb.roomMgr = newRoomMgr(lb)
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
			} else if message.Cmd == proto.CmdUserLogout {
				lb.onUserLogout(uid)
			} else {
				if user, ok := lb.uidUsers[uid]; !ok {
					fmt.Println("******** user not in *********", message.Cmd)
					lb.dog.sendClientMessage(uid, proto.CmdCommonError, &proto.UserCommonError{
						Cmd: message.Cmd,
						ErrCode: "userNotIn",
					})
				} else {
					fmt.Println("-------- user message -------", message.Cmd, user)
					lb.roomMgr.onMessage(user, message.Cmd, message.Msg)
				}
			}
		},
	}
}

func (lb *lobby) onUserOffline(uid uint32) {
	lb.reqChn <- &lbRequest{
		req: func() {
			if user, ok := lb.uidUsers[uid]; ok {
				user.offline = time.Now()
				lb.roomMgr.onUserOffline(user)
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

func (lb *lobby) onUserWechatLogin(uid uint32, data []byte) {

}

func (lb *lobby) onUserLogin(uid uint32, data []byte) {
	var req proto.UserLogin
	msgpacker.UnMarshal(data, &req)

	fmt.Println("handle user login ", req)
	dbLobbyUserLogin(req.Account, func(acc *T_Accounts, u *T_Users ,err int) {
		lb.onDbMessage(func() {
			if err == 0 {
				var uu *userInfo
				reenter := false
				if u1, ok := lb.users[int(u.Userid)]; ok {
					delete(lb.uidUsers, u1.Cid)
					u1.Cid = uid
					lb.uidUsers[uid] = u1
					fmt.Println("lb user reconneted", u1)
					reenter = true
					uu = u1
				} else {
					uu = &userInfo{
						Cid: uid,
						Account: u.Account,
						UserId: int(u.Userid),
						UserName: u.Name,
						Sex: u.Sex,
						RoomCard: int(u.Roomcard),
						RoomId: int(u.Roomid),
						Coins: int(u.Coins),
					}
					lb.users[uu.UserId] = uu
					lb.uidUsers[uid] = uu
				}
				lb.dog.sendClientMessage(uid, proto.CmdUserLogin, &proto.UserLoginRet{
					ErrCode: "ok",
					User: uu,
				})
				if reenter {
					lb.roomMgr.onUserReconnect(uu)
				}
			} else {
				lb.dog.sendClientMessage(uid, proto.CmdUserLogin, &proto.UserLoginRet{
					ErrCode: "error",
				})
			}
		})
	})
}

func (lb *lobby) onUserLogout(uid uint32) {

}


