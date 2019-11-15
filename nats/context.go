package nats

import (
	"context"
	"time"

	"gitlab.com/silenteer/go-nats/log"

	"logur.dev/logur"
)

type Context struct {
	context context.Context
}

func NewContext(c context.Context) *Context {
	return &Context{context: c}
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.context.Deadline()
}

func (c *Context) Err() error {
	return c.context.Err()
}

func (c *Context) Value(key interface{}) interface{} {
	return c.context.Value(key)
}

func (c *Context) Done() <-chan struct{} {
	return c.context.Done()
}

func (c *Context) Logger() logur.Logger {
	logger, ok := c.Value(XLoggerId).(logur.Logger)
	if !ok {
		logger = log.DefaultLogger(map[string]interface{}{})
	}
	return logger
}

func (c *Context) RequestId() string {
	id, ok := c.Value(XRequestId).(string)
	if !ok {
		id = ""
	}
	return id
}
