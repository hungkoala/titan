package nats

import (
	"net/http"
)

// Request is a simple struct
type Request struct {
	Method  string      `json:"method"`
	Headers http.Header `json:"headers"`
	Body    []byte      `json:"body"`
	URL     string      `json:"url"`
}

func (r *Request) GetID() string {
	if r.Headers == nil {
		return ""
	}
	return r.Headers.Get(XRequestId)
}

func (r *Request) SetContentType(contentType string) {
	if r.Headers == nil {
		r.Headers = http.Header{}
	}
	r.Headers.Add("Content-Type", contentType)
}

func (r *Request) GetContentType() string {
	return r.Headers.Get("Content-Type")
}

func NewRequest(method, url string, body []byte) *Request {
	return &Request{Method: method, URL: url, Body: body, Headers: http.Header{}}
}
