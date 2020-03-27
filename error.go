package titan

import (
	"fmt"
	"github.com/go-playground/validator"
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
	ValidationErrors []ValidationError `json:"validationErrors"`
	ServerError      string            `json:"serverError"`
}

type ValidationError struct {
	Namespace string      `json:"namespace"`
	Field     string      `json:"field"`
	Rule      string      `json:"rule"`
	Value     interface{} `json:"value"`
	Param     string      `json:"param"`
}

// mapped from HttpClientResponseException.java
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

// see CommonException.java
type CommonException struct {
	Message     string
	ServerError string
	*stack
}

func NewCommonException(serverError string) *CommonException {
	return &CommonException{
		ServerError: serverError,
		stack:       callers(),
	}
}

func (e *CommonException) Error() string {
	return fmt.Sprintf("Exception: Message '%s', ServerError '%s'", e.Message, e.ServerError)
}

// RecordDeleteFailedException.java
type RecordDeleteFailedException struct {
	*CommonException
}

func NewRecordDeleteFailedException(entityType string, id UUID, serverError string) *RecordDeleteFailedException {
	message := fmt.Sprintf("%s record with id %s doesn't exists or deleted", entityType, id)
	return &RecordDeleteFailedException{
		CommonException: &CommonException{
			ServerError: serverError,
			Message:     message,
			stack:       callers(),
		},
	}
}

//RecordNotFoundException.java
type RecordNotFoundException struct {
	*CommonException
}

func NewRecordNotFoundException(entityType, id, serverError string) *RecordNotFoundException {
	message := fmt.Sprintf("%s record with id %s doesn't exists or deleted", entityType, id)
	return &RecordNotFoundException{
		CommonException: &CommonException{
			ServerError: serverError,
			Message:     message,
			stack:       callers(),
		},
	}
}

//// --------------- ServerResponseError -------
type ServerResponseError struct {
	Status  int
	Body    []byte
	Headers http.Header
}

func (s *ServerResponseError) Error() string {
	return fmt.Sprintf("Server Response Error status %d", s.Status)
}

type causer interface {
	Cause() error
}

func UnwrapErr(err error) error {
Loop:
	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break Loop
		}
		err = cause.Cause()
		switch err.(type) {
		case *CommonException,
			*ClientResponseError,
			*validator.InvalidValidationError,
			validator.ValidationErrors,
			*validator.ValidationErrors:
			break Loop
		}
	}
	return err
}
