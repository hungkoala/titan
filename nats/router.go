package nats

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"logur.dev/logur"

	"github.com/go-chi/chi"
)

type HandlerFunc func(*Context, *Request) *Response

type RouteProvider interface {
	Routes(r Router) // side effect function
}

type Router interface {
	http.Handler
	Register(method, pattern string, h HandlerFunc)
	//RegisterJson(method, pattern string, h HandlerFunc)
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

		newResponse, _ := httpRequestToRequest(r)

		rp := handlerFunc(NewContext(ctx), newResponse)

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
				m.Logger.Error(fmt.Sprintf("Writing response error: %+v\n ", err))
			}
			return
		}

		if rp.StatusCode != 0 {
			w.WriteHeader(rp.StatusCode)
		}
	})
}

func httpRequestToRequest(r *http.Request) (*Request, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		return nil, err
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
