package protocol

import (
	"bytes"
	"encoding/binary"
	// "errors" // Not strictly needed for this basic version but good for future additions
)

// IDataPack 封包、拆包的接口
type IDataPack interface {
	GetHeadLen() uint32
	Pack(msg *Message) ([]byte, error)
	Unpack([]byte) (*Message, error)
}

// DataPack 封包拆包实例，根据服务端Zinx实际使用情况，这里改为 LittleEndian
type DataPack struct{}

// NewDataPack 初始化 DataPack
func NewDataPack() IDataPack {
	return &DataPack{}
}

// GetHeadLen 获取包头长度方法 (uint32 for DataLen + uint32 for ID)
func (dp *DataPack) GetHeadLen() uint32 {
	return 8
}

// Pack 封包方法
func (dp *DataPack) Pack(msg *Message) ([]byte, error) {
	dataBuff := bytes.NewBuffer([]byte{})

	// 写 DataLen (使用 LittleEndian)
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetDataLen()); err != nil {
		return nil, err
	}

	// 写 MsgID (使用 LittleEndian)
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgID()); err != nil {
		return nil, err
	}

	// 写 Data 数据 (数据本身字节序不转换，原样写入)
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetData()); err != nil {
		return nil, err
	}

	return dataBuff.Bytes(), nil
}

// Unpack 拆包方法 (只解出head信息，数据部分由调用者根据DataLen读取)
func (dp *DataPack) Unpack(binaryData []byte) (*Message, error) {
	dataBuff := bytes.NewReader(binaryData)
	msg := &Message{}

	// 读 DataLen (使用 LittleEndian)
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	// 读 MsgID (使用 LittleEndian)
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.ID); err != nil {
		return nil, err
	}

	// 可选：在这里检查DataLen是否超出允许的最大包长度
	// if msg.DataLen > MaxPacketSize {
	// 	 return nil, errors.New("message data length too large")
	// }

	return msg, nil
}
