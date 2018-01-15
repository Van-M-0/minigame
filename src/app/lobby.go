package main

import (
	"exportor/proto"
	"msgpacker"
	"time"
	"mylog"
	"math/rand"
	"strconv"
)

type lbRequest struct {
	req 	func()
}

type userInfo struct {
	Cid 		uint32		`json:"-"`
	Account 	string
	UserId 		int
	UserName 	string
	HeadImg 	string
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
	hp 			*httpServer
}

func newLobby(dog *watchdog) *lobby {
	lb := &lobby{}
	lb.reqChn = make(chan *lbRequest, 1024)
	lb.users = make(map[int]*userInfo)
	lb.uidUsers = make(map[uint32]*userInfo)
	lb.dog = dog
	lb.roomMgr = newRoomMgr(lb)
	lb.hp = newHttpServer(lb)
	return lb
}

func (lb *lobby) start() {
	lb.hp.start()
	dbCheckConnection()
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
					mylog.Infoln("******** user not in *********", message.Cmd)
					lb.dog.sendClientMessage(uid, proto.CmdCommonError, &proto.UserCommonError{
						Cmd: message.Cmd,
						ErrCode: "userNotIn",
					})
				} else {
					mylog.Infoln("-------- user message -------", message.Cmd, user)
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

func (lb *lobby) onDebug(room int, w, r interface {}) {
	lb.reqChn <- &lbRequest{
		req: func(){
			lb.roomMgr.onDebug(room , w, r)
		},
	}
}

func (lb *lobby) onUserWechatLogin(uid uint32, data []byte) {

}

func (lb *lobby) onUserLogin(uid uint32, data []byte) {
	var req proto.UserLogin
	msgpacker.UnMarshal(data, &req)

	var errCode, acc, token, nickName, headImg string
	var sex int

	loginHandler := func() {
		mylog.Infoln("handle user login ", req)
		dbLobbyUserLogin(acc, nickName, headImg, uint8(sex), func(acc *T_Accounts, u *T_Users ,err int) {
			lb.onDbMessage(func() {
				if err == 0 {
					var uu *userInfo
					if u1, ok := lb.users[int(u.Userid)]; ok {
						delete(lb.uidUsers, u1.Cid)
						u1.Cid = uid
						lb.uidUsers[uid] = u1
						mylog.Infoln("lb user reconneted", u1)
						uu = u1
					} else if u.Userid != 0 {
						uu = &userInfo{
							Cid: uid,
							Account: u.Account,
							UserId: int(u.Userid),
							UserName: u.Name,
							HeadImg: u.Headimg,
							Sex: u.Sex,
							RoomCard: int(u.Roomcard),
							RoomId: int(u.Roomid),
							Coins: int(u.Coins),
						}
						lb.users[uu.UserId] = uu
						lb.uidUsers[uid] = uu
					} else {
						lb.dog.sendClientMessage(uid, proto.CmdUserLogin, &proto.UserLoginRet{
							ErrCode: "logintotry",
							LoginType: req.LoginType,
						})
						return
					}
					lb.dog.sendClientMessage(uid, proto.CmdUserLogin, &proto.UserLoginRet{
						ErrCode: "ok",
						LoginType: req.LoginType,
						User: uu,
					})
					lb.afterLogin(uu)
				} else {
					lb.dog.sendClientMessage(uid, proto.CmdUserLogin, &proto.UserLoginRet{
						ErrCode: "error",
						LoginType: req.LoginType,
					})
				}
			})
		})
	}


	if req.LoginType == "wechat" {
		go func() {
			errCode, acc, token, nickName, headImg, sex = lb.hp.wechatLogin(req.WechatCode)
			mylog.Infoln("wechat userinfo ", acc, token, nickName, headImg, sex)
			if errCode == "ok" {
				lb.reqChn <- &lbRequest{
					req: func() {
						loginHandler()
					},
				}
			} else {
				lb.dog.sendClientMessage(uid, proto.CmdUserLogin, &proto.UserLoginRet{
					LoginType: req.LoginType,
					ErrCode: errCode,
				})
			}
		}()
	} else if req.LoginType == "guest" {
		acc = req.Account
		nickName = "guest_" + acc
		sex = rand.Intn(2) + 1
		headId := rand.Intn(3) + 1
		headImg = strconv.Itoa(headId)
		loginHandler()
	}
}



func (lb *lobby) onUserLogout(uid uint32) {

}


func (lb *lobby) afterLogin(user *userInfo) {
	type noticeList struct {
		NoticeList 	[]*proto.LobbyNotice
	}

	nsl := &noticeList{
		NoticeList: []*proto.LobbyNotice {
			{
				Content: "跑马灯1 bbbbbbbb",
				RepeatCount: 3,
				Duration: 10,
			},
			{
				Content: "跑马灯2 aaaaaaa",
				RepeatCount: -1,
				Duration: 30,
			},
		},
	}

	lb.dog.sendClientMessage(user.Cid, proto.CmdNotice, nsl)
}
