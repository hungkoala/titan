package nats

import (
	"errors"
	"net/http"
	"net/url"
)

// Request is a simple struct
type RequestBuilder struct {
	// HTTP method (GET, POST, etc.)
	method string

	headers http.Header

	// body provider
	bodyProvider BodyProvider

	// raw url string for requests
	rawURL  string
	subject string
}

// New returns a new default  Request.
func NewRequestBuilder() *RequestBuilder {
	rq := &RequestBuilder{
		method:  "GET",
		headers: make(http.Header),
	}
	rq.SetContentType(jsonContentType)
	return rq
}

// Method

// Head sets the Request method to HEAD and sets the given pathURL.
func (s *RequestBuilder) Head(pathURL string) *RequestBuilder {
	s.method = "HEAD"
	return s.Url(pathURL)
}

// Get sets the Request method to GET and sets the given pathURL.
func (s *RequestBuilder) Get(pathURL string) *RequestBuilder {
	s.method = "GET"
	return s.Url(pathURL)
}

// Post sets the Request method to POST and sets the given pathURL.
func (s *RequestBuilder) Post(pathURL string) *RequestBuilder {
	s.method = "POST"
	return s.Url(pathURL)
}

// Put sets the Request method to PUT and sets the given pathURL.
func (s *RequestBuilder) Put(pathURL string) *RequestBuilder {
	s.method = "PUT"
	return s.Url(pathURL)
}

// Patch sets the Request method to PATCH and sets the given pathURL.
func (s *RequestBuilder) Patch(pathURL string) *RequestBuilder {
	s.method = "PATCH"
	return s.Url(pathURL)
}

// Delete sets the Request method to DELETE and sets the given pathURL.
func (s *RequestBuilder) Delete(pathURL string) *RequestBuilder {
	s.method = "DELETE"
	return s.Url(pathURL)
}

// Options sets the Request method to OPTIONS and sets the given pathURL.
func (s *RequestBuilder) Options(pathURL string) *RequestBuilder {
	s.method = "OPTIONS"
	return s.Url(pathURL)
}

// Trace sets the Request method to TRACE and sets the given pathURL.
func (s *RequestBuilder) Trace(pathURL string) *RequestBuilder {
	s.method = "TRACE"
	return s.Url(pathURL)
}

// Connect sets the Request method to CONNECT and sets the given pathURL.
func (s *RequestBuilder) Connect(pathURL string) *RequestBuilder {
	s.method = "CONNECT"
	return s.Url(pathURL)
}

// Url extends the rawURL with the given path by resolving the reference to
// an absolute URL. If parsing errors occur, the rawURL is left unmodified.
func (s *RequestBuilder) Url(url string) *RequestBuilder {
	s.rawURL = url
	return s
}

// Header

// Add adds the key, value pair in Headers, appending values for existing keys
// to the key's values. Header keys are canonicalized.
func (s *RequestBuilder) AddHeader(key, value string) *RequestBuilder {
	s.headers.Add(key, value)
	return s
}

// Set sets the key, value pair in Headers, replacing existing values
// associated with key. Header keys are canonicalized.
func (s *RequestBuilder) SetHeader(key, value string) *RequestBuilder {
	s.headers.Set(key, value)
	return s
}

func (r *RequestBuilder) SetContentType(contentType string) {
	r.SetHeader(contentType, contentType)
}

func (s *RequestBuilder) Body(body []byte) *RequestBuilder {
	if body == nil {
		return s
	}
	return s.BodyProvider(byteBodyProvider{body: body})
}

// BodyProvider sets the RequestBuilder's body provider.
func (s *RequestBuilder) BodyProvider(body BodyProvider) *RequestBuilder {
	if body == nil {
		return s
	}
	s.bodyProvider = body

	ct := body.ContentType()
	if ct != "" {
		s.SetHeader(contentType, ct)
	}

	return s
}

func (s *RequestBuilder) BodyJSON(bodyJSON interface{}) *RequestBuilder {
	if bodyJSON == nil {
		return s
	}
	return s.BodyProvider(jsonBodyProvider{payload: bodyJSON})
}

func (s *RequestBuilder) AddHeaders(header http.Header) *RequestBuilder {
	for key, values := range header {
		for _, value := range values {
			s.AddHeader(key, value)
		}
	}
	return s
}

func (s *RequestBuilder) Subject(subject string) *RequestBuilder {
	s.subject = subject
	return s
}

func (s *RequestBuilder) Build() (*Request, error) {
	_, err := url.Parse(s.rawURL)
	if err != nil {
		return nil, errors.New("invalid url " + s.rawURL)
	}
	var body []byte
	if s.bodyProvider != nil {
		body, err = s.bodyProvider.Body()
		if err != nil {
			return nil, errors.New("Invalid body  " + err.Error())
		}
	}
	return &Request{URL: s.rawURL, Method: s.method, Headers: s.headers, Body: body, Subject: s.subject}, nil
}
