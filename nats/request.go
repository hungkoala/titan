package nats

import (
	"net/http"

	"github.com/go-chi/chi"
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

type SRequest struct {
	Method        string
	Path          string
	Headers       http.Header
	Body          []byte
	RequestParams map[string][]string
	RouteParams   chi.RouteParams
}
