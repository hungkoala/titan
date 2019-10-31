package nats

// WithStatusCode is an error with status code
type WithStatusCode interface {
	StatusCode() int
}

// ErrorWithStatusCode implements both error and WithStatusCode
type ErrorWithStatusCode struct {
	error
	statusCode int
}
