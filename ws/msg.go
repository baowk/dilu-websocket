package ws

import "github.com/gorilla/websocket"

type Msg struct {
	WsType int
	Data   []byte `json:"data"`
}

func NewMsg(msType int, data []byte) *Msg {
	return &Msg{
		msType,
		//binary.BigEndian.Uint32(data[:4]),
		data[4:],
	}
}

func NewBinaryMsg(data []byte) *Msg {
	return &Msg{
		WsType: websocket.BinaryMessage,
		Data:   data,
	}
}

func NewTextMsg(text string) *Msg {
	return &Msg{
		WsType: websocket.TextMessage,
		Data:   []byte(text),
	}
}
