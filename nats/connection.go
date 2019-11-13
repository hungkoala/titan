package nats

import (
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type Connection struct {
	Conn *nats.EncodedConn
}

// Connect will attempt to connect to the NATS system.
// The url can contain username/password semantics. e.g. nats://derek:pass@localhost:4222
// Comma separated arrays are also supported, e.g. urlA, urlB.
// Options start with the defaults but can be overridden.
func NewConnection(url string, options ...nats.Option) (*Connection, error) {
	conn, err := nats.Connect(url, options...)
	if err != nil {
		return nil, fmt.Errorf("error connecting to NATS: %v", err)
	}

	enc, err := nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		return nil, fmt.Errorf("cannot construct JSON encoded connection to NATS: %v", err)
	}

	return &Connection{Conn: enc}, nil
}

func (srv *Connection) SendRequest(rq *Request) (*Response, error) {
	if rq.Subject == "" {
		return nil, errors.New("nats subject cannot be nil")
	}
	rp := Response{}
	err := srv.Conn.Request(rq.Subject, rq, &rp, 3*time.Second)
	return &rp, err
}
