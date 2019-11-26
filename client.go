package titan

import (
	"encoding/json"
	"fmt"
	"strings"

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

	subject := Url2Subject(rq.URL)
	fmt.Println("Send request to subject " + subject)
	return c.SendRequest(rq, subject)
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

func GetDefaultClient() *Client {
	config := GetNatsConfig()
	return &Client{config.Servers}
}

func Url2Subject(url string) string {
	// not found
	if !strings.Contains(url, "/") {
		return url
	}
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	s := strings.Split(url, "/")
	l := 4
	if len(s)+1 <= 4 {
		l = len(s)
	}
	s1 := s[1:l]
	return strings.Join(s1, ".")
}
