package main

import (
	"exportor/proto"
	"fmt"
	"msgpacker"
	"runtime/debug"
	"math/rand"
	"time"
)

const (
	EgKind 	 int = 1
)

type erguiRoomReq struct {
	cmd 		string
	user 		*userInfo
	req 		uint32
	data 		interface{}
}

type erguiRoom struct {
	id 			int
	creator 	int
	mgr 		*roomMgr
	reqCh 		chan *erguiRoomReq
	handler 	*egHandler
}

func newErguiRoom(mgr *roomMgr) *erguiRoom {
	eg := &erguiRoom{}
	eg.mgr = mgr
	eg.reqCh = make(chan *erguiRoomReq, 512)
	eg.handler = newEgHandler(eg)
	return eg
}

func (eg *erguiRoom) UserOffline(user *userInfo) {
	fmt.Println("user offline", user)
}

func (eg *erguiRoom) UserReconnect(user *userInfo) {
	fmt.Println("user reconnect ", user)
	eg.handler.reenter(user)
}

func (eg *erguiRoom) UserOfflineTimeout(user *userInfo) {
	fmt.Println("user offline timeout ", user)
}

func (eg *erguiRoom) CreateRoom(id int, user *userInfo, req *proto.UserCreateRoom) string {
	//var c proto.ErguiRoomConf
	//msgpacker.UnMarshal(conf, &c)
	//c := req.Conf.(*proto.ErguiRoomConf)
	ret := eg.handler.createRoom(id, user, req)
	if ret == "ok" {
		eg.run()
		eg.id = id
		eg.creator = user.UserId
	}
	return ret
}

func (eg *erguiRoom) OnReleaseRoom()  {
	eg.handler.onReleaseRoom()
}

func (eg *erguiRoom) EnterRoom(u *userInfo) {
	eg.handler.userEnterRoom(u)
}

func (eg *erguiRoom) GameMessage(user *userInfo, data interface{}) {
	fmt.Println("xxx game message ", data)

	var message proto.UserMessage
	if message.Cmd == proto.CmdEgUserReady {

	} else if message.Cmd == proto.CmdEgUserCallBanker {
		message.Msg = proto.ErguiCallBanker{}
	} else if message.Cmd == proto.CmdEgCallZhu {
		message.Msg = proto.ErguiCallZhu{}
	} else if message.Cmd == proto.CmdEgChangeCard {
		message.Msg = proto.ErguiChangeCard{}
	} else if message.Cmd == proto.CmdEgFindFriend {
		message.Msg = proto.ErguiFindFriend{}
	} else if message.Cmd == proto.CmdEgOutCard {
		message.Msg = proto.ErguiUesrOutCard{}
	}

	if err := msgpacker.UnMarshal(data.([]byte), &message); err != nil {
		fmt.Println("unmarshal game messsage error ", err, data)
		return
	}

	fmt.Println("game message ", message.Cmd, message.Msg)

	if message.Cmd == proto.CmdEgUserReady {
		eg.handler.userReady(user, message.Msg)
	} else if message.Cmd == proto.CmdEgUserCallBanker {
		eg.handler.userCallBanker(user, message.Msg)
	} else if message.Cmd == proto.CmdEgCallZhu {
		eg.handler.onUserCallZhu(user, message.Msg)
	} else if message.Cmd == proto.CmdEgChangeCard {
		eg.handler.onUserChangeCard(user, message.Msg)
	} else if message.Cmd == proto.CmdEgFindFriend {
		eg.handler.onUserFindFriend(user, message.Msg)
	} else if message.Cmd == proto.CmdEgOutCard {
		eg.handler.onUserOutCard(user, message.Msg)
	}
}

func (eg *erguiRoom) sfCall() bool {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("--------------- error stack ----------------")
			fmt.Println(err)
			debug.PrintStack()
			fmt.Println("--------------------------------------------")
		}
	}()
	select {
	case r := <- eg.reqCh:
		fmt.Println("run process ", r)
		if r.cmd == "r"	{
			eg.UserReconnect(r.user)
		} else if r.cmd == "o" {
			eg.UserOffline(r.user)
		} else if r.cmd == "ot" {
			eg.UserOfflineTimeout(r.user)
		} else if r.cmd == "m" {
			eg.GameMessage(r.user, r.data)
		} else if r.cmd == "e" {
			eg.EnterRoom(r.user)
		} else if r.cmd == "t" {
			eg.handler.onTimer(r.req, r.data.(func()))
		} else if r.cmd == "js" {
			eg.OnReleaseRoom()
			return true
		}
	}
	return false
}

func (eg *erguiRoom) run() {

	go func() {
		for {
			if eg.sfCall() {
				return
			}
		}
	}()

	fmt.Println("ergui room run")
}

// ----------------------------------------------------------------
//								card
// ----------------------------------------------------------------
var (
	cardBox = []int {
		0x03,0x04,0x05,0x06,0x07,0x08,0x09,0x0A,0x0B,0x0C,0x0D,0x0E,0x0F,		//方块 3 - 2
		0x13,0x14,0x15,0x16,0x17,0x18,0x19,0x1A,0x1B,0x1C,0x1D,0x1E,0x1F,		//梅花 3 - 2
		0x23,0x24,0x25,0x26,0x27,0x28,0x29,0x2A,0x2B,0x2C,0x2D,0x2E,0x2F,		//红桃 3 - 2
		0x33,0x34,0x35,0x36,0x37,0x38,0x39,0x3A,0x3B,0x3C,0x3D,0x3E,0x3F,		//黑桃 3 - 2
		0x43,0x44,
	}

	friendCardBox = []int {
		0, 	0, 	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0x0D,0x0E,0x0F,		//方块 3 - 2
		0, 	0, 	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0x1D,0x1E,0x1F,		//梅花 3 - 2
		0, 	0, 	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0x2D,0x2E,0x2F,		//红桃 3 - 2
		0, 	0, 	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0x3D,0x3E,0x3F,		//黑桃 3 - 2
		0, 	0, 	0,	0x43,	0x44,												//小鬼，大鬼
	}

	zhuCardBox = []int {
		0, 	0, 	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0x0F,		//方块 3 - 2
		0, 	0, 	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0x1F,		//梅花 3 - 2
		0, 	0, 	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0x2F,		//红桃 3 - 2
		0, 	0, 	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0x3F,		//黑桃 3 - 2
		0, 	0, 	0,	0x43,	0x44,												//小鬼，大鬼
	}

	scoreCard = []int { 0, 0, 0, 0, 0, 5, 0, 0, 0, 0, 10, 0, 0, 10, 0, 0, }

	MaxCardCount = 54
	MaxCardIndex = 0x45
	MaxPlayerCardCount = 12
	MaxBottomCard = 6

	cardMin = 0x03
	cardMax = 0x44
)

type _ErguiCfg struct {
	gameType 		[]int 		//1 经典 2 疯子
	money 			[]int
	round 			[]int
	consumeRoomCard	int
}

var _config = _ErguiCfg {
	gameType: []int{1, 2},
	money: []int{10, 20, 50, 100, 500},
	round: []int{1, 4, 8, 16, 32},
	consumeRoomCard : 0,
}


const (
	BANKER_SCORE_DEFAULT				= 0							//服务器默认值，客户端从下面值开始
	BANKER_SCORE_NO						= 1
	BANKER_SCORE_80						= 2
	BANKER_SCORE_85						= 3
	BANKER_SCORE_90						= 4
	BANKER_SCORE_95						= 5
	BANKER_SCORE_100					= 6
	BANKER_SCORE_100_BAO				= 7
	BANKER_SCORE_100_GOU				= 8
)

func randCards(cards []int) {
	rand.Seed(time.Now().Unix())
	for i := 0; i < MaxCardCount; i++ {
		pos := rand.Intn(MaxCardCount-i) + i
		cards[i], cards[pos] = cards[pos], cards[i]
	}
}

// ----------------------------------------------------------------
//								logic
// ----------------------------------------------------------------


const (
	EgMaxSeat 		int = 4
)

type playerInfo struct {
	*userInfo
	ready 		bool
	seat 		int
	handCard	[]int
	indexs 		[]int
	callScore 	int
}

type egHandler struct {
	room 		*erguiRoom
	t 			*time.Timer
	stoped 		bool
	tid 		uint32
	status 		string
	gameRound	int

	conf 		*proto.UserCreateRoom
	creator 	int
	playerList 	map[int]*playerInfo
	seatList 	[EgMaxSeat]*playerInfo
	cards 		[]int
	bottomCard	[]int

	isCalled 	bool

	banker 		int
	curBanker 	int
	firstCall	int
	//callRound 	int
	canCallGou 	bool
	bankerScore int

	zhuColor 	int

	friend		int
	friendSeat	int

	firstSeat 	int
	firstCard 	int
	curseat 	int
	leftRoud 	int

	outcardList []int

	xiqian 		[]float32
	scoreList 	[]int
	totalWinMoney []float32
}

func newEgHandler(r *erguiRoom) *egHandler {
	handler := &egHandler{}
	handler.room = r
	handler.playerList = make(map[int]*playerInfo)
	handler.cards = make([]int, MaxCardCount)
	handler.banker = -1
	handler.status = "ready"
	handler.outcardList = make([]int, EgMaxSeat)
	handler.xiqian = make([]float32, EgMaxSeat)
	handler.scoreList = make([]int, EgMaxSeat)
	handler.totalWinMoney = make([]float32, EgMaxSeat)
	handler.leftRoud = MaxPlayerCardCount
	handler.gameRound = 1
	return handler
}

func (h *egHandler) createRoom(roomId int, user *userInfo, req *proto.UserCreateRoom) string {

	conf := req

	fmt.Println("recv client conf ", conf)

	exists := func(t int, arr []int) bool {
		have := false
		for _, m := range arr {
			if m == t {
				have = true
				break
			}
		}
		fmt.Println(have)
		return have
	}

	if !exists(conf.GameType, _config.gameType) || !exists(conf.Score, _config.money) || !exists(conf.Round, _config.round) {
		h.room.mgr.lb.dog.sendClientMessage(user.Cid ,proto.CmdUserCreateRoom, &proto.UserCreateRoomRet{
			ErrCode: "conferr",
		})
		return "error"
	}

	h.creator = user.UserId
	h.conf = conf
	fmt.Println("create room conf")

	type rinfo struct {
		RoomId 		int
	}
	h.sendMessage(&playerInfo{userInfo:user}, proto.CmdUserCreateRoom, &proto.UserCreateRoomRet {
		ErrCode: "ok",
		RoomInfo: rinfo{
			RoomId: roomId,
		},
	})
	/*
	h.playerList[user.UserId] = &playerInfo{
		userInfo: user,
		seat: -1,
	}
	*/
	return "ok"
}

func (h *egHandler) onReleaseRoom() {
	for _, u := range h.playerList {
		u.RoomId = 0
	}
}

func (h *egHandler) userEnterRoom(user *userInfo) {

	op, entered := h.playerList[user.UserId]
	if entered == false {
		p := &playerInfo{
			userInfo: user,
			seat: - 1,
			ready: false,
		}

		p.seat = h.setSeat(p)
		if p.seat == -1 {
			h.sendMessage(p, proto.CmdUserEnterRoom, &proto.UserEnterRoomRet{
				ErrCode: "full",
				Kind: EgKind,
			})
			return
		}
		h.seatList[p.seat] = p
		h.playerList[p.UserId] = p

		p.userInfo.RoomId = h.room.id
	} else {
		op.userInfo = user
	}


	type resEnterUser struct {
		*userInfo
		Seat 		int
		Ready 		bool
		ReEnter 	bool
	}

	type egUserEnter struct {
		Creator 	int
		Conf 		*proto.ErguiRoomConf
		Player 		[]*resEnterUser
	}

	var enterRes egUserEnter
	enterRes.Creator = h.room.creator
	enterRes.Conf = &proto.ErguiRoomConf{
		GameType: h.conf.GameType,
		Round: h.conf.Round,
		Score: h.conf.Score,
	}
	for _, player := range h.playerList {
		enterRes.Player = append(enterRes.Player, &resEnterUser{
			userInfo: player.userInfo,
			Seat: player.seat,
			Ready: player.ready,
			ReEnter: entered,
		})
	}

	fmt.Println("enter room ", enterRes)
	h.bcMessage(proto.CmdUserEnterRoom, &proto.UserEnterRoomRet{
		ErrCode: "ok",
		Kind: EgKind,
		Data: enterRes,
	})
}

func (h *egHandler) reenter(user *userInfo) {

	if p, ok := h.playerList[user.UserId]; !ok {
		fmt.Println("user reenter not exists ??? ", user, p)
		return
	} else {
		p.Cid = user.Cid
	}

	type reenterPlayer struct {
		*userInfo
		Ready 		bool
		Seat 		int
		HandCard	[]int
		CallScore 	int
	}

	type reenterRes struct {
		Status 		string
		Creator 	int
		Conf 		*proto.ErguiRoomConf
		GameRound 	int
		RenterUser 	int
		Players 	map[int]*reenterPlayer

		BottomCard	[]int

		Banker 		int
		CurBanker 	int
		FirstCall	int
		BankerScore int

		ZhuColor 	int
		Friend		int

		FirstSeat 	int
		FirstCard 	int
		Curseat 	int

		OutcardList []int

		ScoreList 	[]int
		XiQianMoney	[]float32
	}

	ret := &reenterRes{
		Players: make(map[int]*reenterPlayer),
	}
	for _, p := range h.playerList {
		rp := &reenterPlayer{
			userInfo: p.userInfo,
			Ready: p.ready,
			Seat: p.seat,
			HandCard: p.handCard,
			CallScore: p.callScore,
		}
		ret.Players[p.UserId] = rp
	}

	ret.Status = h.status
	ret.Creator = h.creator
	ret.Conf = &proto.ErguiRoomConf{
		GameType: h.conf.GameType,
		Round: h.conf.Round,
		Score: h.conf.Score,
	}
	ret.GameRound = h.gameRound
	ret.RenterUser = user.UserId
	ret.BottomCard = h.bottomCard
	ret.Banker = h.banker
	ret.CurBanker = h.curBanker
	ret.FirstCall = h.firstCall
	ret.BankerScore = h.bankerScore
	ret.ZhuColor = h.zhuColor
	ret.Friend = h.friend
	ret.FirstSeat = h.firstSeat
	ret.FirstCard = h.firstCard
	ret.Curseat = h.curseat
	ret.OutcardList = h.outcardList
	ret.ScoreList = h.scoreList
	ret.XiQianMoney = h.xiqian

	ret1 := &reenterRes{
		Status: h.status,
		RenterUser: user.UserId,
	}

	for i := 0; i < EgMaxSeat; i++ {
		p := h.playerList[i]
		if p == nil {
			continue
		}
		if p.UserId == user.UserId {
			h.sendGameMessage(p, proto.CmdEgReEnter, ret)
		} else {
			h.sendGameMessage(p, proto.CmdEgReEnter, ret1)
		}
	}
}

func (h *egHandler) userReady(user *userInfo, msg interface{}) {

	fmt.Println("user ready data ", msg)

	type readyRes struct {
		ErrCode 	string
		Seat 		int
		Ready 		bool
	}

	if user.RoomCard < _config.consumeRoomCard {
		p := &playerInfo{
			userInfo: user,
		}
		h.sendGameMessage(p, proto.CmdEgUserReady, &readyRes{
			ErrCode: "roomcard",
		})
		return
	}

	p := h.playerList[user.UserId]
	if p.ready {
		return
	}
	p.ready = true


	res := &readyRes{
		ErrCode: "ok",
		Seat: p.seat,
		Ready: true,
	}

	h.bcMessage(proto.CmdUserGameMessage, &proto.UserMessageRet{
		Cmd: proto.CmdEgUserReady,
		Msg: res,
	})

	cnt := 0
	for _, p := range h.playerList {
		if !p.ready {
			return
		} else {
			cnt++
		}
	}
	if cnt != EgMaxSeat {
		return
	}

	h.startGame()
}

func (h *egHandler) userCallBanker(user *userInfo, msg interface{}) {
	//var req proto.ErguiCallBanker
	//req := msg.(proto.ErguiCallBanker)
	/*
	if err := msgpacker.UnMarshal(msg, &req); err != nil {
		fmt.Println("requet error: callbanker", err)
		return
	}
	*/
	req := msg.(map[string]interface{})
	h.callScore(user, int(req["Score"].(float64)))
}

func (h *egHandler) getDijin() float32 {
	return float32(h.conf.Score) / 100 * 2
}

func (h *egHandler) calXiqian(score int, seat int) {
	xq := 0
	if score == BANKER_SCORE_100_BAO {
		if h.conf.GameType == 1 {
			xq = 10
		} else if h.conf.GameType == 2 {
			xq = 20
		}
	} else if score == BANKER_SCORE_100_GOU {
		if h.conf.GameType == 1 {
			xq = 15
		} else if h.conf.GameType == 2 {
			xq = 25
		}
	}
	for i := 0; i < EgMaxSeat; i++ {
		if i != seat {
			h.xiqian[i] = -1 * h.getDijin() * float32(xq)
			h.xiqian[seat] += h.getDijin() * float32(xq)
		}
	}
}

func (h *egHandler) callScore(user *userInfo, score int) {

	p := h.playerList[user.UserId]
	if h.curBanker != p.seat {
		h.sendGameMessage(p, proto.CmdEgUserCallBanker, &proto.ErguiCallbankerRet{
			ErrCode: "notyou",
		})
		return
	}

	if h.status != "call" {
		fmt.Println(">>>>>>>>user status not in call")
		h.sendGameMessage(p, proto.CmdEgUserCallBanker, &proto.ErguiCallbankerRet{
			ErrCode: "notstatus",
		})
		return
	}

	//第一轮不能喊钩
	if h.canCallGou == false && score == BANKER_SCORE_100_GOU {
		h.sendGameMessage(p, proto.CmdEgUserCallBanker, &proto.ErguiCallbankerRet{
			ErrCode: "firstgou",
		})
		return
	}

	//第一个玩家不能喊过
	if h.isCalled == false && score == BANKER_SCORE_NO {
		h.sendGameMessage(p, proto.CmdEgUserCallBanker, &proto.ErguiCallbankerRet{
			ErrCode: "firstno",
		})
		return
	}
	h.isCalled = true

	// 没有喊过，必须比当前所有喊分的大
	if score != BANKER_SCORE_NO {
		for _, p := range h.playerList {
			if p.callScore != 0 && score <= p.callScore {
				h.sendGameMessage(p, proto.CmdEgUserCallBanker, &proto.ErguiCallbankerRet{
					ErrCode: "small",
				})
				return
			}
		}
	}

	if score == BANKER_SCORE_100_BAO {
		h.calXiqian(score, h.curBanker)
		h.canCallGou = true
	}

	h.killTimer()
	p.callScore = score

	if score == BANKER_SCORE_100_GOU {
		h.calXiqian(score, h.curBanker)
		h.banker = h.curBanker
		h.bankerScore = score
	} else {
		nextBanker := (h.curBanker + 1) % EgMaxSeat
		if h.firstCall == nextBanker {
			//h.callRound++
		}
		nextp := h.seatList[nextBanker]

		//没有叫过分
		if nextp.callScore == BANKER_SCORE_DEFAULT {
			h.curBanker = nextBanker
			h.bcGameMessage(proto.CmdEgUserCallBanker, &proto.ErguiCallbankerRet{
				ErrCode: "ok",
				CurBankScore: score,
				CurCallSeat: p.seat,
				ToCallSeat: h.curBanker,
			})
			h.setTimer("callbankerTimeout", 10, func() {
				h.callScore(nextp.userInfo, BANKER_SCORE_NO)
			})
			return
		} else {
			count := 0
			for _, p := range h.playerList {
				if p.callScore != BANKER_SCORE_NO {
					count++
				}
			}

			if count == 1 {
				h.banker = p.seat
				h.bankerScore = p.callScore
			} else {

				for i := 0; i < EgMaxSeat; i++ {
					if nextp.callScore == BANKER_SCORE_NO {
						nextBanker = (nextBanker + 1) % EgMaxSeat
						nextp = h.seatList[nextBanker]
						continue
					}

					h.curBanker = nextBanker
					h.bcGameMessage(proto.CmdEgUserCallBanker, &proto.ErguiCallbankerRet{
						ErrCode: "ok",
						CurBankScore: score,
						CurCallSeat: p.seat,
						ToCallSeat: h.curBanker,
					})
					break
				}
				h.setTimer("callbankerTimeout", 10, func() {
					h.callScore(nextp.userInfo, BANKER_SCORE_NO)
				})
				return
			}
		}
	}

	h.bcGameMessage(proto.CmdEgUserCallBanker, &proto.ErguiCallbankerRet{
		ErrCode: "ok",
		CurBankScore: score,
		CurCallSeat: p.seat,
		ToCallSeat: -1,
	})

	h.status = "zhu"
	h.setTimer("callzhutimeout", 10, func() {
		h.callzhu(h.playerList[h.banker].userInfo, h.getRecommendZhu())
	})
}

func (h *egHandler) getRecommendZhu() int {
	return rand.Intn(4)
}

func (h *egHandler) onUserCallZhu(user *userInfo, msg interface{}) {
	/*
	var req proto.ErguiCallZhu
	if err := msgpacker.UnMarshal(msg, &req); err != nil {
		fmt.Println("requet error: call zhu", err)
		return
	}
	*/
	req := msg.(map[string]interface{})
	h.callzhu(user, int(req["Color"].(float64)))
}

func (h *egHandler) callzhu(user *userInfo, color int) {

	p := h.playerList[user.UserId]
	if p.seat != h.banker {
		h.sendGameMessage(p, proto.CmdEgCallZhu, &proto.ErguiCallZhuRet{
			ErrCode: "notyou",
		})
		return
	}

	if h.status != "zhu" {
		fmt.Println(">>>>>>>>user status not in zhu")
		h.sendGameMessage(p, proto.CmdEgCallZhu, &proto.ErguiCallZhuRet{
			ErrCode: "notstatus",
		})
		return
	}

	if p.seat != h.banker {
		h.sendGameMessage(p, proto.CmdEgCallZhu, &proto.ErguiCallZhuRet{
			ErrCode: "zhuerror",
		})
		return
	}

	h.killTimer()
	h.zhuColor = color
	fmt.Println("zhu color ", color)

	for _, p := range h.playerList {
		if p.seat == h.banker {
			h.sendGameMessage(p, proto.CmdEgCallZhu, &proto.ErguiCallZhuRet{
				ErrCode: "ok",
				Color: color,
				BottomSeat: h.banker,
				BottomCard: h.bottomCard,
			})
		} else {
			h.sendGameMessage(p, proto.CmdEgCallZhu, &proto.ErguiCallZhuRet{
				ErrCode: "ok",
				Color: color,
				BottomSeat: h.banker,
			})
		}
	}

	for i := 0; i < MaxBottomCard; i++ {
		p.indexs[h.bottomCard[i]]++
	}

	h.status = "change"
	h.setTimer("changetimeout", 10, func() {
		h.changeCard(h.playerList[h.banker].userInfo, h.getRecommendCard())
	})
}

func (h *egHandler) onUserChangeCard(user *userInfo, msg interface{}) {
	/*
	var req proto.ErguiChangeCard
	if err := msgpacker.UnMarshal(msg, &req); err != nil {
		fmt.Println("requet error: change card", err)
		return
	}
	*/
	fmt.Println("user change card ", msg)
	req := msg.(map[string]interface{})
	card := req["BottomCard"].([]interface{})
	icard := []int{}
	for _, c := range card {
		icard = append(icard, int(c.(float64)))
	}
	h.changeCard(user, icard)
}

func (h *egHandler) getRecommendCard() []int {
	return []int{}
}

func (h *egHandler) changeCard(user *userInfo, data []int) {
	fmt.Println("user change card ", data)

	p := h.playerList[user.UserId]
	if p.seat != h.banker {
		h.sendGameMessage(p, proto.CmdEgChangeCard, &proto.ErguiChangeCardRet{
			ErrCode: "notyou",
		})
		return
	}

	if h.status != "change" {
		fmt.Println(">>>>>>>>user status not in change")
		h.sendGameMessage(p, proto.CmdEgChangeCard, &proto.ErguiChangeCardRet{
			ErrCode: "notstatus",
		})
		return
	}

	if len(data) != MaxBottomCard {
		h.sendGameMessage(p, proto.CmdEgChangeCard, &proto.ErguiChangeCardRet{
			ErrCode: "count",
		})
		return
	}

	for i := 0; i < MaxBottomCard; i++ {
		card := data[i]
		if p.indexs[card] == 0 {
			h.sendGameMessage(p, proto.CmdEgChangeCard, &proto.ErguiChangeCardRet{
				ErrCode: "cdata",
			})
			return
		}

		val := card & 0x0F
		if val == 0x05 || val == 0x0A || val == 0x0D {
			h.sendGameMessage(p, proto.CmdEgChangeCard, &proto.ErguiChangeCardRet{
				ErrCode: "notallow",
			})
			return
		}
	}

	h.killTimer()
	for i := 0; i < MaxBottomCard; i++ {
		p.indexs[data[i]]--
		if p.indexs[data[i]] < 0 {
			fmt.Println("change card index < 0")
		}
	}

	h.setTimer("friendtimeout", 10, func() {
		h.findFriend(p.userInfo, h.getRecommendFriend())
	})

	h.bcGameMessage(proto.CmdEgChangeCard, &proto.ErguiChangeCardRet{
		ErrCode: "ok",
		BottomCard: data,
	})

	h.status = "friend"
}

func (h *egHandler) onUserFindFriend(user *userInfo, msg interface{}) {
	/*
	var req proto.ErguiFindFriend
	if err := msgpacker.UnMarshal(msg, &req); err != nil {
		fmt.Println("requet error: findFriend", err)
		return
	}
	*/
	req := msg.(map[string]interface{})
	h.findFriend(user, int(req["Card"].(float64)))
}

func (h *egHandler) getRecommendFriend() int {
	return cardBox[rand.Intn(MaxCardCount)]
}

func (h *egHandler) findFriend(user *userInfo, card int) {

	p := h.playerList[user.UserId]
	if p.seat != h.banker {
		h.sendGameMessage(p, proto.CmdEgFindFriend, &proto.ErguiFindFriendRet{
			ErrCode: "notyou",
		})
		return
	}

	if h.status != "friend" {
		h.sendGameMessage(p, proto.CmdEgFindFriend, &proto.ErguiFindFriendRet{
			ErrCode: "notstatus",
		})
		fmt.Println(">>>>>>>>user status not in change", h.status)
		return
	}

	if int(card) > len(friendCardBox) || friendCardBox[card] == 0 {
		h.sendGameMessage(p, proto.CmdEgFindFriend, &proto.ErguiFindFriendRet{
			ErrCode: "notallow",
		})
		return
	}

	h.friend = card
	h.curseat = p.seat
	h.firstSeat = p.seat
	h.firstCard = -1

	h.bcGameMessage(proto.CmdEgFindFriend, &proto.ErguiFindFriendRet{
		ErrCode: "ok",
		Card: card,
	})

	h.setTimer("outcardtimeout", 10, func() {
		h.outCard(p.userInfo, h.getRecommendOutcard())
	})
	h.status = "outcard"
}

func (h *egHandler) onUserOutCard(user *userInfo, msg interface{}) {
	/*
	var req proto.ErguiUesrOutCard
	if err := msgpacker.UnMarshal(msg, &req); err != nil {
		fmt.Println("requet error: out card", err)
		return
	}
	*/

	req := msg.(map[string]interface{})
	h.outCard(user, int(req["Card"].(float64)))
}

func (h *egHandler) getRecommendOutcard() int {
	return 0
}

func (h *egHandler) checkOutCard(seat int, card int) string {
	firstCard := h.outcardList[h.firstSeat]
	firstColor := (firstCard & 0xF0) >> 4

	zcard := zhuCardBox[firstCard] != 0
	zcolor := firstColor == h.zhuColor

	fmt.Println("checkout ", h.outcardList, firstCard, zcard, zcolor, seat)

	for i, p := range h.seatList {
		fmt.Println("index ", i)
		for m := 0; m < MaxCardIndex; m++ {
			if p.indexs[m] != 0 {
				fmt.Print(m, p.indexs[m], " , ")
			}
			fmt.Println("")
		}
	}

	ccolor := (card & 0xF0) >> 4

	haveZhuCard := func() bool {
		p := h.seatList[seat]
		for i := 0; i < MaxCardIndex; i++ {
			if p.indexs[i] != 0  && zhuCardBox[i] != 0 {
				return true
			}
		}
		return false
	}

	haveZhuColor := func() bool {
		zmin := h.zhuColor << 4 + 0x03
		zmax := h.zhuColor << 4 + 0x0F
		p := h.seatList[seat]
		for i := zmin; i < zmax; i++ {
			if p.indexs[i] != 0 {
				return true
			}
		}
		return false
	}

	if zcard || zcolor {
		if zhuCardBox[card] != 0 {
			return "ok"
		}

		if ccolor == h.zhuColor {
			return "ok"
		}

		if haveZhuCard() {
			return "youzhu-card"
		}

		if haveZhuColor() {
			return "youzhu-color"
		}

	} else {
		if zhuCardBox[card] != 0 {
			return "ok"
		}
		if ccolor == firstColor {
			return "ok"
		}

		fmin := firstColor << 4 + 0x03
		fmax := firstColor << 4 + 0x0F
		p := h.seatList[seat]
		for i := fmin; i < fmax; i++ {
			val := int(i & 0x0F)
			if p.indexs[i] != 0 && val != 2{
				return "you-fu"
			}
		}

		/*
		if haveZhuCard() {
			return "youzhu-fu"
		}
		if haveZhuColor() {
			return "youzhu-fucolor"
		}
		*/
	}

	return "ok"
}

func (h *egHandler) getMaxScoreSeat() int {
	scoreList := [EgMaxSeat]int{}
	firstCard := h.outcardList[h.firstSeat]
	firstColor := (firstCard & 0xF0) >> 4

	out2 := false
	fmt.Println("max ", firstCard, firstColor, h.zhuColor, h.outcardList)
	for i := 0; i < EgMaxSeat; i++ {
		card := h.outcardList[i]
		color := (card & 0xF0) >> 4
		val := int(card & 0x0F)
		fmt.Println("i ", i, card, color, val)

		// 大鬼 10000
		// 小鬼 9900
		// zhu2 8000
		// 2	7000
		// 主颜色 3000

		score := 0
		if card == 0x44 {
			score += 10000
		}
		if card == 0x43 {
			score += 9900
		}
		if val == 2 {
			if color == h.zhuColor {
				score += 8000
			} else {
				score += 7000
			}
			if !out2 {
				score += 500
			}
			out2 = true
		} else if color == h.zhuColor {
			score += 3000
		} else if color == firstColor {
			score += 500
		}
		score += val

		/*
		score := val
		if color == 4 { //鬼牌
			score += 10000
			if val == 4 {	//大		15000
				score += 5000
			} else if val == 3 { //小	13000
				score += 3000
			}
		} else if color == h.zhuColor {
			score += 1000		// 主牌颜色	1000
			if val == 2 {
				score += 500	// 主牌颜色+主牌 1500
			}
		} else if int(color) == firstColor {
			score += 100			// 和第一家牌相同	100
			score += val 			// 103 - 10x
		}
		*/
		scoreList[i] = score
	}

	maxSeat := 0
	maxVal := scoreList[maxSeat]
	for i := 1; i < EgMaxSeat; i++ {
		if scoreList[i] > maxVal {
			maxVal = scoreList[i]
			maxSeat = i
		}
	}

	fmt.Println("score list ", scoreList, maxSeat)
	return maxSeat
}

func (h *egHandler) computeScore(maxSeat int) {
	for i := 0; i < EgMaxSeat; i++ {
		val := h.outcardList[i] & 0x0F
		h.scoreList[i] += scoreCard[val]
	}
}

func (h *egHandler) outCard(user *userInfo, card int) {
	fmt.Println("user out card ", user, card)

	p := h.playerList[user.UserId]
	if p == nil {
		fmt.Println("out card ", h.playerList)
		return
	}

	fmt.Print("out seat", p.seat, h.curseat, h.leftRoud)
	if h.status != "outcard" {
		h.sendGameMessage(p, proto.CmdEgOutCard, &proto.ErguiUserOutCardRet{
			ErrCode: "notstatus",
		})
		return
	}

	if p.seat != h.curseat {
		h.sendGameMessage(p, proto.CmdEgOutCard, &proto.ErguiUserOutCardRet{
			ErrCode: "notyou",
		})
		return
	}

	if card < cardMin || card > cardMax {
		h.sendGameMessage(p, proto.CmdEgOutCard, &proto.ErguiUserOutCardRet{
			ErrCode: "notallow",
		})
		return
	}

	if card > len(p.indexs) || p.indexs[card] == 0 {
		h.sendGameMessage(p, proto.CmdEgOutCard, &proto.ErguiUserOutCardRet{
			ErrCode: "notexists",
		})
		return
	}

	if h.outcardList[p.seat] != 0 {
		h.sendGameMessage(p, proto.CmdEgOutCard, &proto.ErguiUserOutCardRet{
			ErrCode: "already",
		})
		return
	}

	if p.seat != h.firstSeat {
		err := h.checkOutCard(p.seat, card)
		if err != "ok" {
			h.sendGameMessage(p, proto.CmdEgOutCard, &proto.ErguiUserOutCardRet{
				ErrCode: err,
			})
			return
		}
	}

	if card == h.friend {
		h.friendSeat = p.seat
	}

	p.indexs[card]--
	h.outcardList[p.seat] = card

	nextSeat := (h.curseat + 1) % EgMaxSeat
	h.curseat = nextSeat

	newRound := false
	if h.curseat == h.firstSeat {
		maxSeat := h.getMaxScoreSeat()
		h.computeScore(maxSeat)

		fmt.Println("left round ", h.leftRoud)
		h.leftRoud--
		if h.leftRoud == 0 {
			h.bcGameMessage(proto.CmdEgOutCard, &proto.ErguiUserOutCardRet{
				ErrCode: "ok",
				Card: card,
				OutSeat: p.seat,
				NextSeat: -1,
			})
			h.finishGame()
			return
		}

		newRound = true
		nextFirstSeat := maxSeat

		h.curseat = nextFirstSeat
		h.firstSeat = nextFirstSeat
		h.outcardList = make([]int, EgMaxSeat)
	}

	fc := false
	if h.friend == card {
		fc = true
	}

	h.bcGameMessage(proto.CmdEgOutCard, &proto.ErguiUserOutCardRet{
		ErrCode: "ok",
		Card: card,
		OutSeat: p.seat,
		NextSeat: h.curseat,
		NewRound: newRound,
		FirstSeat: h.firstSeat,
		Score: h.scoreList,
		FriendCard: fc,
	})
	h.setTimer("outcardtimeout", 10, func() {
		h.outCard(h.playerList[h.curseat].userInfo, h.getRecommendOutcard())
	})
}

func (h *egHandler) finishGame() {

	ts, bs, xs := 0, 0, 0
	for i := 0; i < EgMaxSeat; i++ {
		ts += h.scoreList[i]
		if i != h.banker && i != h.friendSeat {
			xs += h.scoreList[i]
		}
		if i == h.banker {
			bs += h.scoreList[i]
		}
		if i == h.friendSeat {
			if h.banker != h.friendSeat {
				bs += h.scoreList[i]
			}
		}
	}
	fmt.Println("finish score ", h.scoreList, h.banker, h.friendSeat, ts, bs, xs)

	winScore := 0
	if h.bankerScore + xs == 100 {

	} else if h.bankerScore + xs > 100 {
		winScore = 100 - h.bankerScore - xs
	} else if h.bankerScore + xs < 100 {
		winScore = h.bankerScore + xs - 100
	}

	winScoreList := [EgMaxSeat]int{}

	bankerBaoChang := false

	if winScore < 0 { // banker lost
		if h.bankerScore == BANKER_SCORE_80 {
			if xs >= 50 {
				bankerBaoChang = true
			}
		} else if h.bankerScore == BANKER_SCORE_85 || h.bankerScore == BANKER_SCORE_90 || h.bankerScore == BANKER_SCORE_95 {
			if xs >= 40 {
				bankerBaoChang = true
			}
		} else if h.bankerScore == BANKER_SCORE_100 {
			if xs >= 30 {
				bankerBaoChang = true
			}
		} else if h.bankerScore == BANKER_SCORE_100_BAO {
			if xs >= 20 {
				bankerBaoChang = true
			}
		} else if h.bankerScore == BANKER_SCORE_100_GOU {
			if xs > 0 {
				bankerBaoChang = true
			}
		} else {
			fmt.Println("error in banker score")
		}

		for i := 0; i < EgMaxSeat; i++ {
			if i != h.banker && i != h.friendSeat {
				winScoreList[i] = -1 * winScore
				winScoreList[h.banker] += winScore
			}
			if i == h.friendSeat {
				if h.banker == h.friendSeat {
					winScoreList[h.banker] += winScore
				} else {
					if bankerBaoChang {
					} else {
						winScoreList[h.friendSeat] = winScore
						winScoreList[h.banker] += -1 * winScore
					}
				}
			}
		}

	} else {	// banker win
		for i := 0; i < EgMaxSeat; i++ {
			if i != h.banker && i != h.friendSeat {
				winScoreList[i] = -1 * winScore
				winScoreList[h.banker] += winScore
			}
		}
		if h.banker != h.friendSeat {
			winScoreList[h.banker] += -1 * winScore
			winScoreList[h.friendSeat] = winScore
		}
	}

	multiple := 1
	if h.bankerScore == BANKER_SCORE_100_BAO {
		multiple = 2
	} else if h.bankerScore == BANKER_SCORE_100_GOU {
		multiple = 4
	}
	for i := 0; i < EgMaxSeat; i++ {
		winScoreList[i] = winScoreList[i] * multiple
	}


	winMoney := [EgMaxSeat]float32{}
	for i := 0; i < EgMaxSeat; i++ {
		winMoney[i] = float32(winScoreList[i]) * h.getDijin()
		h.totalWinMoney[i] += winMoney[i]
	}

	guangtou := false
	selfBanker := false
	if xs <= 0 {
		guangtou = true
	}
	if h.curBanker == h.friendSeat {
		selfBanker = true
	}

	gs := &proto.ErguiGameFinish{
		ErrCode: "ok",
		BankerBaoChang: bankerBaoChang,
		GuangTou: guangtou,
		BankerIsFriend: selfBanker,
		GameScore: h.scoreList[:],
		XiQianMoney: h.xiqian[:],
		WinMoney: winMoney[:],
	}
	if h.gameRound + 1 == h.conf.Round {
		gs.TotalWinMoney = h.totalWinMoney[:]
		h.bcGameMessage(proto.CmdEgGameFinish, gs)
		h.room.mgr.lb.reqChn <- &lbRequest{
			req: func() {
				h.room.mgr.releaseRoom(h.room.id)
			},
		}
		fmt.Println("room release start")
		return
	} else {
		h.bcGameMessage(proto.CmdEgGameFinish, gs)
	}

	h.gameRound++
	fmt.Println("finish game winscore ", winScore)

	// clear
	for _, p := range h.playerList {
		p.ready = false
		p.callScore = 0
	}

	h.scoreList = make([]int, EgMaxSeat)
	h.xiqian = make([]float32, EgMaxSeat)
	h.zhuColor = -1
	h.outcardList = make([]int, EgMaxSeat)
	h.firstCard = -1
	h.canCallGou = false
	h.status = "finish"
	h.curseat = -1
	h.bankerScore = 0
	h.isCalled = false
	h.leftRoud = MaxPlayerCardCount
	h.friendSeat = -1
	h.friend = -1
}

func (h *egHandler) startGame() {
	copy(h.cards, cardBox)
	randCards(h.cards)

	fmt.Println("game start ", h.playerList)

	index := 0
	for _, p := range h.playerList {
		p.RoomCard = p.RoomCard - _config.consumeRoomCard
		p.handCard = h.cards[index:index+MaxPlayerCardCount]
		fmt.Print("c ", p.UserName, p.handCard)
		index += MaxPlayerCardCount
		p.indexs = make([]int, MaxCardIndex)
		for i := 0; i < MaxPlayerCardCount; i++ {
			p.indexs[p.handCard[i]]++
		}
	}
	h.bottomCard = h.cards[index:]

	if h.banker == -1 {
		h.banker = rand.Intn(EgMaxSeat)
	}
	//h.curBanker = (h.banker + 1) % EgMaxSeat
	h.curBanker = h.banker
	h.firstCall = h.curBanker
	h.canCallGou = false

	h.setTimer("callbankerTimeout", 10, func() {
		h.callScore(h.playerList[h.curBanker].userInfo, BANKER_SCORE_NO)
	})

	h.status = "call"

	for _, p := range h.playerList {
		cl := make([]int, len(p.handCard))
		for i := 0; i < len(p.handCard); i++ {
			cl[i] = int(p.handCard[i])
		}
		h.sendGameMessage(p, proto.CmdEgGameStart, &proto.ErguiGameStart{
			Banker:   h.banker,
			CardList: cl,
			CurRound: h.gameRound,
		})
	}
}

func (h *egHandler) setTimer(_ string, t int, fn func()) {
	/*
	h.tid++
	h.stoped = false
	h.t = time.AfterFunc(time.Duration(t) * time.Second, func() {
		h.room.reqCh <- &erguiRoomReq{
			cmd: "t",
			req: h.tid,
			data: fn,
		}
	})
	*/
}

func (h *egHandler) killTimer() {
	/*
	if !h.stoped {
		h.stoped = true
		h.t.Stop()
	}
	*/
}

func (h *egHandler) onTimer(req uint32, fn func()) {
	if h.stoped || req < h.tid {
		return
	}
	fn()
}

func (h *egHandler) setSeat(p *playerInfo) int {
	for i := 0; i < EgMaxSeat; i++ {
		if h.seatList[i] == nil {
			return i
		}
	}
	return -1
}


func (h *egHandler) sendMessage(p *playerInfo, cmd uint32, data interface{}) {
	h.room.mgr.lb.dog.sendClientMessage(p.Cid, cmd, data)
}

func (h *egHandler) bcMessage(cmd uint32, data interface{}) {
	uids := make([]uint32, 0)
	for _, p := range h.playerList {
		uids = append(uids, p.Cid)
	}
	h.room.mgr.lb.dog.bcClientMessage(uids, cmd, data)
}


func marshalData(data interface{}) []byte{
	if data, err := msgpacker.Marshal(data); err != nil {
		fmt.Println("marshal data error ", err)
		return nil
	} else {
		return data
	}
}

func (h *egHandler) sendGameMessage(p *playerInfo, cmd uint32, data interface{}) {
	//msg := marshalData(data)
	fmt.Println("send game message ", p.UserId, cmd, data)
	h.sendMessage(p, proto.CmdUserGameMessage, &proto.UserMessageRet{
		Cmd: cmd,
		Msg: data,
	})
}

func (h *egHandler) bcGameMessage(cmd uint32, data interface{}) {
	//msg := marshalData(data)
	fmt.Println("bc game message", cmd, data)
	h.bcMessage(proto.CmdUserGameMessage, &proto.UserMessageRet{
		Cmd: cmd,
		Msg: data,
	})
}
