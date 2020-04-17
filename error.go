package titan

import (
	"fmt"

	"github.com/go-playground/validator"
)

/**
 * A class that can be used to represent JSON errors that complies to Vnd.Error without the content type requirements.
 * see RFC 2119.
 */
// make it compatible old micronaut code, it's a subject to change latter
// see HttpClientResponseExceptionHandler.java
type DefaultJsonError struct {
	Message  string                   `json:"message"`
	LogRef   string                   `json:"logref"`
	Path     string                   `json:"path"`
	Links    map[string][]string      `json:"_links"`
	Embedded map[string][]interface{} `json:"_embedded"`

	TraceId          string            `json:"traceId"`
	ValidationErrors []ValidationError `json:"validationErrors"`
	ServerError      string            `json:"serverError"`
	ServerErrorParam interface{}       `json:"serverErrorParam"` // to support parameter message like "Beim Transformieren des Types {0} in den Type {1} ist ein Fehler aufgetreten."
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

//----------------------------------------------------------------------------------------------

type ServerError interface {
	GetServerError() string
	GetServerErrorParam() interface{}
}

// ----------- copy from old micronaut  infrastructure.exception
// see CommonException.java
// todo: really really want to change it from Exception to Error. Go does not use Exception. Will change it soon
type CommonException struct {
	Status           int // http status
	Message          string
	ServerError      string
	ServerErrorParam interface{}
}

func (c *CommonException) GetServerError() string {
	return c.ServerError
}

func (c *CommonException) GetServerErrorParam() interface{} {
	return c.ServerErrorParam
}

func (e *CommonException) Error() string {
	return fmt.Sprintf("Server error : Message '%s', ServerError '%s'", e.Message, e.ServerError)
}

func NewCommonException(serverError string, param ...interface{}) *CommonException {
	var p interface{}
	if len(param) > 0 {
		p = param[0]
	}
	return &CommonException{
		ServerError:      serverError,
		ServerErrorParam: p,
	}
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
		case *CommonException, ServerError,
			*ClientResponseError,
			*validator.InvalidValidationError,
			validator.ValidationErrors,
			*validator.ValidationErrors:
			break Loop
		}
	}
	return err
}
