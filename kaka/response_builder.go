package kaka

import (
	"net/http"
)

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
