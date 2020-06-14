package socket

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"gitlab.com/silenteer-oss/titan"
	"logur.dev/logur"
)

// ----------------------- Login user socket -----------------------------------------------------------------
type DefaultSocket struct {
	BaseSocket
}

func CreateDefaultSocket(session *Session, conn *websocket.Conn, socketManager *SocketManager, logger logur.Logger) *DefaultSocket {
	id := titan.RandomString(20)
	socket := &DefaultSocket{
		BaseSocket: BaseSocket{
			session:       session,
			conn:          conn,
			socketManager: socketManager,
			send:          make(chan []byte, 9000),
			id:            id,
			logger:        logur.WithFields(logger, map[string]interface{}{"id": id}),
		},
		//natClient: titan.GetDefaultClient(),
	}
	socket.onMessage = socket.OnMessage
	socketManager.Register(socket)
	return socket
}

func (c *DefaultSocket) handleOnMessage(message []byte) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Debug("Panic Recovered in socket handleOnMessage")
		}
	}()

	var request MessageRequest

	if err := json.Unmarshal(message, &request); err != nil {
		c.logger.Error(fmt.Sprintf("Unmashal common socket error %+v\n ", err))
		c.sendValidationErrorResponse(fmt.Sprintf("Unmashal error %+v\n ", err), 400, request.ResponseTopic, request.Id)
		return
	}

	// handle ping pong
	if request.Topic == "PING" {
		pongMessage := PONG
		code := int64(200)
		pong, err := json.Marshal(MessageResponse{Topic: &pongMessage, StatusCode: &code})
		if err != nil {
			c.logger.Error(fmt.Sprintf("Mashal Pong message  error %+v\n ", err))
			return
		}

		c.socketManager.Broadcast(pong, func(session *Session) bool {
			return session.UserId != "" && c.session.UserId == session.UserId
		})
		return
	}

	// should pass this request to request handler
	//c.sendMessageResponse(messageResponse)
}

func (c *DefaultSocket) OnMessage(message []byte) {
	go c.handleOnMessage(message)
}

func (c *DefaultSocket) sendValidationErrorResponse(message string, statusCode int, responseTopic Topic, id string) {
	c.logger.Error(message)
	if responseTopic == "" {
		return
	}
	body, _ := json.Marshal(titan.DefaultJsonError{
		Message: message,
		Links:   map[string][]string{"self": {}},
		TraceId: "",
	})
	bodyString := string(body)
	statusCode64 := int64(statusCode)

	messageResponse := &MessageResponse{
		Topic:       &responseTopic,
		MessageBody: &bodyString, // convert message to json object
		StatusCode:  &statusCode64,
		Id:          &id,
	}
	c.sendMessageResponse(messageResponse)
}

func (c *DefaultSocket) sendMessageResponse(message *MessageResponse) {
	responseByte, err := json.Marshal(message)
	if err != nil {
		c.logger.Error(fmt.Sprintf("Mashal response message  error %+v\n ", err))
	} else {
		c.Send(responseByte)
	}
}
