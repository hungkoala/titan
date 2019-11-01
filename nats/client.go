package nats

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	Addr    string
	Subject string
}

func (srv *Client) Request(rq *Request) (*Response, error) {
	c, err := NewConnection(srv.Addr)
	if err != nil {
		return nil, err
	}
	defer func(c *Connection) {
		fmt.Print("close connection .......")
		c.Conn.Close()
	}(c)

	return c.SendRequest(srv.Subject, rq)
}

func Get(addr string, subject string, path string) (*Response, error) {
	return SendJsonRequest(addr, subject, path, nil, nil, "GET")
}

func Post(addr string, subject string, path string, body *interface{}) (*Response, error) {
	return SendJsonRequest(addr, subject, path, body, nil, "POST")
}

func Put(addr string, subject string, path string, body *interface{}) (*Response, error) {
	return SendJsonRequest(addr, subject, path, body, nil, "PUT")
}

func Delete(addr string, subject string, path string, body *interface{}) (*Response, error) {
	return SendJsonRequest(addr, subject, path, body, nil, "PUT")
}
func Head(addr string, subject string, path string) (*Response, error) {
	return SendJsonRequest(addr, subject, path, nil, nil, "HEAD")
}
func Trace(addr string, subject string, path string, body *interface{}) (*Response, error) {
	return SendJsonRequest(addr, subject, path, body, nil, "TRACE")
}
func Patch(addr string, subject string, path string, body *interface{}) (*Response, error) {
	return SendJsonRequest(addr, subject, path, body, nil, "PATCH")
}

func SendJsonRequest(addr string, subject string, path string, body interface{}, header http.Header, method string) (*Response, error) {
	client := &Client{Addr: addr, Subject: subject}
	jBody, err := bodyToJson(body)
	if err != nil {
		return nil, err
	}
	var h http.Header
	h = header
	if h == nil {
		h = http.Header{}
	}
	rq := &Request{URL: path, Method: method, Headers: h, Body: jBody}
	return client.Request(rq)
}

func bodyToJson(body interface{}) ([]byte, error) {
	if body != nil {
		return json.Marshal(body)
	}
	return nil, nil
}
