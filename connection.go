package titan

import (
	"fmt"
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

	return &Connection{Conn: enc}, nil
}

func (srv *Connection) SendRequest(rq *Request, subject string) (*Response, error) {
	if subject == "" {
		return nil, errors.New("nats subject cannot be nil")
	}
	rp := Response{}
	err := srv.Conn.Request(subject, rq, &rp, GetNatsConfig().GetReadTimeoutDuration()+5*time.Second)
	return &rp, err
}

func (srv *Connection) ensureGlobalSubscriber() error {
	rq, _ := NewReqBuilder().
		Get("/some/random/non/existent/path/just/to/ensure/the/global/subscriber/to/be/created").
		Build()

	var rs interface{}
	err := srv.Conn.Request("pleasedonotuse.thisuglyhackystuff.pleasepleaseplease", rq, &rs, 1*time.Millisecond)
	if err != nil && err.Error() == "nats: timeout" {
		fmt.Println("Expected nats: timeout while ensure global subscriber")
		return srv.Conn.Flush()
	}

	return err
}
