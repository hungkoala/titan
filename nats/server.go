package nats

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/middleware"

	"github.com/go-chi/chi"

	oNats "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"gitlab.com/silenteer/go-nats/log"
	"logur.dev/logur"
)

// Option is a function on the options for a connection.
type Option func(*Options) error

// Options can be used to create a customized connection.
type Options struct {
	Addr        string // TCP address to listen on, ":http" if empty
	Subject     string
	router      Router
	ReadTimeout time.Duration
	Logger      logur.Logger
}

func GetDefaultOptions() Options {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	return Options{
		Addr:        oNats.DefaultURL,
		Subject:     "test",
		router:      NewRouter(r),
		ReadTimeout: 3 * time.Second,
		Logger:      log.DefaultLogger(nil),
	}
}

func Subject(subject string) Option {
	return func(o *Options) error {
		o.Subject = subject
		return nil
	}
}

func Address(address string) Option {
	return func(o *Options) error {
		o.Addr = address
		return nil
	}
}

func RouterProvider(r RouteProvider) Option {
	return func(o *Options) error {
		r.Routes(o.router)
		return nil
	}
}

func ReadTimeout(timeout time.Duration) Option {
	return func(o *Options) error {
		o.ReadTimeout = timeout
		return nil
	}
}

func Logger(logger logur.Logger) Option {
	return func(o *Options) error {
		o.Logger = logger
		return nil
	}
}

func NewServer(options ...Option) (*Server, error) {
	opts := GetDefaultOptions()
	for _, opt := range options {
		if opt != nil {
			if err := opt(&opts); err != nil {
				return nil, err
			}
		}
	}
	return &Server{
		Subject:     opts.Subject,
		Handler:     opts.router,
		Addr:        opts.Addr,
		ReadTimeout: opts.ReadTimeout,
		Logger:      opts.Logger,
	}, nil
}

type Server struct {
	Addr    string // TCP address to listen on, ":http" if empty
	Subject string

	Handler http.Handler // handler to invoke, http.DefaultServeMux if nil

	ReadTimeout time.Duration
	Logger      logur.Logger
}

func (srv *Server) Start() error {

	if srv.Handler == nil {
		return errors.New("nats: Handler not found")
	}

	if srv.Subject == "" {
		return errors.New("nats: Subject can not be empty")
	}

	if srv.Addr == "" {
		return errors.New("nats: Address can not be empty")
	}

	if srv.Logger == nil {
		return errors.New("nats: Logger can not be empty")
	}

	if srv.ReadTimeout == 0 {
		return errors.New("nats: ReadTimeout can not be empty")
	}

	timeoutHandler := http.TimeoutHandler(srv.Handler, srv.ReadTimeout, "nats handler  timeout")

	srv.Logger.Info("Connecting to NATS Server at: ", map[string]interface{}{"add": srv.Addr, "subject": srv.Subject})
	conn, err := NewConnection(srv.Addr)
	if err != nil {
		return errors.WithMessage(err, "Nats connection error ")
	}

	subscription, err := srv.serve(conn.Conn, srv.Logger, srv.Subject, timeoutHandler)
	if err != nil {
		return errors.WithMessage(err, "Nats serve error ")
	}

	// Handle SIGINT and SIGTERM.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	srv.Logger.Info("Nats server is being closed")
	_ = subscription.Unsubscribe()
	conn.Conn.Close()

	return nil
}

func (srv *Server) serve(conn *oNats.EncodedConn, logger logur.Logger, subject string, handler http.Handler) (*oNats.Subscription, error) {
	return conn.Subscribe(subject, func(addr string, rpSubject string, rq *Request) {
		logger.Debug("Nats received message", map[string]interface{}{"url": rq.URL, "subject": subject})

		go func(enc *oNats.EncodedConn) {
			defer handlePanic(conn, logger, rpSubject)

			rp := &Response{
				StatusCode: 200, // internal server error as default
				Status:     "",
				Headers:    http.Header{},
			}

			rq, err := requestToHttpRequest(rq)
			if err != nil {
				replyError(enc, logger, err, rpSubject)
				return
			}

			handler.ServeHTTP(rp, rq)

			err = enc.Publish(rpSubject, rp)
			if err != nil {
				logger.Error(fmt.Sprintf("Nats error on publish result back: %+v\n ", err))
			}

		}(conn)
	})
}

func handlePanic(enc *oNats.EncodedConn, logger logur.Logger, rpSubject string) {
	if r := recover(); r != nil {
		var ok bool
		var err error
		err, ok = r.(error)
		if !ok {
			err = fmt.Errorf("pkg: %v", r)
		}
		replyError(enc, logger, err, rpSubject)
	}
}

func replyError(enc *oNats.EncodedConn, logger logur.Logger, err error, rpSubject string) {
	logger.Error(fmt.Sprintf("Nats error: : %+v\n ", err))
	resp := &Response{
		StatusCode: 500, // internal server error as default
		Status:     "",
		Headers:    http.Header{},
	}
	er := enc.Publish(rpSubject, resp)
	if er != nil {
		logger.Error(fmt.Sprintf("Nats error on reply back: %+v\n ", er))
	}
}

func requestToHttpRequest(rq *Request) (*http.Request, error) {
	var body io.Reader
	if rq.Body != nil {
		body = bytes.NewReader(rq.Body)
	}
	request, err := http.NewRequest(rq.Method, rq.URL, body)
	if err != nil {
		return nil, errors.WithMessage(err, "Nats: Something wrong with creating the request")
	}

	if rq.Headers != nil {
		request.Header = rq.Headers
	}
	request.Header.Add("Connection", "close")

	return request, nil
}
