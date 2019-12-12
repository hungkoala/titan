package titan

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"logur.dev/logur"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
)

type Handler interface{}
type HandlerFunc func(*Context, *Request) *Response

type Router interface {
	http.Handler
	Register(method, pattern string, h HandlerFunc)
	RegisterJson(method, pattern string, h Handler)
}

type Mux struct {
	Router chi.Router
	Logger logur.Logger
}

func NewRouter(r chi.Router) *Mux {
	return &Mux{Router: r}
}

// implement http.Handler
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Router.ServeHTTP(w, r)
}

func (m *Mux) Register(method, pattern string, handlerFunc HandlerFunc) {
	m.Router.MethodFunc(method, pattern, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, XPathParams, ParsePathParams(ctx))

		// add request to context
		newRequest, err := httpRequestToRequest(r)
		if err != nil {
			m.Logger.Error(fmt.Sprintf("request coverting error: %+v\n ", err))
		}

		ctx = context.WithValue(ctx, XRequest, newRequest)

		// cal handler
		rp := handlerFunc(NewContext(ctx), newRequest)

		err = writeResponse(w, rp)
		if err != nil {
			m.Logger.Error(fmt.Sprintf("reposne writing error: %+v\n ", err))
		}
	})
}

func (m *Mux) RegisterJson(method, pattern string, h Handler) {
	m.Router.MethodFunc(method, pattern, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, XPathParams, ParsePathParams(ctx))

		// add request to context
		newRequest, err := httpRequestToRequest(r)
		if err != nil {
			m.Logger.Error(fmt.Sprintf("request coverting error: %+v\n ", err))
		}

		ctx = context.WithValue(ctx, XRequest, newRequest)

		rp := handleJsonRequest(NewContext(ctx), newRequest, h)
		err = writeResponse(w, rp)
		if err != nil {
			m.Logger.Error(fmt.Sprintf("json reposne writing error: %+v\n ", err))
		}
	})
}

func httpRequestToRequest(r *http.Request) (*Request, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.WithMessage(err, "Error reading body:")
	}

	defer func() { _ = r.Body.Close() }()
	if len(body) == 0 {
		body = nil
	}

	return &Request{
		Body:    body,
		URL:     r.RequestURI,
		Method:  r.Method,
		Headers: r.Header,
	}, nil
}

// side effect function
func writeResponse(w http.ResponseWriter, rp *Response) error {
	// write header
	for name, values := range rp.Header() {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// write status code
	if rp.StatusCode != 0 {
		w.WriteHeader(rp.StatusCode)
	}

	// write body
	if rp.Body != nil {
		_, err := w.Write(rp.Body)
		if err != nil {
			return errors.WithMessage(err, "Writing response error")
		}

	}

	return nil
}

func handleJsonRequest(c *Context, r *Request, cb Handler) *Response {
	logger := c.Logger()

	builder := NewResBuilder()

	//1. call function handler
	ret, err := callJsonHandler(c, r.Body, cb)
	if err != nil {
		logger.Error(fmt.Sprintf("Json handler error: %+v\n ", err))

		// see old code CommonExceptionHandler.java
		comErr, ok := err.(commonError)
		if ok {
			code, mess := comErr.CommonError()
			logger.Error(fmt.Sprintf("Common error: %s ", mess))
			return builder.
				StatusCode(400). //bad request
				BodyJSON(&JsonError{
					Message: code,
					Links:   map[string][]string{"self": {r.URL}},
				}).
				Build()
		}

		// error is a server response type
		if serverError, ok := err.(*ServerResponseError); ok {
			return builder.
				StatusCode(serverError.Status).
				Body(serverError.Body).
				Build()
		}

		return builder.
			StatusCode(500).
			BodyJSON(&DefaultJsonError{
				Message: "Json handler error:" + err.Error(),
				TraceId: c.RequestId(),
				Links:   map[string][]string{"self": {r.URL}},
			}).
			Build()
	}

	if ret == nil {
		return builder.
			StatusCode(200).
			Build()
	}

	switch v := ret.(type) {
	case string:
		return builder.Body([]byte(ret.(string))).Build()
	case int64, uint32, int, uint, float32, float64, bool:
		return builder.Body([]byte(fmt.Sprintf("%v", ret))).Build()
	case *Response:
		return ret.(*Response)
	default:
		_ = v
		//2. process result
		retJson, err := json.Marshal(ret)
		if err != nil {
			logger.Error(fmt.Sprintf("response json encoding error: %+v\n", err))
			return builder.
				StatusCode(500).
				BodyJSON(&DefaultJsonError{
					Message: "response json encoding error:" + err.Error(),
					TraceId: c.RequestId(),
					Links:   map[string][]string{"self": {r.URL}},
				}).
				Build()
		}
		return builder.
			Body(retJson).
			Build()
	}
}

//var emptyStringType = reflect.TypeOf("")
var emptyReqType = reflect.TypeOf(&Request{})

//var emptyResType = reflect.TypeOf(&Response{})
var emptyContextType = reflect.TypeOf(&Context{})
var errorType = reflect.TypeOf((*error)(nil)).Elem()
var handlerFormatError = errors.New("Handler needs to be a func \n `func(c *Context, interface{}) (interface{}, error)` or \n `func(c *Context) (interface{}, error)`")
var handlerExample = "\n Example: `func(c *Context, interface{}) (interface{}, error)` or \n `func(c *Context) (interface{}, error)`"

func callJsonHandler(c *Context, body []byte, cb interface{}) (interface{}, error) {
	if cb == nil {
		return nil, errors.New("nats: Handler is required")
	}
	cbType := reflect.TypeOf(cb)

	if cbType.Kind() != reflect.Func {
		return nil, handlerFormatError
	}

	numIn := cbType.NumIn()
	numOut := cbType.NumOut()

	if numIn == 0 || numIn > 2 {
		return nil, errors.New("Handler requires one or two parameters " + handlerExample)
	}

	if cbType.In(0) != emptyContextType {
		return nil, errors.New("Handler requires first parameter must be instance of nats.Context " + handlerExample)
	}

	if numOut == 0 || numOut > 2 {
		return nil, errors.New("Handler requires one or two return values " + handlerExample)
	}

	if cbType.Out(numOut-1) != errorType {
		return nil, errors.New("Handler requires second return value is an `error` " + handlerExample)
	}

	argType := cbType.In(numIn - 1)

	if argType == nil {
		return nil, errors.New("nats: Handler requires at least one argument")
	}

	cbValue := reflect.ValueOf(cb)
	oV := []reflect.Value{reflect.ValueOf(c)}

	if numIn == 2 {
		if len(body) == 0 {
			return nil, errors.New("Body is empty")
		}
		var oPtr reflect.Value
		if argType == emptyReqType {
			oPtr = reflect.ValueOf(body)
		} else {
			if argType.Kind() != reflect.Ptr {
				oPtr = reflect.New(argType)
			} else {
				oPtr = reflect.New(argType.Elem())
			}
			if err := decode(body, oPtr.Interface()); err != nil {
				return nil, err
			}
			if argType.Kind() != reflect.Ptr {
				oPtr = reflect.Indirect(oPtr)
			}
		}
		oV = append(oV, oPtr)
	}

	res := cbValue.Call(oV)

	if numOut == 2 {
		ret := res[0].Interface()
		var err error
		if v := res[1].Interface(); v != nil {
			err = v.(error)
		}
		return ret, err
	} else {
		var err error
		if v := res[0].Interface(); v != nil {
			err = v.(error)
		}
		return nil, err
	}
}

// Decode
func decode(data []byte, vPtr interface{}) (err error) {
	switch arg := vPtr.(type) {
	case *string:
		// If they want a string and it is a JSON string, strip quotes
		// This allows someone to send a struct but receive as a plain string
		// This cast should be efficient for Go 1.3 and beyond.
		str := string(data)
		if strings.HasPrefix(str, `"`) && strings.HasSuffix(str, `"`) {
			*arg = str[1 : len(str)-1]
		} else {
			*arg = str
		}
	case *[]byte:
		*arg = data
	default:
		err = json.Unmarshal(data, arg)
	}
	return
}
