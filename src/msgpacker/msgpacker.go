package msgpacker

import (
	"bytes"
	"encoding/binary"
	"exportor/proto"
	"encoding/json"
)

type MsgPacker struct {}

func NewMsgPacker() *MsgPacker {
	return &MsgPacker{}
}

func (mp *MsgPacker) GetHeadLen() int {
	return 8
}

func (mp *MsgPacker) Unpack(data []byte) (*proto.Message, error){
	reader := bytes.NewReader(data)
	header := &proto.Message{}

	if err := binary.Read(reader, binary.LittleEndian, &header.Len); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &header.Cmd); err != nil {
		return nil, err
	}

	return header, nil
}

func (mp *MsgPacker) Pack(cmd uint32, data interface{}) ([] byte, error) {
	writer := bytes.NewBuffer([]byte{})
	body, err := Marshal(data)
	if err != nil {
		return nil, err
	}

	if err := binary.Write(writer, binary.LittleEndian, uint32(len(body))); err != nil {
		return nil, err
	}
	if err := binary.Write(writer, binary.LittleEndian, cmd); err != nil {
		return nil, err
	}
	if err := binary.Write(writer, binary.LittleEndian, body); err != nil {
		return nil, err
	}

	return writer.Bytes(), nil
}

func Marshal(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func UnMarshal(data []byte, p interface{}) error {
	return json.Unmarshal(data, p)
}