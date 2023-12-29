package sever

import (
	"dilu-websocket/ws"
	"encoding/binary"
	"net/http"

	"github.com/baowk/dilu-core/core"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type DefWsHandler struct {
	hub        *ws.Hub
	wsUpgrader websocket.Upgrader
	config     *ws.WsConfig
}

var wsHandler = NewDefWsHandler()

func NewDefWsHandler() *DefWsHandler {
	cfg := &ws.WsConfig{
		CheckOrigin: true,
		// WriteWait:   10 * 1000,
		// ReadWait:    10 * 1000,
		// PingPeriod:  60 * 1000,
		// MaxMessage:  1024 * 1024 * 10,
	}
	wsUpgrader := websocket.Upgrader{
		// 允许所有CORS跨域请求
		CheckOrigin: func(r *http.Request) bool {
			return cfg.CheckOrigin
		},
	}
	return &DefWsHandler{
		hub:        ws.NewHub(),
		wsUpgrader: wsUpgrader,
		config:     cfg,
	}
}

// 基于gin的链接
func (wsh *DefWsHandler) ConnectGin(c *gin.Context) {
	wsh.Connect(c.Writer, c.Request)
}

// 链接
func (wsh *DefWsHandler) Connect(w http.ResponseWriter, r *http.Request) {
	core.Log.Debug("wsHandler.Connect")
	wsSocket, err := wsh.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	wsc := &ws.WsChannal{
		Conn:      wsSocket,
		Output:    make(chan *ws.Msg, 1000),
		CloseChan: make(chan byte),
		WsHandler: wsh,
	}
	wsc.Run()
	core.Log.Debug("wsHandler.Connect end")
}

// 心跳
func (wsh *DefWsHandler) Heartbeat(wsc *ws.WsChannal) {
	core.Log.Debug("wsHandler.Heartbeat")
	// var failCnt int
	// for {
	// 	time.Sleep(3 * time.Second)
	// 	if err := wsc.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
	// 		failCnt++
	// 		core.Log.Error("socket write err", zap.Int("failcount", failCnt), zap.Error(err))
	// 		if failCnt > 3 {
	// 			wsc.Close()
	// 			break
	// 		}
	// 	} else {
	// 		failCnt = 0
	// 	}
	// }
}

// 单个发送消息
func (wsh *DefWsHandler) Send(deviceId string, data []byte) error {
	core.Log.Debug("wsHandler.Send")
	return wsh.hub.Get(deviceId).Write(ws.NewMsg(1, data))
}

// 发送组内消息
func (wsh *DefWsHandler) SendToGroup(groupId string, data []byte) error {
	core.Log.Debug("wsHandler.SendToGroup")
	for _, v := range wsh.hub.GetGroup(groupId) {
		v.Write(ws.NewMsg(1, data))
	}
	return nil
}

// 全部发送
func (wsh *DefWsHandler) SendAll(data []byte) error {
	core.Log.Debug("wsHandler.SendAll")
	for _, v := range wsh.hub.All() {
		v.Write(ws.NewMsg(1, data))
	}
	return nil
}

// 断开链接
func (wsh *DefWsHandler) Disconnect(wsc *ws.WsChannal) error {
	core.Log.Debug("wsHandler.Disconnect")

	return nil
}

const (
	MsgType_HEART = 1 + iota
	MsgType_LOGIN
	MsgType_NEW_NOTICE
)

// 读取消息处理
func (wsh *DefWsHandler) MsgHandler(wsc *ws.WsChannal, msg *ws.Msg) {
	core.Log.Debug("wsHandler.HandlerMsg")
	if len(msg.Data) < 4 {
		return
	}
	msgType := binary.BigEndian.Uint32(msg.Data[:4])
	switch msgType {
	case MsgType_HEART:
		wsc.Write(msg)
		return
	case MsgType_LOGIN:
		wsh.loginHandler(wsc, msg.Data[4:])
		return
	case MsgType_NEW_NOTICE:
		return
	default:
		return
	}
}

func (wsh *DefWsHandler) loginHandler(wsc *ws.WsChannal, data []byte) error {
	core.Log.Debug("wsHandler.loginHandler", zap.Any("data", string(data)))
	wsc.Write(ws.NewMsg(MsgType_LOGIN, data))
	return nil
}
