package proto

/*
	1. 服务器的端口号 9090

	2. 安装mysql, 创建数据库
		CREATE DATABASE IF NOT EXISTS mygame default charset utf8 COLLATE utf8_general_ci;
		账号名 root 密码 1

	3. 所有给服务器发送的数据，json中的key必须是大写, game message(UserMessage)里，所有的子结构（json对象)，必须先字符串化，在发送

	4. 启动服务器

	5. 创建任意账号，服务器自动注册
*/


const (
	CmdWechatLogin 		= 1
	CmdUserLogin 		= 2
	CmdUserLogout 		= 3

	CmdUserCreateRoom 	= 4
	CmdUserEnterRoom	= 5
	CmdUserDisRoom 		= 6
	CmdUserAgrDisRoom	= 7
	CmdLeaveRoom 		= 8

	CmdGetUserInfo 		= 9

	CmdCommonError 		= 10

	CmdUserGameMessage  = 20

		CmdEgUserReady		= 1
		CmdEgGameStart		= 2
		CmdEgUserCallBanker = 3
		CmdEgCallZhu 		= 4
		CmdEgChangeCard		= 5
		CmdEgFindFriend		= 6
		CmdEgOutCard		= 7
		CmdEgGameFinish		= 8
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

type UserLoginRet struct {
	ErrCode 	string
	User 		interface{}
}

type ErguiRoomConf struct {
	A 		int
}

type UserCreateRoom struct {
	Kind 		int				// 1 ergui
	Conf 		[]byte			// 必须是字符串
}

type UserCreateRoomRet struct {
	ErrCode 	string
	RoomInfo 	interface{}
}

type UserEnterRoom struct {
	RoomId 		int
	Test 		string
}

type UserEnterRoomRet struct {
	ErrCode 	string
	Kind 		int			// 1 ergui
	Data 		interface{}
}

type UserMessage struct {
	Cmd 		uint32		// 子命令
	Msg 		interface{}		// 字符串

	/*
		准备
			cmd:1
			msg: null

		叫庄
			cmd:2
			msg: stringfy(ErguiCallBanker)
	*/
}

//和上面对应
type UserMessageRet struct {
	Cmd 		uint32
	Msg 		interface{}
}

type UserCommonError struct {
	Cmd 		uint32
	ErrCode 	string
}

type ErguiGameStart struct {
	Banker 		int
	CardList 	[]int		//游戏开始，发送手牌
}

type ErguiCallBanker struct {
	/*
	BANKER_SCORE_DEFAULT				= 0							//服务器默认值，客户端从下面值开始
	BANKER_SCORE_NO						= 1
	BANKER_SCORE_80						= 2
	BANKER_SCORE_85						= 3
	BANKER_SCORE_90						= 4
	BANKER_SCORE_95						= 5
	BANKER_SCORE_100					= 6
	BANKER_SCORE_100_BAO				= 7
	BANKER_SCORE_100_GOU				= 8
	 */
	Score 		int
}

type ErguiCallbankerRet struct {
	ErrCode 		string		//错误码
	CurBankScore	int			//当前叫庄椅子号
	CurCallSeat 	int			//当前叫庄分数
	ToCallSeat 		int			//下一个叫庄以自豪, -1标识结束
}

/*
	card = []byte {
		0x03,0x04,0x05,0x06,0x07,0x08,0x09,0x0A,0x0B,0x0C,0x0D,0x0E,0x0F,		//方块 3 - 2
		0x13,0x14,0x15,0x16,0x17,0x18,0x19,0x1A,0x1B,0x1C,0x1D,0x1E,0x1F,		//梅花 3 - 2
		0x23,0x24,0x25,0x26,0x27,0x28,0x29,0x2A,0x2B,0x2C,0x2D,0x2E,0x2F,		//红桃 3 - 2
		0x33,0x34,0x35,0x36,0x37,0x38,0x39,0x3A,0x3B,0x3C,0x3D,0x3E,0x3F,		//黑桃 3 - 2
		0x43,0x44,
	}
 */
type ErguiCallZhu struct {
	Color 			int			//猪排颜色 0, 1, 2, 3, 4
}

type ErguiCallZhuRet struct {
	ErrCode 		string
	Color 			int			//主牌颜色
	BottomSeat 		int			//交换底牌的椅子号
	BottomCard		[]int		//底牌数据
}

type ErguiChangeCard struct {
	BottomCard		[]int		//交换的拍数据
}

type ErguiChangeCardRet struct {
	ErrCode 		string
}

type ErguiFindFriend struct {
	Card 			int		//朋友牌
}

type ErguiFindFriendRet struct {
	ErrCode 		string
	Card 			int		//朋友牌
}

type ErguiUesrOutCard struct {
	Card 			int		//出牌
}

type ErguiUserOutCardRet struct {
	ErrCode 		string
	Card 			int		//出的牌
	OutSeat			int			//出牌位置
	NextSeat 		int			//下个出牌玩家
	NewRound 		bool		//新一轮
	FirstSeat		int			//每轮第一次出牌的玩家
}

type ErguiGameFinish struct {

}