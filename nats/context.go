package nats

import (
	"context"
	"time"

	"github.com/go-chi/chi"

	"gitlab.com/silenteer/go-nats/log"

	"logur.dev/logur"
)

type QueryParams map[string][]string
type PathParams map[string]string

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

func (c *Context) QueryParams() QueryParams {
	requestParams, ok := c.Value(XQueryParams).(QueryParams)
	if !ok {
		requestParams = QueryParams{}
	}
	return requestParams
}

func (c *Context) PathParams() PathParams {
	pathParams, ok := c.Value(XPathParams).(PathParams)
	if !ok {
		pathParams = PathParams{}
	}
	return pathParams
}

func ParsePathParams(ctx context.Context) PathParams {
	oParams := chi.RouteContext(ctx).URLParams
	rParams := PathParams{}
	if oParams.Keys != nil {
		for i, k := range oParams.Keys {
			if oParams.Values != nil && len(oParams.Values) > i {
				rParams[k] = oParams.Values[i]
			} else {
				rParams[k] = ""
			}
		}
	}
	return rParams
}
