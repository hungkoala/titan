package socket

import (
	"bytes"
	"fmt"
	"time"

	"gitlab.com/silenteer-oss/titan"

	"github.com/gorilla/websocket"
	"logur.dev/logur"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 5000 * 1024
)

type BaseSocket struct {
	Session *Session
	// The websocket connection.
	conn          *websocket.Conn
	SocketManager *SocketManager

	// Buffered channel of outbound messages.
	send chan []byte

	OnMessage func([]byte)

	Id string

	Logger logur.Logger

	IsClosed bool
}

func NewBaseSocket(session *Session, conn *websocket.Conn, socketManager *SocketManager, logger logur.Logger) BaseSocket {
	id := titan.RandomString(20)
	return BaseSocket{
		Session:       session,
		conn:          conn,
		SocketManager: socketManager,
		send:          make(chan []byte, 9000),
		Id:            id,
		Logger:        logur.WithFields(logger, map[string]interface{}{"id": id}),
	}
}

func (a *BaseSocket) Close() {
	if a.IsClosed {
		return
	}
	close(a.send)
	a.IsClosed = true
}

func (a *BaseSocket) Send(message []byte) {
	if a.IsClosed {
		return
	}
	select {
	case a.send <- message:
	default: //cannot send to it because channel has been closed
		a.SocketManager.UnRegister(a)
		a.Close()
	}
}

func (a *BaseSocket) GetSession() *Session {
	return a.Session
}

func (a *BaseSocket) GetId() string {
	return a.Id
}

func (a *BaseSocket) StartReader() {
	defer func() {
		if r := recover(); r != nil {
			a.Logger.Debug("Panic Recovered in socket reader")
		}
		a.Logger.Debug("Reader: Socket connection  is closing")
		a.SocketManager.UnRegister(a)
		a.conn.Close()
	}()

	a.conn.SetReadLimit(maxMessageSize)
	_ = a.conn.SetReadDeadline(time.Now().Add(pongWait))
	a.conn.SetPongHandler(func(string) error {
		_ = a.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := a.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				a.Logger.Error(fmt.Sprintf("Socket Unexpected Close Error %+v\n ", err))
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		a.OnMessage(message)

	}
}

func (a *BaseSocket) StartWriter() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		if r := recover(); r != nil {
			a.Logger.Debug("Panic Recovered in socket writer")
		}
		a.Logger.Debug("Writer: Socket connection  is closing")
		ticker.Stop()
		a.conn.Close()
	}()
	for {
		select {
		case message, ok := <-a.send:
			//a.logger.Debug("Socket server is writing to browser ")

			_ = a.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				_ = a.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := a.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = a.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := a.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				a.Logger.Debug("can't ping client", map[string]interface{}{"err": err})
				return
			}
		}
	}
}

func (a *BaseSocket) StartHandler() {
	go a.StartReader()
	go a.StartWriter()
}
