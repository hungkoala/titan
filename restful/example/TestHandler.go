package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"gitlab.com/silenteer-oss/titan"

	"gitlab.com/silenteer-oss/titan/socket"
	"logur.dev/logur"
)

type TestHandler struct {
}

func Handle(socketManager *socket.SocketManager, logger logur.Logger, w http.ResponseWriter, r *http.Request) {
	// check and set header
	if r.Header.Get(titan.XRequestId) == "" {
		r.Header.Set(titan.XRequestId, titan.RandomString(6))
	}

	upgrader := websocket.Upgrader{}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("Upgrade socket connection error %+v", err))
		return
	}

	//Parse user to Session
	session := &socket.Session{}

	client := CreateDefaultSocket(session, wsConn, socketManager, logger)
	client.StartHandler()
	go client.SendKaka()
}

// ----------------------- Login user socket -----------------------------------------------------------------
type DefaultSocket struct {
	socket.BaseSocket
}

func CreateDefaultSocket(session *socket.Session, conn *websocket.Conn, socketManager *socket.SocketManager, logger logur.Logger) *DefaultSocket {
	soc := &DefaultSocket{
		BaseSocket: socket.NewBaseSocket(session, conn, socketManager, logger),
	}
	soc.OnMessage = soc.HandleOnMessage
	socketManager.Register(soc)
	return soc
}

func (c *DefaultSocket) handleOnMessage(message []byte) {
	defer func() {
		if r := recover(); r != nil {

			c.Logger.Debug("Panic Recovered in socket handleOnMessage")
		}
	}()

	//c.sendMessageResponse(messageResponse)
}

func (c *DefaultSocket) HandleOnMessage(message []byte) {
	go c.handleOnMessage(message)
}

func (c *DefaultSocket) sendMessageResponse(message *longLatStruct) {
	responseByte, err := json.Marshal(message)
	if err != nil {
		c.Logger.Error(fmt.Sprintf("Mashal response message  error %+v\n ", err))
	} else {
		c.Send(responseByte)
	}
}

func (c *DefaultSocket) SendKaka() {
	rand.Seed(42)
	ticker := time.NewTicker(500 * time.Millisecond)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Tick at", t)
				k := longLatStruct{
					Long: rand.Float64(),
					Lat:  rand.Float64(),
				}
				latlong := fmt.Sprintf("%f %f", k.Lat, k.Long)
				c.Send([]byte(latlong))
				//c.sendMessageResponse(&k)
			}
		}
	}()

	time.Sleep(10 * time.Minute)
	ticker.Stop()
	done <- true
	fmt.Println("Ticker stopped")
}

type longLatStruct struct {
	Long float64 `json:"longitude"`
	Lat  float64 `json:"latitude"`
}
