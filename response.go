package titan

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
	//r.WriteHeader(http.StatusOK)
	return len(b), nil
}

//Deprecated: please  use response builder instead
func (r *Response) WriteHeader(code int) {
	if _, hasType := r.Headers[contentType]; !hasType {
		r.Headers.Add(contentType, "application/json; charset=utf-8")
	}
	r.StatusCode = code
}

// ----------------- response builder code ----------------------

// Request is a simple struct
type ResponseBuilder struct {
	statusCode int

	headers http.Header

	bodyProvider BodyProvider
}

// New returns a new default  Request.
func NewResBuilder() *ResponseBuilder {
	rq := &ResponseBuilder{
		statusCode: 200,
		headers:    make(http.Header),
	}
	rq.SetContentType(jsonContentType)
	return rq
}

func (r *ResponseBuilder) AddHeader(key, value string) *ResponseBuilder {
	r.headers.Add(key, value)
	return r
}

func (r *ResponseBuilder) SetHeader(key, value string) *ResponseBuilder {
	r.headers.Set(key, value)
	return r
}

func (r *ResponseBuilder) GetHeader(key string) string {
	return r.headers.Get(key)
}

func (r *ResponseBuilder) SetContentType(value string) {
	r.SetHeader(contentType, value)
}

func (r *ResponseBuilder) Body(body []byte) *ResponseBuilder {
	if body == nil {
		return r
	}
	return r.BodyProvider(byteBodyProvider{body: body})
}

func (r *ResponseBuilder) StatusCode(status int) *ResponseBuilder {
	r.statusCode = status
	return r
}

// BodyProvider sets the RequestBuilder's body provider.
func (r *ResponseBuilder) BodyProvider(body BodyProvider) *ResponseBuilder {
	if body == nil {
		return r
	}
	r.bodyProvider = body

	ct := body.ContentType()
	if ct != "" {
		r.SetHeader(contentType, ct)
	}

	return r
}

func (r *ResponseBuilder) BodyJSON(bodyJSON interface{}) *ResponseBuilder {
	if bodyJSON == nil {
		return r
	}
	return r.BodyProvider(jsonBodyProvider{payload: bodyJSON})
}

func (r *ResponseBuilder) Build() *Response {
	var body []byte
	var err error
	if r.bodyProvider != nil {
		body, err = r.bodyProvider.Body()
		if err != nil {
			return &Response{StatusCode: 500, Headers: r.headers, Body: []byte("Invalid body return")}
		}
	}
	return &Response{StatusCode: r.statusCode, Headers: r.headers, Body: body}
}
