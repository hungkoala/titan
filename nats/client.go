package nats

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type Client struct {
	Addr string
}

func (srv *Client) request(ctx *Context, rq *Request) (*Response, error) {
	c, err := NewConnection(srv.Addr)
	if err != nil {
		return nil, err
	}
	defer func(c *Connection) {
		_ = c.Conn.Flush()
		c.Conn.Close()
	}(c)

	return c.SendRequest(rq)
}

func (srv *Client) SendAndReceiveJson(ctx *Context, rq *Request, receive interface{}) error {
	msg, err := srv.SendRequest(ctx, rq)
	if err != nil {
		return err
	}

	if msg.Body == nil || len(msg.Body) == 0 {
		return nil
	}

	err = json.Unmarshal(msg.Body, &receive)
	if err != nil {
		return errors.WithMessage(err, "nats client json parsing error")
	}
	return nil
}

func (srv *Client) SendRequest(ctx *Context, rq *Request) (*Response, error) {
	msg, err := srv.request(ctx, rq)

	if err != nil {
		return nil, errors.WithMessage(err, "nats client error")
	}

	if msg.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("nats client error, status code %d", msg.StatusCode))
	}
	return msg, nil
}

func NewClient(addr string) *Client {
	return &Client{addr}
}
