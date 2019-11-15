package nats

import (
	"context"
	"time"

	"logur.dev/logur"
)

type Context struct {
	context context.Context
	logger  logur.Logger
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
	return c.logger
}
