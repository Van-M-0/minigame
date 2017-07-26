package codec
/*
import (
	"net"
	"exportor/proto"
)

type mpCodec struct {
	conn		net.Conn
	decode 		*codec.MsgpackHandle
	recvBuf 	[maxRecvBufLen]byte
	header 		[2]byte
}

func NewMpCodec() *mpCodec {
	return &mpCodec{
		decode: new(codec.MsgpackHandle),
	}
}

func (md *mpCodec) Conn(conn net.Conn) {
	md.conn = conn
}

func (md *mpCodec) EncodeMsg(m *proto.Message) error {
/*
	body, err := msgpack.Marshal(m)
	if err != nil {
		return err
	}

	binary.BigEndian.PutUint16(md.header[:], uint16(len(body)))

	data := make([]byte, len(body) + 2)
	copy(data[:2], md.header[:])
	copy(data[2:], body[:])

	if _, err := md.conn.Write(data); err != nil {
		return err
	}
	return nil
}


func (md *mpCodec) DecodeMsg() (*proto.Message, error) {
	if _, err := io.ReadFull(md.conn, md.recvBuf[:2]); err != nil {
		return nil, err
	}

	bodyLen := binary.BigEndian.Uint16(md.recvBuf[:2])

	if _, err := io.ReadFull(md.conn, md.recvBuf[:bodyLen]); err != nil {
		return nil, err
	}

	var m proto.Message
	if err := msgpack.Unmarshal(md.recvBuf[:bodyLen], m); err != nil {
		return nil, err
	}

	m1, err := proto.NewMessage(m.Cmd)
	if err != nil {
		return nil, err
	}
	if err := msgpack.Unmarshal(md.recvBuf[:bodyLen], m1); err != nil {
		return nil, err
	}

	return m1, nil
}

func (md *mpCodec) EncodeGate(message *proto.Message) error {
	return nil
}

func (md *mpCodec) DecodeGate() (*proto.GateGameHeader, error) {
	return nil, nil
}
*/
