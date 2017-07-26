package proto

const (
	CmdWechatLogin 		= 1
	CmdUserLogin 		= 2

	CmdUserCreateRoom 	= 3
	CmdUserEnterRoom	= 4
	CmdUserDisRoom 		= 5
	CmdUserAgrDisRoom	= 6


)

type Message struct {
	Len 	uint32
	Cmd 	uint32
	Msg 	[]byte
}

type RegisterServer struct {
	Type 		string
	ServerId 	int
}

type UserWeChatLogin struct {

}

type UserLogin struct {
	Account 	string
}