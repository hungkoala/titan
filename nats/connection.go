package nats

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type Connection struct {
	Conn *nats.EncodedConn
}

// NewConnection connects to default nats address
func NewConnection(address string) (*Connection, error) {
	log.Println("Connecting to NATS Server at: " + address)

	conn, err := nats.Connect(address)
	if err != nil {
		return nil, fmt.Errorf("error connecting to NATS: %v", err)
	}

	enc, err := nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		return nil, fmt.Errorf("cannot construct JSON encoded connection to NATS: %v", err)
	}

	return &Connection{Conn: enc}, nil
}

func (srv *Connection) SendRequest(subject string, rq *Request) (*Response, error) {
	rp := Response{}
	err := srv.Conn.Request("test", rq, &rp, 3*time.Second)
	return &rp, err
}
