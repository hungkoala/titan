package titan

import (
	"fmt"
	"net/http"
	"runtime"
)

/**
 * A class that can be used to represent JSON errors that complies to Vnd.Error without the content type requirements.
 * see RFC 2119.
 */
// make it compatible old micronaut code, it's a subject to change latter
type JsonError struct {
	Message  string                 `json:"message"`
	LogRef   string                 `json:"logref"`
	Path     string                 `json:"path"`
	Links    map[string][]string    `json:"_links"`
	Embedded map[string][]JsonError `json:"_embedded"`
}

// see HttpClientResponseExceptionHandler.java
type DefaultJsonError struct {
	Message  string                 `json:"message"`
	LogRef   string                 `json:"logref"`
	Path     string                 `json:"path"`
	Links    map[string][]string    `json:"_links"`
	Embedded map[string][]JsonError `json:"_embedded"`

	TraceId          string            `json:"traceId"`
	ValidationErrors map[string]string `json:"validationErrors"`
	ServerError      string            `json:"serverError"`
}

type ClientResponseError struct {
	Message  string
	Response *Response
	Cause    error
}

func (h *ClientResponseError) GetMessage() string {
	return h.Message
}

func (h *ClientResponseError) Error() string {
	return h.Message
}

// ----------- copy from old micronaut  infrastructure.exception
// stack represents a stack of program counters.
type stack []uintptr

func callers() *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var st stack = pcs[0:n]
	return &st
}

type commonError interface {
	CommonError() (string, string)
}

type CommonException struct {
	Message string
	Code    string
	*stack
}

func NewCommonException(code string) *CommonException {
	return &CommonException{
		Code:  code,
		stack: callers(),
	}
}

func (e *CommonException) Error() string {
	return fmt.Sprintf("Exception: message '%s', code '%s'", e.Message, e.Code)
}

func (e *CommonException) CommonError() (string, string) {
	return e.Code, e.Message
}

type RecordDeleteFailedException struct {
	*CommonException
}

func NewRecordDeleteFailedException(entityType string, id UUID, code string) *RecordDeleteFailedException {
	message := fmt.Sprintf("%s record with id %s doesn't exists or deleted", entityType, id)
	return &RecordDeleteFailedException{
		CommonException: &CommonException{
			Code:    code,
			Message: message,
			stack:   callers(),
		},
	}
}

type RecordNotFoundException struct {
	*CommonException
}

func NewRecordNotFoundException(entityType, id, code string) *RecordNotFoundException {
	message := fmt.Sprintf("%s record with id %s doesn't exists or deleted", entityType, id)
	return &RecordNotFoundException{
		CommonException: &CommonException{
			Code:    code,
			Message: message,
			stack:   callers(),
		},
	}
}

//// ---------------
type ServerResponseError struct {
	Status  int
	Body    []byte
	Headers http.Header
}

func (s *ServerResponseError) Error() string {
	return fmt.Sprintf("Server Response Error status %d", s.Status)
}
