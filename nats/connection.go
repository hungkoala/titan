package nats

import (
	"github.com/nats-io/nats.go"
	"log"
)

// Connection holds exposed connection from nats
type Connection struct {
	Conn *nats.Conn
	Enc  *nats.EncodedConn
}

// NewConnection connects to default nats address
func NewConnection(address string) *Connection {
	log.Println("Connecting to NATS Server at: " + address)

	if address == "" {
		address = nats.DefaultURL
	}

	conn, err := nats.Connect(address)
	if err != nil {
		panic("Cannot connect to NATS Server")
	}

	enc, err := nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal("Cannot construct JSON encoded connection")
	}

	return &Connection{
		Conn: conn,
		Enc:  enc,
	}
}
