package kaka

import (
	"bytes"
	"encoding/json"
)

// BodyProvider provides Body content for http.Request attachment.
type BodyProvider interface {
	// ContentType returns the Content-Type of the body.
	ContentType() string
	// Body returns the io.Reader body.
	Body() ([]byte, error)
}

// bodyProvider provides the wrapped body value as a Body for reqests.
type byteBodyProvider struct {
	body []byte
}

func (p byteBodyProvider) ContentType() string {
	return ""
}

func (p byteBodyProvider) Body() ([]byte, error) {
	return p.body, nil
}

type jsonBodyProvider struct {
	payload interface{}
}

func (p jsonBodyProvider) ContentType() string {
	return jsonContentType
}

func (p jsonBodyProvider) Body() ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(p.payload)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
