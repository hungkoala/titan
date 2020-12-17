package titan

import (
	"fmt"
	"runtime/debug"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"logur.dev/logur"
)

type MessageHandler func(*Message) error

type Registration struct {
	Subject string
	Queue   string
	Handler MessageHandler
}

type MessageSubscriber struct {
	logger        logur.Logger
	registrations []*Registration
	subscriptions []*nats.Subscription
}

func NewMessageSubscriber(logger logur.Logger) *MessageSubscriber {
	return &MessageSubscriber{logger: logger}
}

func (s *MessageSubscriber) Register(subject string, queue string, handler MessageHandler) {
	s.registrations = append(s.registrations, &Registration{
		Subject: subject,
		Queue:   queue,
		Handler: s.createHandlerWithRecover(handler),
	})
}

func (s *MessageSubscriber) subscribe(conn *nats.EncodedConn) error {
	for index, registration := range s.registrations {
		sub, err := conn.QueueSubscribe(registration.Subject, registration.Queue, registration.Handler)
		if err != nil {
			return errors.WithMessagef(err, "Nats subscription [%d] error ", index)
		}
		s.subscriptions = append(s.subscriptions, sub)
	}

	s.registrations = nil

	return nil
}

func (s *MessageSubscriber) drain() {
	for _, sub := range s.subscriptions {
		er := sub.Drain()
		if er != nil {
			s.logger.Error(fmt.Sprintf("Drain error: %+v\n ", er))
		}
	}
}

func (s *MessageSubscriber) createHandlerWithRecover(next MessageHandler) MessageHandler {
	return func(msg *Message) (err error) {
		var ctx *Context
		defer func() {
			if _err := recover(); _err != nil {
				err = fmt.Errorf("panicking from subscriber %+v", _err)
				errMsg := fmt.Sprintf("stacktrace from panic subscriber: %s", string(debug.Stack()))
				if ctx != nil {
					ctx.Logger().Error(errMsg)
				} else {
					s.logger.Error(errMsg)
				}
				fmt.Println(errMsg)
			}
		}()
		ctx, _ = msg.context()
		return next(msg)
	}
}
