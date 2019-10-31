package nats

import (
	"net/http"
)

// Request is a simple struct
type Response struct {
	Status     string      `json:"reason"` // e.g. "200 OK"
	StatusCode int         `json:"code"`   // e.g. 200
	Headers    http.Header `json:"headers"`
	Body       []byte      `json:"body"`
}

func (res *Response) SetContentType(value string) {
	res.Headers.Set("Content-Type", value)
}

func (r *Response) Flush() {
}

func (r *Response) Header() http.Header {
	return r.Headers
}

func (r *Response) Write(p []byte) (n int, err error) {
	r.Body = p
	r.WriteHeader(http.StatusOK)
	return len(p), nil
}

func (r *Response) WriteHeader(code int) {
	// Set a default Content-Type
	if _, hasType := r.Headers["Content-Type"]; !hasType {
		r.Headers.Add("Content-Type", "application/json; charset=utf-8")
	}
	r.StatusCode = code
}
