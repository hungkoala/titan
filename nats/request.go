package nats

import (
	"net/http"
)

const (
	XRequestId      = "X-REQUEST-ID"
	contentType     = "Content-Type"
	jsonContentType = "application/json"
)

// Request is a simple struct
type Request struct {
	Method  string      `json:"method"`
	Headers http.Header `json:"headers"`
	Body    []byte      `json:"body"`
	URL     string      `json:"url"`
	Subject string      `json:"subject"`
}

func (r *Request) GetID() string {
	if r.Headers == nil {
		return ""
	}
	return r.Headers.Get(XRequestId)
}
