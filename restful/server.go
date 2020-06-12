package restful

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/middleware"

	"gitlab.com/silenteer-oss/titan"

	"github.com/go-chi/chi"

	"github.com/pkg/errors"
	"gitlab.com/silenteer-oss/titan/log"
	"logur.dev/logur"
)

// Option is a function on the options for a connection.
type Option func(*Options) error

// Options can be used to create a customized connection.
type Options struct {
	logger logur.Logger
	queue  string
	//config            *NatsConfig
	router titan.Router
	//messageSubscriber *MessageSubscriber
}

func Logger(logger logur.Logger) Option {
	return func(o *Options) error {
		o.logger = logger
		return nil
	}
}

func Queue(queue string) Option {
	return func(o *Options) error {
		o.queue = queue
		return nil
	}
}

//
//func Config(config *NatsConfig) Option {
//	return func(o *Options) error {
//		o.config = config
//		return nil
//	}
//}

func Routes(r func(titan.Router)) Option {
	return func(o *Options) error {
		r(o.router)
		return nil
	}
}

func NewServer(port string, options ...Option) *Server {
	// default logger
	logger := log.WithFields(titan.GetLogger(), map[string]interface{}{"port": port})

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(titan.NewMiddleware("Http", logger))
	//r.Use(middleware.Timeout(60 * time.Second))

	// set default handlers - health check and build info
	defaultHandlers := &titan.DefaultHandlers{Subject: ""}
	withDefaultOptions := append(options, Routes(defaultHandlers.Routes))

	// default options
	opts := Options{
		logger: logger,
		//config:            GetNatsConfig(),
		router: titan.NewRouter(r),
		queue:  "workers",
		//messageSubscriber: NewMessageSubscriber(logger),
	}

	// merge options with user define
	for _, opt := range withDefaultOptions {
		if opt != nil {
			if err := opt(&opts); err != nil {
				logger.Error(fmt.Sprintf("Nats server creation error: %+v\n ", err))
				os.Exit(1)
			}
		}
	}

	return &Server{
		port: port,
		//config:            opts.config,
		handler: opts.router,
		//messageSubscriber: opts.messageSubscriber,
		logger: log.WithFields(opts.logger, map[string]interface{}{"queue": opts.queue}),
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
	port string
	//config            *NatsConfig
	handler http.Handler // handler to invoke, http.DefaultServeMux if nil
	//messageSubscriber *MessageSubscriber
	logger logur.Logger
	stop   chan interface{}
}

func (srv *Server) start(started ...chan interface{}) (err error) {

	if srv.handler == nil {
		return errors.New("Handler not found")
	}

	if srv.port == "" {
		return errors.New("Port not found")
	}

	//config := srv.config
	//if config == nil {
	//	return errors.New("nats: NatsConfig can not be nil")
	//}

	if srv.logger == nil {
		return errors.New("Logger can not be empty")
	}

	h := &http.Server{
		Addr:    ":" + srv.port,
		Handler: srv.handler,
	}

	go func() {
		srv.logger.Info("Https server started")
		for i := range started {
			started[i] <- true
		}
		if err = h.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

	srv.logger.Info("Https server stopped")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer func() {
		cancel()
	}()

	if err = h.Shutdown(ctxShutDown); err != nil {
		srv.logger.Error(fmt.Sprintf("server Shutdown Failed:%+s", err))
	}

	srv.logger.Info("Https server exited properly")

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
