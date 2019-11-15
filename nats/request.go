package nats

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

const (
	XRequestId      = "X-REQUEST-ID"
	XLoggerId       = "X-LOGGER-ID"
	XHostName       = "hostname"
	contentType     = "Content-Type"
	jsonContentType = "application/json"
)

type RequestParams map[string][]string

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

func (r *Request) HasBody() bool {
	return nil != r.Body
}

func (r *Request) BodyJson(v interface{}) error {
	if !r.HasBody() {
		return errors.New("body not found")
	}
	if err := json.Unmarshal(r.Body, &v); err != nil {
		return errors.WithMessage(err, "Json Unmarshal error ")
	}
	return nil
}
