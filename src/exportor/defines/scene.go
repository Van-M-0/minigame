package defines

import "exportor/proto"

type ISceneManager interface {
	SendMessage(uid uint32, message *proto.Message)
	BroadcastMessage(uid []uint32, message *proto.Message)
	RouteMessage(channel string, message *proto.Message)
}
