package socket

import (
	"encoding/json"

	"gitlab.com/silenteer-oss/titan"

	"gitlab.com/silenteer-oss/hestia/socket_service/api"

	"strconv"
	"time"

	"emperror.dev/errors"
	"logur.dev/logur"
)

type Session titan.UserInfo

type Socket interface {
	GetId() string
	Send([]byte)
	Close()
	GetSession() *Session
}

// --------------------------- Socket manager code

type BroadcastInfo struct {
	message []byte
	filter  func(session *Session) bool
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type SocketManager struct {
	clients map[string]Socket

	// Inbound messages from the clients.
	broadcast chan *BroadcastInfo

	// Register requests from the clients.
	register chan Socket

	// Unregister requests from clients.
	unregister chan Socket

	logger logur.Logger

	stop chan interface{}
}

func InitSocketManager(logger logur.Logger) *SocketManager {
	return &SocketManager{
		broadcast:  make(chan *BroadcastInfo, 1000),
		register:   make(chan Socket, 1000),
		unregister: make(chan Socket, 1000),
		clients:    make(map[string]Socket),
		logger:     logger,
		stop:       make(chan interface{}, 1),
	}
}

func (m *SocketManager) Register(client Socket) {
	select {
	case m.register <- client:
	default:
		m.logger.Error("Cannot write to channel 'register'. Please increase its buffer size")
	}

}

func (m *SocketManager) UnRegister(client Socket) {
	select {
	case m.unregister <- client:
	default:
		m.logger.Error("Cannot write to channel 'unregister'. Please increase its buffer size")
	}

}

func (m *SocketManager) Broadcast(message []byte, filter func(session *Session) bool) {
	m.broadcast <- &BroadcastInfo{
		message: message,
		filter:  filter,
	}
}

func (m *SocketManager) BroadcastM(message *api.SendRequestMessage, filter func(session *Session) bool) error {
	messageByte, err := json.Marshal(message)
	if err != nil {
		return errors.WithMessage(err, "Marshal message error")
	}
	m.Broadcast(messageByte, filter)
	return nil
}

func (m *SocketManager) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case client := <-m.register:
			m.clients[client.GetId()] = client
			m.logger.Debug("Register client ", map[string]interface{}{"id": client.GetId(), "Total remain": len(m.clients)})

		case client := <-m.unregister:
			if _, ok := m.clients[client.GetId()]; ok {
				delete(m.clients, client.GetId())
				client.Close()
			}
			m.logger.Debug("Unregister client ", map[string]interface{}{"id": client.GetId(), "Total remain": len(m.clients)})
		case info := <-m.broadcast:
			for _, client := range m.clients {
				if info.filter(client.GetSession()) {
					client.Send(info.message)
				}
			}
		case <-ticker.C:
			m.logger.Debug("++++++++++++++++++ socket manager is running ....., total= " + strconv.Itoa(len(m.clients)))
		case <-m.stop:
			return
		}
	}
}

func (m *SocketManager) Start() {
	go panicRecover(m.run)
}

func (m *SocketManager) Stop() {
	if m.stop != nil {
		m.stop <- "stop"
	}
}

func (m *SocketManager) CloseAll() {
	for _, socketClient := range m.clients {
		socketClient.Close()
	}
}

// never die goroutine
func panicRecover(f func()) {
	defer func() {
		if err := recover(); err != nil {
			go panicRecover(f)
		}
	}()
	f()
}
