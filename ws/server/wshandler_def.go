package server

import (
	"net/http"
	"time"

	"github.com/baowk/dilu-websocket/ws"

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
		CheckOrigin:   true,
		PingEnabled:   true,
		PingPeriod:    time.Second * 10,
		PingFailCount: 3,
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
	ws.NewWsChannl(wsSocket, wsh, wsh.config.PingEnabled, wsh.config.PingPeriod, wsh.config.PingFailCount)
	core.Log.Debug("wsHandler.Connect end")
}

// 心跳
func (wsh *DefWsHandler) Heartbeat(wsc *ws.WsChannal) error {
	core.Log.Debug("wsHandler.Heartbeat", zap.String("ID:", wsc.GetId()))
	//TODO 心跳下发
	return nil
}

// 单个发送消息
func (wsh *DefWsHandler) Send(deviceId string, data []byte) error {
	core.Log.Debug("wsHandler.Send", zap.String(deviceId, string(data)))
	c := wsh.hub.Get(deviceId)
	if c != nil {
		return c.Write(ws.NewBinaryMsg(data))
	}
	return nil
}

// 发送组内消息
func (wsh *DefWsHandler) SendToGroup(groupId string, data []byte) error {
	core.Log.Debug("wsHandler.SendToGroup", zap.String(groupId, string(data)))
	cs := wsh.hub.GetGroup(groupId)
	if cs != nil {
		for _, v := range cs {
			core.Log.Debug("wsHandler.SendToGroup：", zap.String(v.GetId(), string(data)))
			v.Write(ws.NewBinaryMsg(data))
		}
	}
	return nil
}

// 全部发送
func (wsh *DefWsHandler) Broadcast(data []byte) error {
	core.Log.Debug("wsHandler.Broadcast")
	for _, v := range wsh.hub.All() {
		v.Write(ws.NewBinaryMsg(data))
	}
	return nil
}

// 断开链接
func (wsh *DefWsHandler) Disconnect(wsc *ws.WsChannal) error {
	core.Log.Debug("wsHandler.Disconnect start")
	//TODO 断链处理
	core.Log.Debug("wsHandler.Disconnect end")
	return nil
}

// 读取消息处理
func (wsh *DefWsHandler) MsgHandler(wsc *ws.WsChannal, msg *ws.Msg) {
	core.Log.Debug("wsHandler.HandlerMsg")
	//TODO 消息处理
	return
}

func (wsh *DefWsHandler) GetConfig() *ws.WsConfig {
	return wsh.config
}
