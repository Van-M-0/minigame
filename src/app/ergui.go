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




}

func (eg *erguiRoom) UserOfflineTimeout(user *userInfo) {
	fmt.Println("user offline timeout ", user)
}

func (eg *erguiRoom) CreateRoom(id int, user *userInfo, req *proto.UserCreateRoom, conf []byte) string {
	var c proto.ErguiRoomConf
	msgpacker.UnMarshal(conf, &c)
	ret := eg.handler.createRoom(id, user, req, &c)
	if ret == "ok" {
		eg.run()
		eg.id = id
	}
	return ret
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

func (eg *erguiRoom) sfCall() {
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
		}
	}
}

func (eg *erguiRoom) run() {

	go func() {
		for {
			eg.sfCall()
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

	creator 	int
	playerList 	map[int]*playerInfo
	seatList 	[EgMaxSeat]*playerInfo
	cards 		[]int
	bottomCard	[]int

	banker 		int
	curBanker 	int
	bankerScore int

	zhuColor 	int


	friend		int


	firstSeat 	int
	firstCard 	int
	curseat 	int
	leftRoud 	int

	outcardList []int

	scoreList 	[]int
}

func newEgHandler(r *erguiRoom) *egHandler {
	handler := &egHandler{}
	handler.room = r
	handler.playerList = make(map[int]*playerInfo)
	handler.cards = make([]int, MaxCardCount)
	handler.banker = -1
	handler.status = "ready"
	handler.outcardList = make([]int, EgMaxSeat)
	handler.scoreList = make([]int, EgMaxSeat)
	handler.leftRoud = MaxPlayerCardCount
	return handler
}

func (h *egHandler) createRoom(roomId int, user *userInfo, req *proto.UserCreateRoom, conf *proto.ErguiRoomConf) string {
	h.creator = user.UserId

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
		Player 		[]*resEnterUser
	}

	var enterRes egUserEnter
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

func (h *egHandler) userReady(user *userInfo, msg interface{}) {

	fmt.Println("user ready data ", msg)

	p := h.playerList[user.UserId]
	if p.ready {
		return
	}
	p.ready = true

	type readyRes struct {
		Seat 		int
		Ready 		bool
	}

	res := &readyRes{
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

	for _, p := range h.playerList {
		if p.callScore != 0 && score <= p.callScore {
			h.sendGameMessage(p, proto.CmdEgUserCallBanker, &proto.ErguiCallbankerRet{
				ErrCode: "small",
			})
			return
		}
	}

	h.killTimer()
	p.callScore = score

	if score == BANKER_SCORE_100_GOU {
		h.banker = h.curBanker
		h.bankerScore = score
	} else {
		nextBanker := (h.curBanker + 1) % EgMaxSeat
		nextp := h.seatList[nextBanker]
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
				for _, p := range h.playerList {
					if p.callScore != BANKER_SCORE_NO {
						count++
					}
				}
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

	fmt.Println("checkout ", firstCard, zcard, zcolor, seat)

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
			if p.indexs[i] != 0 {
				return "you-fu"
			}
		}

		if haveZhuCard() {
			return "youzhu-fu"
		}
		if haveZhuColor() {
			return "youzhu-fucolor"
		}
	}

	return "ok"
}

func (h *egHandler) getMaxScoreSeat() int {
	scoreList := [EgMaxSeat]int{}
	firstColor := h.firstCard & 0xF0

	for i := 0; i < EgMaxSeat; i++ {
		card := h.outcardList[i]
		color := card & 0xF0
		val := int(card & 0x0F)
		if color == 0x40 {
			val += 10000
		} else if color == h.zhuColor {
			val += 1000
		} else if int(color) == firstColor {
			val += 100
		}
		scoreList[i] = val
	}

	maxSeat := 0
	maxVal := scoreList[maxSeat]
	for i := 1; i < EgMaxSeat; i++ {
		if scoreList[i] > maxVal {
			maxVal = scoreList[i]
		}
	}

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

	if p.seat != h.firstSeat {
		err := h.checkOutCard(p.seat, card)
		if err != "ok" {
			h.sendGameMessage(p, proto.CmdEgOutCard, &proto.ErguiUserOutCardRet{
				ErrCode: err,
			})
			return
		}
	}

	p.indexs[card]--
	h.outcardList[p.seat] = card

	nextSeat := (h.curseat + 1) % EgMaxSeat
	h.curseat = nextSeat

	newRound := false
	if h.curseat == h.firstSeat {
		maxSeat := h.getMaxScoreSeat()
		h.computeScore(maxSeat)

		if h.leftRoud == 0 {
			h.finishGame()
			return
		}

		newRound = true
		nextFirstSeat := maxSeat

		h.curseat = nextFirstSeat
		h.firstSeat = nextFirstSeat
		h.leftRoud--
		h.outcardList = make([]int, EgMaxSeat)
	}

	h.bcGameMessage(proto.CmdEgOutCard, &proto.ErguiUserOutCardRet{
		ErrCode: "ok",
		Card: card,
		OutSeat: p.seat,
		NextSeat: h.curseat,
		NewRound: newRound,
		FirstSeat: h.firstSeat,
	})
	h.setTimer("outcardtimeout", 10, func() {
		h.outCard(h.playerList[h.curseat].userInfo, h.getRecommendOutcard())
	})
}

func (h *egHandler) finishGame() {
	multiple := 1
	if h.bankerScore == BANKER_SCORE_100_BAO {
		multiple = 4
	} else if h.bankerScore == BANKER_SCORE_100_GOU {
		multiple = 2
	}

	winScore := [EgMaxSeat]int{}
	for i := 0; i < EgMaxSeat; i++ {
		winScore[i] = h.scoreList[i] * multiple
	}
	fmt.Println("finish game winscore ", winScore)
}

func (h *egHandler) startGame() {
	copy(h.cards, cardBox)
	randCards(h.cards)

	fmt.Println("game start ", h.playerList)

	index := 0
	for _, p := range h.playerList {
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
