package defines

type PlayerInfo struct {
	Uid 		uint32
	UserId 		uint32
	OpenId 		string
	HeadImg 	string
	Name 		string
	Account		string
	Diamond 	int
	Gold 		int64
	RoomCard 	int
	Sex 		byte

	RoomId 		uint32
}

type CreateRoomConf struct {
	RoomId 		uint32
	Conf 		[]byte
}

type IGameManager interface {
	SendUserMessage(info *PlayerInfo, cmd uint32, data interface{})
	BroadcastMessage(cmd uint32, data interface{})
	SetTimer(id uint32, data interface{}) error
	KillTimer(id uint32) error
}

type IGame interface {
	OnInit(manager IGameManager, Option interface{}) error
	OnRelease()
	OnGameCreate(info *PlayerInfo, conf *CreateRoomConf) error
	OnUserEnter(info *PlayerInfo) error
	OnUserLeave(info *PlayerInfo)
	OnUserOffline(info *PlayerInfo)
	OnUserMessage(info *PlayerInfo, cmd uint32, data []byte) error
	OnTimer(id uint32, data interface{})
}
