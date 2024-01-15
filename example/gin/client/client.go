// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore
// +build ignore

package main

import (
	"encoding/binary"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/baowk/dilu-websocket/example/gin/msg"
	"github.com/baowk/dilu-websocket/ws/utils"
	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:5888", "http service address")

var token = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDQ0MjU4NjIsImp0aSI6IjU0NDRlMDY4ZmJlMDQ0ZTVhYzFhYTBmZjllYmZjZWIxIiwiaXNzIjoiZTg2Yzc3ZDFjNzljNGQwY2JmMjAyMDVhNTA1ODZkNWMifQ.C59tNaAtdbpWZgzb7m8xkjxgeFRwCH0kGyE4VEuLujE"

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			log.Println("ping message")
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			if mt == websocket.PingMessage {
				c.WriteMessage(websocket.PongMessage, message)
				continue
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			log.Println("send:", t)

			lm := []byte("token")
			err := c.WriteMessage(websocket.BinaryMessage, utils.BuildMsg(msg.MsgType_LOGIN, lm))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
func Int32ToBytes(i int32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(i))
	return buf
}
