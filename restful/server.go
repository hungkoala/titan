package restful

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/silenteer-oss/titan/log"

	"github.com/go-chi/chi/middleware"

	"gitlab.com/silenteer-oss/titan"

	"github.com/go-chi/chi"

	"github.com/pkg/errors"
	"logur.dev/logur"
)

// Option is a function on the options for a connection.
type Option func(*Options) error

// Options can be used to create a customized connection.
type Options struct {
	logger logur.Logger

	router titan.Router

	tlsEnable bool

	// base64 encoding of DER format
	tlsKey string

	// base64 encoding of DER format
	tlsCert string

	port string
}

func Logger(logger logur.Logger) Option {
	return func(o *Options) error {
		o.logger = logger
		return nil
	}
}

func TlsEnable(v bool) Option {
	return func(o *Options) error {
		o.tlsEnable = v
		return nil
	}
}

func TlsKey(v string) Option {
	return func(o *Options) error {
		o.tlsKey = v
		return nil
	}
}

func TlsCert(v string) Option {
	return func(o *Options) error {
		o.tlsCert = v
		return nil
	}
}

func Port(v string) Option {
	return func(o *Options) error {
		o.port = v
		return nil
	}
}

func Routes(r func(titan.Router)) Option {
	return func(o *Options) error {
		r(o.router)
		return nil
	}
}

func NewServer(options ...Option) *Server {
	// default logger
	logger := getLogger(options)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(titan.NewMiddleware("Http", logger))
	//r.Use(middleware.Timeout(60 * time.Second))

	// set default handlers - health check and build info
	defaultHandlers := &titan.DefaultHandlers{Subject: ""}
	withDefaultOptions := append(append(getDefaultConfig(), options...), Routes(defaultHandlers.Routes))

	// default options
	opts := Options{
		logger: logger,
		router: titan.NewRouter(r),
	}

	// merge options with user define
	for _, f := range withDefaultOptions {
		if f != nil {
			if err := f(&opts); err != nil {
				logger.Error(fmt.Sprintf("Nats server creation error: %+v\n ", err))
				os.Exit(1)
			}
		}
	}

	return &Server{
		tlsEnable: opts.tlsEnable,
		tlsKey:    opts.tlsKey,
		tlsCert:   opts.tlsCert,
		port:      opts.port,
		handler:   opts.router,
		logger:    opts.logger,
	}
}

func (srv *Server) Start(started ...chan interface{}) {
	err := srv.start(started...)
	if err != nil {
		srv.logger.Error(fmt.Sprintf("Nats server start error: %+v\n ", err))
		os.Exit(1)
	}
}

type IServer interface {
	Stop()
	Start(started ...chan interface{})
}

type Server struct {
	tlsEnable bool
	// base64 encoding of DER format
	tlsKey string
	// base64 encoding of DER format
	tlsCert string
	port    string
	handler http.Handler // handler to invoke, http.DefaultServeMux if nil
	logger  logur.Logger
	stop    chan interface{}
}

func (srv *Server) start(started ...chan interface{}) (err error) {
	var server *http.Server
	var tlsConfig *tls.Config

	if srv.handler == nil {
		return errors.New("Handler not found")
	}

	if srv.logger == nil {
		return errors.New("Logger can not be empty")
	}

	if srv.tlsEnable {
		if srv.tlsKey == "" {
			return errors.New("tlsKey is missing")
		}
		if srv.tlsCert == "" {
			return errors.New("tlsCert is missing")
		}

		cert, err := tls.X509KeyPair([]byte(srv.tlsCert), []byte(srv.tlsKey))
		if err != nil {
			return errors.WithMessage(err, "cannot load key pair")
		}

		// Construct a tls.config
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			// Other options
		}
	}

	go func() {
		server = &http.Server{
			Handler: srv.handler,
			// Other options
		}

		if srv.port != "" {
			server.Addr = ":" + srv.port
		}

		for i := range started {
			started[i] <- true
		}

		srv.logger.Info("Http server started")

		if srv.tlsEnable {
			server.TLSConfig = tlsConfig
			err = server.ListenAndServeTLS("", "")
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			srv.logger.Error(fmt.Sprintf("listen:%+s\n", err))
		}
	}()

	// Handle SIGINT and SIGTERM.
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	srv.stop = make(chan interface{}, 1)

	// wait for signal
	select {
	case <-srv.stop:
	case <-done:
	}

	srv.logger.Info("Http server stopped")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer func() {
		cancel()
	}()

	if err = server.Shutdown(ctxShutDown); err != nil {
		srv.logger.Error(fmt.Sprintf("server Shutdown Failed:%+s", err))
	}

	srv.logger.Info("Http server exited properly")

	if err == http.ErrServerClosed {
		err = nil
	}

	return nil
}

func (srv *Server) Stop() {
	if srv != nil && srv.stop != nil {
		srv.stop <- "stop"
		close(srv.stop)
	}
}

func getLogger(options []Option) logur.Logger {
	var logger logur.Logger
	opts := Options{
		router: titan.NewRouter(chi.NewRouter()),
	}
	for _, f := range options {
		f(&opts)
	}
	if opts.logger != nil {
		logger = opts.logger
	} else {
		logger = titan.GetLogger()
	}

	return log.WithFields(logger, map[string]interface{}{"tlsEnable": opts.tlsEnable, "port": opts.port})
}
