package titan

import (
	"time"

	"github.com/pkg/errors"

	"github.com/nats-io/nats.go"
)

type IConnection interface {
	Publish(subject string, v interface{}) error
	SendRequest(rq *Request, subject string) (*Response, error)
	Flush() error
	Close()
	Drain()
	Subscribe(subject string, cb Handler) (ISubscription, error)
}

type ISubscription interface {
	Unsubscribe() error
	Drain() error
	SetPendingLimits(msgLimit, bytesLimit int) error
}

type Connection struct {
	Conn *nats.EncodedConn
}

func (c *Connection) Subscribe(subject string, cb Handler) (ISubscription, error) {
	return c.Conn.QueueSubscribe(subject, "", cb)
}

//
func newIConnection() IConnection {
	return &Connection{}
}

// Connect will attempt to connect to the NATS system.
// The url can contain username/password semantics. e.g. nats://derek:pass@localhost:4222
// Comma separated arrays are also supported, e.g. urlA, urlB.
// Options start with the defaults but can be overridden.
func NewConnection(url string, options ...nats.Option) (*Connection, error) {
	conn, err := nats.Connect(url, options...)
	if err != nil {
		return nil, errors.WithMessage(err, "Error connecting to NATS")
	}

	enc, err := nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		return nil, errors.WithMessage(err, "Cannot construct JSON encoded connection to NATS")
	}

	return &Connection{Conn: enc}, nil
}

func (c *Connection) SendRequest(rq *Request, subject string) (*Response, error) {
	if subject == "" {
		return nil, errors.New("nats subject cannot be nil")
	}
	rp := Response{}
	err := c.Conn.Request(subject, rq, &rp, GetNatsConfig().GetReadTimeoutDuration()+5*time.Second)
	return &rp, err
}

func (c *Connection) Publish(subject string, v interface{}) error {
	return c.Conn.Publish(subject, v)
}

func (c *Connection) Flush() error {
	return c.Conn.Flush()
}

func (c *Connection) Close() {
	c.Conn.Close()
}

func (c *Connection) Drain() {
	_ = c.Conn.Drain()
}
