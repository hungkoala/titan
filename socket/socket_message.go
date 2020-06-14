package socket

import "gitlab.com/silenteer-oss/titan"

// enum definitions
type Topic string

const (
	PING Topic = "PING"
	PONG Topic = "PONG"
)

type MessageResponse struct {
	Id          *string             `json:"id,omitempty"`
	Topic       *Topic              `json:"topic,omitempty"`
	MessageBody *string             `json:"messageBody,omitempty"`
	Headers     map[string][]string `json:"headers,omitempty"`
	StatusCode  *int64              `json:"statusCode,omitempty"`
}

type MessageRequest struct {
	Id            string              `json:"id"`
	Topic         Topic               `json:"topic"`
	MessageBody   string              `json:"messageBody"`
	Headers       map[string][]string `json:"headers"`
	StatusCode    int64               `json:"statusCode"`
	Subject       string              `json:"subject"`
	ResponseTopic Topic               `json:"responseTopic"`
}

type Session titan.UserInfo
