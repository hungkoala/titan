package titan

import (
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"time"
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
		return nil, errors.WithMessage(err, "Error connecting to NATS")
	}

	enc, err := nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		return nil, errors.WithMessage(err, "Cannot construct JSON encoded connection to NATS")
	}

	err = ensureGlobalSubscriber(enc)
	if err != nil {
		return nil, errors.WithMessage(err, "Cannot ensure that global subscriber has been created and flushed")
	}

	return &Connection{Conn: enc}, nil
}

func (srv *Connection) SendRequest(rq *Request, subject string) (*Response, error) {
	if subject == "" {
		return nil, errors.New("nats subject cannot be nil")
	}
	rp := Response{}
	err := srv.Conn.Request(subject, rq, &rp, GetNatsConfig().GetReadTimeoutDuration())
	return &rp, err
}

func ensureGlobalSubscriber(conn *nats.EncodedConn) error {
	rq, _ := NewReqBuilder().
		Get("/some/random/non/existent/path/just/to/ensure/the/global/subscriber/to/be/created").
		Build()

	var rs interface{}
	err := conn.Request("pleasedonotuse.thisuglyhackystuff.pleasepleaseplease", rq, &rs, 1*time.Millisecond)
	if err != nil && err.Error() == "nats: timeout" {
		// expected error, do nothing
		return conn.Flush()
	}

	return err
}
