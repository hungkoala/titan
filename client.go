package titan

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type Client struct {
	Addr    string
	Subject string
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

	return c.SendRequest(rq, srv.Subject)
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
	rp, err := srv.request(ctx, rq)

	if err != nil {
		var rpErr *Response
		if err.Error() == "nats: timeout" {
			rpErr = &Response{Status: "Request Timeout", StatusCode: 408}
		} else {
			rpErr = &Response{Status: "Internal Server Error", StatusCode: 500}
		}
		return nil, &HttpClientResponseError{Message: "Nats Client Request Timeout", Response: rpErr, Cause: err}
	}

	if rp.StatusCode >= 400 {
		return nil, &HttpClientResponseError{Message: rp.Status, Response: rp}
	}

	if rp.StatusCode >= 300 {
		return nil, &HttpClientResponseError{Message: "HTTP 3xx Redirection was not implemented yet", Response: rp}
	}

	if rp.StatusCode >= 200 {
		return rp, nil
	}

	if rp.StatusCode < 200 {
		return rp, &HttpClientResponseError{Message: "HTTP 1xx Informational response was not implemented yet", Response: rp}
	}
	return rp, nil
}

func NewClient(config *Config) *Client {
	return &Client{config.Servers, config.Subject}
}
