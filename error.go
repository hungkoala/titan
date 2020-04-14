package titan

import (
	"fmt"
	"github.com/go-playground/validator"
	"strings"
)

/**
 * A class that can be used to represent JSON errors that complies to Vnd.Error without the content type requirements.
 * see RFC 2119.
 */
// make it compatible old micronaut code, it's a subject to change latter
// see HttpClientResponseExceptionHandler.java
type DefaultJsonError struct {
	Message  string                   `json:"message"` // deprecated, wait for client migration, it will be deleted
	LogRef   string                   `json:"logref"`
	Path     string                   `json:"path"`
	Links    map[string][]string      `json:"_links"`
	Embedded map[string][]interface{} `json:"_embedded"`

	TraceId          string            `json:"traceId"`
	ValidationErrors []ValidationError `json:"validationErrors"`
	ServerError      string            `json:"serverError"` // deprecated, wait for client migration, it will be deleted

	// new errors message that supports multiple error messages
	Messages []ErrorMessage `json:"messages"`
}

type ValidationError struct {
	Namespace string      `json:"namespace"`
	Field     string      `json:"field"`
	Rule      string      `json:"rule"`
	Value     interface{} `json:"value"`
	Param     string      `json:"param"`
}

type ErrorMessage struct {
	Key   string      `json:"key"`
	Param interface{} `json:"param"`
}

func (e *ErrorMessage) String() string {
	return fmt.Sprintf("%s,%s ", e.Key, e.Param)
}

type IServerError interface {
	Error() string
	GetMessages() []ErrorMessage
}

type ServerError struct {
	Messages []ErrorMessage `json:"messages"`
}

func (s *ServerError) Error() string {
	var messages []string
	for _, m := range s.Messages {
		messages = append(messages, m.String())
	}
	return strings.Join(messages, ", ")
}

func (s *ServerError) GetMessages() []ErrorMessage {
	return s.Messages
}

func NewServerError(key string, param interface{}) *ServerError {
	message:=ErrorMessage{Key:key, Param: param}
	errors  :=  []ErrorMessage{message}
	return &ServerError{errors}
}

//----------------------------------------------------------------------------------------------
// mapped from HttpClientResponseException.java
// Error when invoke another microservices
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

//----------------------------------------------------------------------------------------------
// ----------- copy from old micronaut  infrastructure.exception
// see CommonException.java
// Deprecated: use ServerError instead).
type CommonException struct {
	Status      int // http status
	Message     string
	ServerError string
}

func NewCommonException(serverError string) *CommonException {
	return &CommonException{
		ServerError: serverError,
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
		},
	}
}

//----------------------------------------------------------------------------------------------

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
			*validator.ValidationErrors,
			*ServerError, IServerError:
			break Loop
		}
	}
	return err
}
