package proto

const (
	_ = iota
	CmdAuth
	CmdProbe
	CmdData
)

// common header
// 2bytes content len
// 1byte cmd
// 1byte version

var HeaderSize = int32(4)

type Header []byte

func NewHeader() Header {
	return make([]byte, HeaderSize)
}

func (h Header) GetCmd() int {
	return int(h[3])
}

func (h Header) SetCmd(cmd int) {
	h[3] = byte(cmd)
}

func (h Header) GetBodyLen() int32 {
	return (int32(h[1]) << 8) + int32(h[2])
}

func (h Header) SetBodyLen(size int32) {
	h[1] = byte(size >> 8)
	h[2] = byte(size)
}
