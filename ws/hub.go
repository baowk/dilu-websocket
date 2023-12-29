package ws

import (
	"sync"

	"github.com/baowk/dilu-core/core"
	"go.uber.org/zap"
)

type Hub struct {
	clients  map[string]*WsChannal
	group    map[string][]*WsChannal
	isClosed bool
	rwmutex  *sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*WsChannal),
		rwmutex: new(sync.RWMutex),
	}
}

func (h *Hub) Add(clientId string, wsChan *WsChannal) {
	core.Log.Debug("add client", zap.String(clientId, "start"))
	h.rwmutex.Lock()
	defer h.rwmutex.Unlock()
	h.clients[clientId] = wsChan
	core.Log.Debug("add client", zap.String(clientId, "end"))
}

func (h *Hub) Del(clientId string) {
	core.Log.Debug("del client", zap.String(clientId, "start"))
	h.rwmutex.Lock()
	defer h.rwmutex.Unlock()
	delete(h.clients, clientId)
	core.Log.Debug("del client", zap.String(clientId, "end"))
}

func (h *Hub) Get(clientId string) *WsChannal {
	h.rwmutex.RLock()
	defer h.rwmutex.RUnlock()
	return h.clients[clientId]
}

func (h *Hub) Len() int {
	h.rwmutex.RLock()
	defer h.rwmutex.RUnlock()
	return len(h.clients)
}

func (h *Hub) Closed() bool {
	h.rwmutex.RLock()
	defer h.rwmutex.RUnlock()
	return h.isClosed
}

func (h *Hub) All() []*WsChannal {
	h.rwmutex.RLock()
	defer h.rwmutex.RUnlock()
	s := make([]*WsChannal, 0, len(h.clients))
	for _, v := range h.clients {
		s = append(s, v)
	}
	return s
}

func (h *Hub) AddGroup(groupId string, wsChan *WsChannal) {
	core.Log.Debug("add AddGroup", zap.String(groupId, "start"))
	h.rwmutex.Lock()
	defer h.rwmutex.Unlock()
	if h.group == nil {
		h.group = make(map[string][]*WsChannal)
	}
	h.group[groupId] = append(h.group[groupId], wsChan)
	core.Log.Debug("add AddGroup", zap.String(groupId, "end"))
}

func (h *Hub) GetGroup(groupId string) []*WsChannal {
	h.rwmutex.Lock()
	defer h.rwmutex.Unlock()
	if h.group == nil {
		return nil
	}
	return h.group[groupId]
}

func (h *Hub) DelFromGroup(groupId string, wsChan *WsChannal) {
	core.Log.Debug("del DelFromGroup start", zap.String(groupId, "ok"))
	h.rwmutex.Lock()
	defer h.rwmutex.Unlock()
	if h.group == nil {
		return
	}
	if group, ok := h.group[groupId]; ok {
		for i, v := range group {
			if v == wsChan {
				group = append(group[:i], group[i+1:]...)
				h.group[groupId] = group
				break
			}
		}
		if len(group) == 0 {
			delete(h.group, groupId)
		}
	}
	core.Log.Debug("del DelFromGroup end", zap.String(groupId, "ok"))
}
