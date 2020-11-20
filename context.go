package titan

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/go-chi/chi"

	"logur.dev/logur"
)

type QueryParams map[string][]string
type PathParams map[string]string

type Context struct {
	context            context.Context
	cachedUserInfoJson *String
	mux                sync.Mutex
}

func NewBackgroundContext() *Context {
	return NewContext(context.Background())
}

func NewContext(c context.Context) *Context {
	v := c.Value(XGlobalCache)
	if v != nil {
		return &Context{context: c}
	}
	return &Context{context: context.WithValue(c, XGlobalCache, &GlobalCache{Data: map[string]interface{}{}})}
}

func (c *Context) WithValue(key, val interface{}) *Context {
	ctx := context.WithValue(c, key, val)
	return &Context{context: ctx}
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
		logger = GetLogger()
	}
	return logger
}

func (c *Context) Request() *Request {
	request, _ := c.Value(XRequest).(*Request)
	return request
}

func (c *Context) RequestId() string {
	id, ok := c.Value(XRequestId).(string)
	if !ok {
		id = ""
	}
	return id
}

func (c *Context) Origin() string {
	id, ok := c.Value(XOrigin).(string)
	if !ok {
		id = ""
	}
	return id
}

func (c *Context) UberTraceID() string {
	id, ok := c.Value(UberTraceID).(string)
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

func (c *Context) GetPathParam(name string) string {
	return c.PathParams()[name]
}

func (c *Context) UserInfo() *UserInfo {
	userInfo, ok := c.Value(XUserInfo).(*UserInfo)
	if ok {
		return userInfo
	}
	return nil
}

func (c *Context) GlobalCache() *GlobalCache {
	globalCache, ok := c.Value(XGlobalCache).(*GlobalCache)
	if ok {
		return globalCache
	}
	return nil
}

func (c *Context) UserInfoJson() string {
	//if c.cachedUserInfoJson == nil {
	c.mux.Lock()
	useInfo := c.UserInfo()
	value := ""
	if useInfo != nil {
		b, err := json.Marshal(useInfo)
		if err == nil {
			value = string(b)
		}
	}
	c.cachedUserInfoJson = &String{Value: value}
	c.mux.Unlock()
	//}
	return c.cachedUserInfoJson.Value
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

// dangerously! only use this function after authentication
func (c *Context) LoginToCareProviderAsRoleMustBeUsedAfterAuthentication(careProviderId string, role Role) *Context {
	var userInfo UserInfo
	if c.UserInfo() == nil {
		userInfo = UserInfo{}
	} else {
		u := c.UserInfo()
		userInfo = UserInfo{
			ExternalUserId: u.ExternalUserId,
			UserId:         u.UserId,
			DeviceId:       u.DeviceId,
			Attributes:     u.Attributes,
		}
	}
	userInfo.CareProviderId = UUID(careProviderId)
	userInfo.CareProviderKey = ""
	userInfo.Role = role
	return c.WithValue(XUserInfo, &userInfo)
}

type GlobalCache struct {
	Data map[string]interface{}
}
