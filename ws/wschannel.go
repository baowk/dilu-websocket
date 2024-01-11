package ws

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/baowk/dilu-core/core"
	"github.com/gorilla/websocket"
)

type WsChannal struct {
	Conn        *websocket.Conn //socket链接
	Ctx         map[string]any  //上下文
	Output      chan *Msg       //写队列
	open        bool            //是否打开
	rwmutex     *sync.RWMutex   //读写锁
	CloseChan   chan byte       // 关闭通知
	WsHandler   WsHandler       //处理器
	hbEnabled   bool            //心跳开关
	hbDuration  time.Duration   //心跳时长
	hbFailCount uint8           //心跳失败次数
}

func NewWsChannl(conn *websocket.Conn, wsHandler WsHandler, heartbeatEnabled bool, heartbeatDuration time.Duration, heartbeatFailCount uint8) *WsChannal {
	wsc := &WsChannal{
		Conn:        conn,
		WsHandler:   wsHandler,
		hbDuration:  heartbeatDuration,
		hbFailCount: heartbeatFailCount,
		hbEnabled:   heartbeatEnabled,
		CloseChan:   make(chan byte, 1),
		Output:      make(chan *Msg, 1000),
		rwmutex:     new(sync.RWMutex),
		open:        true,
	}
	if wsc.hbDuration < time.Second {
		wsc.hbDuration = time.Second * 10
	}
	go wsc.readLoop()
	go wsc.writeLoop()
	if wsc.hbEnabled {
		go wsc.heartbeat()
	}
	return wsc
}

func (wsc *WsChannal) heartbeat() {
	var failCnt uint8
	for {
		if wsc.IsClosed() {
			return
		}
		time.Sleep(wsc.hbDuration)
		err := wsc.WsHandler.Heartbeat(wsc)
		if err != nil {
			failCnt++
		} else {
			failCnt = 0
		}
		if failCnt > wsc.hbFailCount {
			wsc.Close()
			return
		}
	}
}

func (wsc *WsChannal) readLoop() {
	for {
		// 读一个message
		msgType, data, err := wsc.Conn.ReadMessage()
		if err != nil {
			goto error
		}

		if msgType == websocket.CloseMessage {
			goto closed
		}

		if msgType == websocket.PingMessage {
			core.Log.Debug("socket ping")
			if err := wsc.Conn.WriteMessage(websocket.PongMessage, nil); err != nil {
				goto error
			}
			continue
		}

		req := &Msg{
			WsType: msgType,
			Data:   data,
		}
		core.Log.Debug("Read message")
		go wsc.WsHandler.MsgHandler(wsc, req)
		if !wsc.open {
			goto closed
		}
	}
error:
	wsc.Close()
closed:
}

func (wsc *WsChannal) writeLoop() {
	for {
		select {
		// 取一个应答
		case msg := <-wsc.Output:
			// 写给websocket
			if err := wsc.Conn.WriteMessage(msg.WsType, msg.Data); err != nil {
				goto error
			}
		case <-wsc.CloseChan:
			goto closed
		}
	}
error:
	wsc.Close()
closed:
}

const id_name = "id"

func (c *WsChannal) GetId() string {
	// c.rwmutex.RLock()
	// defer c.rwmutex.RUnlock()
	v, ok := c.Ctx[id_name]
	if !ok {
		v = c.Conn.RemoteAddr().String()
		c.Ctx[id_name] = v
	}
	return v.(string)
}

func (c *WsChannal) SetId(id string) {
	c.rwmutex.Lock()
	defer c.rwmutex.Unlock()
	c.Ctx[id_name] = id
}

func (c *WsChannal) Write(msg *Msg) error {
	select {
	case c.Output <- msg:
	case <-c.CloseChan:
		return errors.New("websocket closed")
	}
	return nil
}

func (c *WsChannal) Close() {
	core.Log.Debug("close websocket" + c.GetId())
	c.rwmutex.Lock()
	defer c.rwmutex.Unlock()
	//处理关闭
	if c.Conn != nil {
		c.Conn.Close()
	}
	c.Conn = nil
	if c.open {
		c.WsHandler.Disconnect(c)
		c.open = false
		close(c.CloseChan)
	}
	//释放资源
	c = nil
}

func (wsc *WsChannal) Set(key string, value interface{}) {
	wsc.rwmutex.Lock()
	defer wsc.rwmutex.Unlock()

	if wsc.Ctx == nil {
		wsc.Ctx = make(map[string]any)
	}
	wsc.Ctx[key] = value
}

func (wsc *WsChannal) Get(key string) (value any, exists bool) {
	// wsc.rwmutex.RLock()
	// defer wsc.rwmutex.RUnlock()
	if wsc.Ctx != nil {
		value, exists = wsc.Ctx[key]
	}
	return
}

func (wsc *WsChannal) GetString(key string) string {
	if v, exists := wsc.Get(key); exists {
		return v.(string)
	}
	return ""
}

func (wsc *WsChannal) GetInt(key string) int {
	if v, exists := wsc.Get(key); exists {
		return v.(int)
	}
	return 0
}

func (wsc *WsChannal) GetBool(key string) bool {
	if v, exists := wsc.Get(key); exists {
		return v.(bool)
	}
	return false
}

func (wsc *WsChannal) MustGet(key string) any {
	if value, exists := wsc.Get(key); exists {
		return value
	}

	panic("Key \"" + key + "\" does not exist")
}

func (wsc *WsChannal) Del(key string) {
	wsc.rwmutex.Lock()
	defer wsc.rwmutex.Unlock()
	if wsc.Ctx != nil {
		delete(wsc.Ctx, key)
	}
}

func (wsc *WsChannal) IsClosed() bool {
	return !wsc.open
}

func (wsc *WsChannal) LocalAddr() net.Addr {
	return wsc.Conn.LocalAddr()
}

func (wsc *WsChannal) RemoteAddr() net.Addr {
	return wsc.Conn.RemoteAddr()
}
