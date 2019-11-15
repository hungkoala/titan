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

	// used in server side
	requestParams map[string][]string
	routeParams   map[string]string
}

func (r *Request) RequestParams() map[string][]string {
	return r.requestParams
}

func (r *Request) RouteParams() map[string]string {
	return r.routeParams
}
