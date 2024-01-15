package main

import (
	"time"

	"github.com/baowk/dilu-websocket/example/gin/handler"
	"github.com/baowk/dilu-websocket/ws"
	"github.com/gin-gonic/gin"
)

func main() {
	h := handler.NewDefWsHandler(&ws.WsConfig{
		CheckOrigin:   true,
		PingEnabled:   true,
		PingPeriod:    time.Second * 10,
		PingFailCount: 3,
	})
	r := gin.Default()
	r.GET("/ws", h.ConnectGin)
	r.Run(":5888")
}
