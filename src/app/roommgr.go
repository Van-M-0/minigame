package main

import (
	"exportor/proto"
	"msgpacker"
	"mylog"
	"math/rand"
)

type roomMgr struct {
	lb 		*lobby
	rooms 	map[int]*erguiRoom
}

func newRoomMgr(lb *lobby) *roomMgr {
	mgr := &roomMgr{}
	mgr.rooms = make(map[int]*erguiRoom)
	mgr.lb = lb
	return mgr
}

func (mgr *roomMgr) onMessage(user *userInfo, cmd uint32, data []byte) {
	if cmd == proto.CmdUserCreateRoom {
		mgr.onUserCreateRoom(user, data)
	} else if cmd == proto.CmdUserEnterRoom {
		mgr.onUserEnterRoom(user, data)
	} else if cmd == proto.CmdLeaveRoom {
		mgr.onUserLeaveRoom(user, data)
	} else if cmd == proto.CmdUserGameMessage {
		mgr.onUserGameMessage(user, data)
	}
}


func (mgr *roomMgr) genRoomId() int {
	for i := 0; i < 20; i++ {
		d := rand.Intn(899999) + 100000
		if _, ok := mgr.rooms[d]; ok {
			continue
		}
		return d
	}
	return -1
}

func (mgr *roomMgr) onUserCreateRoom(user *userInfo, data []byte) {
	var req proto.UserCreateRoom
	msgpacker.UnMarshal(data, &req)

	if user.RoomId != 0 {
		mgr.lb.dog.sendClientMessage(user.Cid ,proto.CmdUserCreateRoom, &proto.UserCreateRoomRet{
			ErrCode: "haveroom",
		})
		return
	}

	var room *erguiRoom
	if req.Kind == 1 {
		room = newErguiRoom(mgr)
	}
	if room == nil {
		mgr.lb.dog.sendClientMessage(user.Cid ,proto.CmdUserCreateRoom, &proto.UserCreateRoomRet{
			ErrCode: "kind error",
		})
		return
	}

	id := mgr.genRoomId()
	if id == -1 {
		mgr.lb.dog.sendClientMessage(user.Cid ,proto.CmdUserCreateRoom, &proto.UserCreateRoomRet{
			ErrCode: "roomid err",
		})
	}

	ret := room.CreateRoom(id, user, &req)
	if ret != "ok" {
		mgr.lb.dog.sendClientMessage(user.Cid ,proto.CmdUserCreateRoom, &proto.UserCreateRoomRet{
			ErrCode: ret,
		})
	} else {
		mgr.rooms[id] = room
	}
}

func (mgr *roomMgr) releaseRoom(id int) {
	if r, ok := mgr.rooms[id]; ok {
		delete(mgr.rooms, id)
		mylog.Infoln("room destroy ", id)
		r.reqCh <- &erguiRoomReq{
			cmd: "js",
		}
	} else {
		mylog.Infoln("release room error ", id)
	}
}

func (mgr *roomMgr) onUserEnterRoom(user *userInfo, data []byte) {
	mylog.Infoln("user enter room", user)

	var req proto.UserEnterRoom
	msgpacker.UnMarshal(data, &req)

	if user.RoomId != 0 {
		if _, ok := mgr.rooms[req.RoomId]; ok {
			mgr.onUserReconnect(user)
		} else {
			user.RoomId = 0
			mgr.sendClientMessage(user, proto.CmdUserEnterRoom, &proto.UserEnterRoomRet{
				ErrCode: "roomNotExistsReEnter",
			})
		}
		return
	}

	if room, ok := mgr.rooms[req.RoomId]; ok {
		room.reqCh <- &erguiRoomReq {
			cmd: "e",
			user: user,
			data: data,
		}
	} else {
		mgr.sendClientMessage(user, proto.CmdUserEnterRoom, &proto.UserEnterRoomRet{
			ErrCode: "roomNotExists",
		})
	}
}

func (mgr *roomMgr) onUserDisRoom(user *userInfo, data []byte) {

}

func (mgr *roomMgr) onUserAgreeDisRoom(user *userInfo, data []byte) {

}

func (mgr *roomMgr) onUserGameMessage(user *userInfo, data []byte) {
	if room, ok := mgr.rooms[user.RoomId]; ok {
		room.reqCh <- &erguiRoomReq {
			cmd: "m",
			user: user,
			data: data,
		}
	} else {
		mylog.Infoln("user message not in room", user, data)
	}

}

func (mgr *roomMgr) onUserLeaveRoom(user *userInfo, data []byte) {

}

func (mgr *roomMgr) onUserOffline(user *userInfo) {
	mylog.Infoln("room mgr offline", user, mgr.rooms)
	if room, ok := mgr.rooms[user.RoomId]; ok {
		room.reqCh <- &erguiRoomReq{
			cmd: "o",
			user: user,
		}
	}
}

func (mgr *roomMgr) onUserReconnect(user *userInfo) {
	mylog.Infoln("room mgr reconnect", user, mgr.rooms)
	if room, ok := mgr.rooms[user.RoomId]; ok {
		room.reqCh <- &erguiRoomReq{
			cmd: "r",
			user: user,
		}
	}
}

func (mgr *roomMgr) onUserOfflineTimeout(user *userInfo) {
	mylog.Infoln("room mgr offline timeout", user, mgr.rooms)
	if room, ok := mgr.rooms[user.RoomId]; ok {
		room.reqCh <- &erguiRoomReq{
			cmd: "ot",
			user: user,
		}
	}
}

func (mgr *roomMgr) sendClientMessage(user *userInfo, cmd uint32, data interface{}) {
	mgr.lb.dog.sendClientMessage(user.Cid, cmd, data)
}

func (mgr *roomMgr) bcClientMessage(user []*userInfo, cmd uint32, data interface{}) {

}

func (mgr *roomMgr) onDebug(room int, w, r interface {}) {
	if room, ok := mgr.rooms[room]; ok {
		room.reqCh <- &erguiRoomReq{
			cmd: "debug",
			data: w,
		}
	} else {
		mylog.Infoln("debug room not exists ", room)
	}
}