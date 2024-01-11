package ws

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type WsHandler interface {
	ConnectGin(c *gin.Context)
	Connect(w http.ResponseWriter, r *http.Request)
	Heartbeat(wsc *WsChannal) error
	Send(deviceId string, data []byte) error
	SendToGroup(groupId string, data []byte) error
	Broadcast(data []byte) error
	MsgHandler(wsc *WsChannal, msg *Msg)
	Disconnect(wsc *WsChannal) error
}

type WsConfig struct {
	CheckOrigin bool
	WriteWait   time.Duration
	ReadWait    time.Duration
	PingPeriod  time.Duration
	MaxMessage  int64
}
