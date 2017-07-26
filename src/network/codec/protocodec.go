package codec
/*
import (
	"github.com/golang/protobuf/proto"
	myproto "exportor/proto"
	"net"
	"encoding/binary"
	"gameproto/clipb"
	"io"
	"math"
)

const maxRecvBufLen = math.MaxUint16

type protoCodec struct {
	conn       net.Conn
	order      binary.ByteOrder
	cliHeader  *gameproto.CliMsgHeader
	dcliHeader *gameproto.CliMsgHeader
	recvBuf    [maxRecvBufLen]byte
	header     [2]byte
}

func NewClientCodec() *protoCodec {
	return &protoCodec{
		dcliHeader: &gameproto.CliMsgHeader{},
		cliHeader: &gameproto.CliMsgHeader{},
	}
}

func (pd *protoCodec) Encode(m *myproto.Message) error {
/*
	body, err := proto.Marshal(m.Msg.(proto.Message))
	if err != nil {
		return nil
	}

	pd.cliHeader.Cmd = uint32(m.Cmd)
	pd.cliHeader.Len = uint32(len(body))
	pd.cliHeader.Code = 1000

	header, err := proto.Marshal(pd.cliHeader)
	if err != nil {
		return nil
	}

	msgLen := 2 + uint32(len(header)) + pd.cliHeader.Len
	data := make([]byte, msgLen)
	binary.BigEndian.PutUint16(data[:2], uint16(len(header)))
	data = append(data, header...)
	data = append(data, body...)

	if _, err := pd.conn.Write(data); err != nil {
		return err
	}

	return nil
}

func (pd *protoCodec) Decode() (*myproto.Message, error) {
	if _, err := io.ReadFull(pd.conn, pd.recvBuf[:2]); err != nil {
		return err
	}

	headerLen := binary.BigEndian.Uint16(pd.header)
	if _, err := io.ReadFull(pd.conn, pd.recvBuf[:headerLen]); err != nil {
		return err
	}

	err := proto.Unmarshal(pd.recvBuf[:headerLen], pd.dcliHeader)
	if err != nil {
		return err
	}

	if _, err := io.ReadFull(pd.conn, pd.recvBuf[:pd.dcliHeader.Len]); err != nil {
		return err
	}

	return &myproto.Message{
		Cmd:   pd.dcliHeader.Cmd,
		Magic: "pb",
		Msg:   myproto.NewPbMessage(pd.dcliHeader.Cmd),
	}, nil
	return nil, nil
}

func (pd *protoCodec) DecodeRaw(size int) ([]byte, error) {

	return nil, nil
}
*/
