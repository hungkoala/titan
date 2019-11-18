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

func (r *Response) GetBody() []byte {
	return r.Body
}

func (r *Response) GetHeaders() http.Header {
	return r.Headers
}

//Deprecated: please  use response builder instead
func (r *Response) Header() http.Header {
	return r.Headers
}

//Deprecated: please  use response builder instead
func (r *Response) Write(b []byte) (n int, err error) {
	r.Body = b
	r.WriteHeader(http.StatusOK)
	return len(b), nil
}

//Deprecated: please  use response builder instead
func (r *Response) WriteHeader(code int) {
	if _, hasType := r.Headers[contentType]; !hasType {
		r.Headers.Add(contentType, "application/json; charset=utf-8")
	}
	r.StatusCode = code
}
