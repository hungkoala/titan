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
		Handler: createHandlerWithRecover(handler),
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

//func (s *MessageSubscriber) unsubscribe() {
//	for _, sub := range s.subscriptions {
//		er := sub.Unsubscribe()
//		if er != nil {
//			s.logger.Error(fmt.Sprintf("Unsubscribe error: %+v\n ", er))
//		}
//	}
//}

func (s *MessageSubscriber) drain() {
	for _, sub := range s.subscriptions {
		er := sub.Drain()
		if er != nil {
			s.logger.Error(fmt.Sprintf("Drain error: %+v\n ", er))
		}
	}
}

func createHandlerWithRecover(next MessageHandler) MessageHandler {
	return func(msg *Message) (err error) {
		defer func() {
			if _err := recover(); _err != nil {
				err = fmt.Errorf("panicking from subscriber %+v", _err)
				fmt.Println("stacktrace from panic subscriber: \n" + string(debug.Stack()))
			}
		}()
		return next(msg)
	}
}
