package defines

import (
	"time"
)

type IServer interface {
	Start() error
	Stop() error
}

type IGateway interface {
	IServer
}

type ILobby interface {
	IServer
}

type IGameServer interface {
	IServer
}

type IDbProxy interface {
	IServer
}

type CommunicatorCb func([]byte)
type ICommunicatorClient interface {
	Notify(chanel string, v interface{})	error
	JoinChanel(chanel string, reg bool, time int, cb CommunicatorCb) error
	WaitChannel(channel string, time int) ([] byte, error)
}

type IMsgPublisher interface {
	IServer
	WaitPublish(channel string, key string, data interface{}) error
	SendPublish(channel string, data interface{}) error
}

type IMsgConsumer interface {
	IServer
	WaitMessage(channel string, key string, t time.Duration) interface{}
	GetMessage(channel string, key string) interface{}
}

type ICommunicator interface {
	WaitPublish(channel string, key string, data interface{}) error
	SendPublish(channel string, data interface{}) error
	WaitMessage(channel string, key string, t time.Duration) interface{}
	GetMessage(channel string, key string) interface{}
}

type ICacheClient interface {
	IServer
}

type ICacheLoader interface {
	LoadUsers(name string)
}
