package nats

import (
	"fmt"
	"net/http"

	"logur.dev/logur"

	"github.com/go-chi/chi"
)

type HandlerFunc func(r *http.Request) *Response

type RouteProvider interface {
	Routes(r Router) // side effect function
}

type Router interface {
	http.Handler
	MethodFunc(method, pattern string, h HandlerFunc)
	Get(pattern string, h HandlerFunc)
	Put(pattern string, h HandlerFunc)
	Post(pattern string, h HandlerFunc)
	Delete(pattern string, h HandlerFunc)
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

// implement router interface
func (m *Mux) Get(pattern string, h HandlerFunc) {
	m.MethodFunc("GET", pattern, h)
}

func (m *Mux) Put(pattern string, h HandlerFunc) {
	m.MethodFunc("PUT", pattern, h)
}

func (m *Mux) Post(pattern string, h HandlerFunc) {
	m.MethodFunc("POST", pattern, h)
}

func (m *Mux) Delete(pattern string, h HandlerFunc) {
	m.MethodFunc("DELETE", pattern, h)
}

func (m *Mux) MethodFunc(method, pattern string, handlerFunc HandlerFunc) {
	m.Router.MethodFunc(method, pattern, func(w http.ResponseWriter, r *http.Request) {
		rp := handlerFunc(r)

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
