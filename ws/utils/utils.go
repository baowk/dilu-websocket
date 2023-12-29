package utils

import "encoding/binary"

func BuildMsg(msgType int, data []byte) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(msgType))
	return append(buf, data...)
}
