package nats

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

type HandlerFunc func(*Context, *Request) *Response

//type HandlerJsonFunc func(*Context, interface{}) (interface{}, error)

type RouteProvider interface {
	Routes(r Router) // side effect function
}

type Router interface {
	http.Handler
	Register(method, pattern string, h HandlerFunc)
	RegisterJson(method, pattern string, h interface{})
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

		newRequest, err := httpRequestToRequest(r)
		if err != nil {
			m.Logger.Error(fmt.Sprintf("request coverting error: %+v\n ", err))
		}

		// cal handler
		rp := handlerFunc(NewContext(ctx), newRequest)

		err = writeResponse(w, rp)
		if err != nil {
			m.Logger.Error(fmt.Sprintf("reposne writing error: %+v\n ", err))
		}
	})
}

func (m *Mux) RegisterJson(method, pattern string, handlerFunc interface{}) {
	m.Router.MethodFunc(method, pattern, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, XPathParams, ParsePathParams(ctx))

		rp := handleJsonRequest(NewContext(ctx), r, handlerFunc)
		err := writeResponse(w, rp)
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

	// write body
	if rp.Body != nil {
		_, err := w.Write(rp.Body)
		if err != nil {
			return errors.WithMessage(err, "Writing response error")
		}

	}
	// write status code
	if rp.StatusCode != 0 {
		w.WriteHeader(rp.StatusCode)
	}
	return nil
}

// Dissect the cb Handler's signature
func argInfo(cb interface{}) (reflect.Type, int) {
	cbType := reflect.TypeOf(cb)
	if cbType.Kind() != reflect.Func {
		panic("nats: Handler needs to be a func")
	}
	numArgs := cbType.NumIn()
	if numArgs == 0 {
		return nil, numArgs
	}
	return cbType.In(numArgs - 1), numArgs
}

var emptyMsgType = reflect.TypeOf(&Request{})

func handleJsonRequest(c *Context, r *http.Request, cb interface{}) *Response {
	logger := c.Logger()

	builder := NewResBuilder()

	//1. body to json
	body, err := extractJsonBody(r)
	if err != nil {
		logger.Error(fmt.Sprintf("Body parsing error: %+v\n ", err))
		return builder.
			StatusCode(400).
			BodyJSON(&DefaultJsonError{
				Message: "Body parsing error:" + err.Error(),
				TraceId: c.RequestId(),
				Links:   map[string][]string{"self": {r.URL.String()}},
			}).
			Build()
	}

	//2. call function handler
	ret, err := callJsonHandler(c, body, cb)
	if err != nil {
		logger.Error(fmt.Sprintf("Json handler error: %+v\n ", err))
		return builder.
			StatusCode(500).
			BodyJSON(&DefaultJsonError{
				Message: "Json handler error:" + err.Error(),
				TraceId: c.RequestId(),
				Links:   map[string][]string{"self": {r.URL.String()}},
			}).
			Build()
	}

	//3. process result
	retJson, err := json.Marshal(ret)
	if err != nil {
		logger.Error(fmt.Sprintf("response json encoding error: %+v\n ", err))
		return builder.
			StatusCode(500).
			BodyJSON(&DefaultJsonError{
				Message: "response json encoding error:" + err.Error(),
				TraceId: c.RequestId(),
				Links:   map[string][]string{"self": {r.URL.String()}},
			}).
			Build()
	}
	return builder.
		Body(retJson).
		Build()
}

func extractJsonBody(r *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.WithMessage(err, "Body reading error:")
	}

	defer func() { _ = r.Body.Close() }()
	if len(body) == 0 {
		return nil, errors.New("Body is empty error")
	}
	return body, nil
}

func callJsonHandler(c *Context, body []byte, cb interface{}) (interface{}, error) {
	if cb == nil {
		return nil, errors.New("nats: Handler required for EncodedConn Subscription")
	}

	argType, numArgs := argInfo(cb)

	if numArgs != 2 {
		return nil, errors.New("nats: Handler requires 2 arguments")
	}

	if argType == nil {
		return nil, errors.New("nats: Handler requires at least one argument")
	}

	cbValue := reflect.ValueOf(cb)

	wantsRaw := argType == emptyMsgType

	var oV []reflect.Value
	var oPtr reflect.Value
	if wantsRaw {
		fmt.Println("Want raw.....")
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

	cV := reflect.ValueOf(c)
	oV = []reflect.Value{cV, oPtr}

	res := cbValue.Call(oV)
	ret := res[0].Interface()
	var err error
	if v := res[1].Interface(); v != nil {
		err = v.(error)
	}

	return ret, err
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
