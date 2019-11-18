package nats

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
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
func NewReqBuilder() *RequestBuilder {
	rq := &RequestBuilder{
		method:  "GET",
		headers: make(http.Header),
	}
	rq.SetContentType(jsonContentType)
	return rq
}

// Method

// Head sets the Request method to HEAD and sets the given pathURL.
func (r *RequestBuilder) Head(pathURL string) *RequestBuilder {
	r.method = "HEAD"
	return r.Url(pathURL)
}

// Get sets the Request method to GET and sets the given pathURL.
func (r *RequestBuilder) Get(pathURL string) *RequestBuilder {
	r.method = "GET"
	return r.Url(pathURL)
}

// Post sets the Request method to POST and sets the given pathURL.
func (r *RequestBuilder) Post(pathURL string) *RequestBuilder {
	r.method = "POST"
	return r.Url(pathURL)
}

// Put sets the Request method to PUT and sets the given pathURL.
func (r *RequestBuilder) Put(pathURL string) *RequestBuilder {
	r.method = "PUT"
	return r.Url(pathURL)
}

// Patch sets the Request method to PATCH and sets the given pathURL.
func (r *RequestBuilder) Patch(pathURL string) *RequestBuilder {
	r.method = "PATCH"
	return r.Url(pathURL)
}

// Delete sets the Request method to DELETE and sets the given pathURL.
func (r *RequestBuilder) Delete(pathURL string) *RequestBuilder {
	r.method = "DELETE"
	return r.Url(pathURL)
}

// Options sets the Request method to OPTIONS and sets the given pathURL.
func (r *RequestBuilder) Options(pathURL string) *RequestBuilder {
	r.method = "OPTIONS"
	return r.Url(pathURL)
}

// Trace sets the Request method to TRACE and sets the given pathURL.
func (r *RequestBuilder) Trace(pathURL string) *RequestBuilder {
	r.method = "TRACE"
	return r.Url(pathURL)
}

// Connect sets the Request method to CONNECT and sets the given pathURL.
func (r *RequestBuilder) Connect(pathURL string) *RequestBuilder {
	r.method = "CONNECT"
	return r.Url(pathURL)
}

// Url extends the rawURL with the given path by resolving the reference to
// an absolute URL. If parsing errors occur, the rawURL is left unmodified.
func (r *RequestBuilder) Url(url string) *RequestBuilder {
	r.rawURL = url
	return r
}

// Header

// Add adds the key, value pair in Headers, appending values for existing keys
// to the key's values. Header keys are canonicalized.
func (r *RequestBuilder) AddHeader(key, value string) *RequestBuilder {
	r.headers.Add(key, value)
	return r
}

// Set sets the key, value pair in Headers, replacing existing values
// associated with key. Header keys are canonicalized.
func (r *RequestBuilder) SetHeader(key, value string) *RequestBuilder {
	r.headers.Set(key, value)
	return r
}

func (r *RequestBuilder) SetContentType(contentType string) {
	r.SetHeader(contentType, contentType)
}

func (r *RequestBuilder) Body(body []byte) *RequestBuilder {
	if body == nil {
		return r
	}
	return r.BodyProvider(byteBodyProvider{body: body})
}

// BodyProvider sets the RequestBuilder's body provider.
func (r *RequestBuilder) BodyProvider(body BodyProvider) *RequestBuilder {
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

func (r *RequestBuilder) BodyJSON(bodyJSON interface{}) *RequestBuilder {
	if bodyJSON == nil {
		return r
	}
	return r.BodyProvider(jsonBodyProvider{payload: bodyJSON})
}

func (r *RequestBuilder) AddHeaders(header http.Header) *RequestBuilder {
	for key, values := range header {
		for _, value := range values {
			r.AddHeader(key, value)
		}
	}
	return r
}

func (r *RequestBuilder) Subject(subject string) *RequestBuilder {
	r.subject = subject
	return r
}

func (r *RequestBuilder) Build() (*Request, error) {
	_, err := url.Parse(r.rawURL)
	if err != nil {
		return nil, errors.New("invalid url " + r.rawURL)
	}
	var body []byte
	if r.bodyProvider != nil {
		body, err = r.bodyProvider.Body()
		if err != nil {
			return nil, errors.WithMessage(err, "Invalid body format ")
		}
	}
	return &Request{URL: r.rawURL, Method: r.method, Headers: r.headers, Body: body, Subject: r.subject}, nil
}