package network

import (
	"exportor/defines"
	"net"
	"errors"
	"io"
	"msgpacker"
	"exportor/proto"
	"mylog"
)


type message struct {
	cmd 	uint32
	data 	interface{}
}

type tcpClient struct {
	*netContext
	id 				uint32
	conn 			net.Conn
	opt 			*defines.NetClientOption
	sendCh			chan *message
	headerBuf 		[8]byte
	packer 			*msgpacker.MsgPacker
	authed 			bool
}

func newTcpClient(opt *defines.NetClientOption) *tcpClient {
	client := &tcpClient{
		opt: opt,
		sendCh: make(chan *message, 1024),
		packer: msgpacker.NewMsgPacker(),
		netContext: newNetContext(),
	}
	return client
}

func (client *tcpClient) Id(id uint32) {
	client.id = id
}

func (client *tcpClient) GetId() uint32 {
	return client.id
}

func (client *tcpClient) GetRemoteAddress() string {
	return client.conn.RemoteAddr().String()
}

func (client *tcpClient) Connect() error {
	conn, err := net.Dial("tcp", client.opt.Host)
	if err != nil {
		mylog.Infoln("connect error", err)
		return err
	}
	mylog.Infoln("connect addr ", client.opt.Host)
	client.conn = conn
	client.opt.ConnectCb(client)
	if client.opt.AuthCb(client) != nil {
		client.Close()
		return errors.New("connect auth error")
	}
	go client.sendLoop()
	client.recvLoop()

	return nil
}

func (client *tcpClient) Close() error {
	if client.opt.CloseCb != nil {
		client.opt.CloseCb(client)
	}
	return nil
}

func (client *tcpClient) Send(cmd uint32, data interface{}) error {
	//mylog.Infoln("send message 1", cmd, data)
	client.sendCh <- &message{cmd: cmd, data: data}
	//mylog.Infoln("send message 2", cmd, data)
	return nil
}

func (client *tcpClient) sendLoop() {
	//mylog.Infoln("client send loop error 1")
	for {
		select {
		case m:= <- client.sendCh:
			//mylog.Infoln("send message 2 --------------->", m)
			if raw, err :=client.packer.Pack(m.cmd, m.data); err != nil {
				//mylog.Infoln("send msg error ", m, err)
			} else {
				mylog.Infoln("send message 2 --------------____>", raw[:8], string(raw))
				client.conn.Write(raw)
			}
		}
	}
	mylog.Infoln("client send loop error")
}

func (client *tcpClient) recvLoop() {
	client.authed = true
	defer func() {
		client.Close()
	}()

	go func() {
		for {
			m, err := client.readMessage()
			if err != nil {
				mylog.Infoln("client recv lopp decode msg error", err)
				return
			}
			//mylog.Infoln("callcb ", m)
			client.opt.MsgCb(client, m)
		}
	}()
}

func (client *tcpClient) Auth() (*proto.Message, error) {
	if client.authed {
		return nil, errors.New("already authed")
	}
	return client.readMessage()
}

func (client *tcpClient) readMessage() (*proto.Message, error) {
	//mylog.Infoln("client recv message ")
	if _, err := io.ReadFull(client.conn, client.headerBuf[:]); err != nil {
		return nil, err
	}

	mylog.Infoln("client recv message headerBuf ", client.headerBuf)
	header, err := client.packer.Unpack(client.headerBuf[:])
	if err != nil {
		return nil, err
	}

	//mylog.Infoln("client recv message header ", header)
	body := make([]byte, header.Len)
	if _, err := io.ReadFull(client.conn, body[:]); err != nil {
		return nil, err
	}
	header.Msg = body
	mylog.Infoln("client recv message finish", header.Len, header.Cmd,string(header.Msg))
	return header, nil
}

func (client *tcpClient) configureConn(conn net.Conn) {
	client.conn = conn
}
