package titan

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"logur.dev/logur"
)

type MessageHandler func(*Message) error

type Registration struct {
	Subject string
	Queue string
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
		Queue: queue,
		Handler: handler,
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

func (s *MessageSubscriber) unsubscribe() {
	for _, sub := range s.subscriptions {
		er := sub.Unsubscribe()
		if er != nil {
			s.logger.Error(fmt.Sprintf("Unsubscribe error: %+v\n ", er))
		}
	}
}