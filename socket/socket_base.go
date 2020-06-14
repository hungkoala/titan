package socket

import (
	"bytes"
	"fmt"
	"time"

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
	session *Session
	// The websocket connection.
	conn          *websocket.Conn
	socketManager *SocketManager

	// Buffered channel of outbound messages.
	send chan []byte

	onMessage func([]byte)

	id string

	logger logur.Logger

	isClosed bool
}

func (a *BaseSocket) Close() {
	if a.isClosed {
		return
	}
	close(a.send)
	a.isClosed = true
}

func (a *BaseSocket) Send(message []byte) {
	if a.isClosed {
		return
	}
	select {
	case a.send <- message:
	default: //cannot send to it because channel has been closed
		a.socketManager.UnRegister(a)
		a.Close()
	}
}

func (a *BaseSocket) GetSession() *Session {
	return a.session
}

func (a *BaseSocket) GetId() string {
	return a.id
}

func (a *BaseSocket) StartReader() {
	defer func() {
		if r := recover(); r != nil {
			a.logger.Debug("Panic Recovered in socket reader")
		}
		a.logger.Debug("Reader: Socket connection  is closing")
		a.socketManager.UnRegister(a)
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
				a.logger.Error(fmt.Sprintf("Socket Unexpected Close Error %+v\n ", err))
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		a.onMessage(message)

	}
}

func (a *BaseSocket) StartWriter() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		if r := recover(); r != nil {
			a.logger.Debug("Panic Recovered in socket writer")
		}
		a.logger.Debug("Writer: Socket connection  is closing")
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
				a.logger.Debug("can't ping client", map[string]interface{}{"err": err})
				return
			}
		}
	}
}

func (a *BaseSocket) StartHandler() {
	go a.StartReader()
	go a.StartWriter()
}
