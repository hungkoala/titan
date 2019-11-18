package nats

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"

	oNats "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"gitlab.com/silenteer/go-nats/log"
	"logur.dev/logur"
)

var hostname string

func init() {
	var err error
	hostname, err = os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}
}

// Option is a function on the options for a connection.
type Option func(*Options) error

// Options can be used to create a customized connection.
type Options struct {
	Addr        string // TCP address to listen on, ":http" if empty
	Subject     string
	router      Router
	ReadTimeout time.Duration
	Logger      logur.Logger
	ServiceName string
}

func GetDefaultOptions() Options {
	r := chi.NewRouter()
	//r.Use(middleware.RequestID)
	//r.Use(middleware.Logger)
	//r.Use(middleware.Recoverer)
	r.Use(RouteParamsMiddleware)

	return Options{
		Addr:        oNats.DefaultURL,
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

func Routes(r func(r Router)) Option {
	return func(o *Options) error {
		r(o.router)
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

func ServiceName(name string) Option {
	return func(o *Options) error {
		o.ServiceName = name
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
		subject:     opts.Subject,
		handler:     opts.router,
		addr:        opts.Addr,
		readTimeout: opts.ReadTimeout,
		logger:      opts.Logger,
		serviceName: opts.ServiceName,
	}, nil
}

func NewServerAndStart(options ...Option) *Server {
	server, err := NewServer(options...)
	logger := log.DefaultLogger(nil)
	if err != nil {
		logger.Error(fmt.Sprintf("Nats server creation error: %+v\n ", err))
		os.Exit(1)
	}

	err = server.Start()
	if err != nil {
		logger.Error(fmt.Sprintf("Nats server start error: %+v\n ", err))
		os.Exit(1)
	}

	return server
}

func NewServerAndStartRoutine(options ...Option) *Server {
	server, err := NewServer(options...)
	logger := log.DefaultLogger(nil)
	if err != nil {
		logger.Error(fmt.Sprintf("Nats server creation error: %+v\n ", err))
		os.Exit(1)
	}

	// start routine
	go func() {
		err = server.Start()
		if err != nil {
			logger.Error(fmt.Sprintf("Nats server start error: %+v\n ", err))
			os.Exit(1)
		}
	}()
	return server
}

type Server struct {
	addr    string // TCP address to listen on, ":http" if empty
	subject string

	handler http.Handler // handler to invoke, http.DefaultServeMux if nil

	readTimeout time.Duration
	logger      logur.Logger
	stop        chan interface{}
	serviceName string
}

func (srv *Server) Start() error {

	if srv.handler == nil {
		return errors.New("nats: Handler not found")
	}

	if srv.subject == "" {
		return errors.New("nats: Subject can not be empty")
	}

	if srv.addr == "" {
		return errors.New("nats: Address can not be empty")
	}

	if srv.logger == nil {
		return errors.New("nats: Logger can not be empty")
	}

	if srv.readTimeout == 0 {
		return errors.New("nats: ReadTimeout can not be empty")
	}

	timeoutHandler := http.TimeoutHandler(srv.handler, srv.readTimeout, "nats handler  timeout")

	srv.logger.Info("Connecting to NATS Server at: ", map[string]interface{}{"add": srv.addr, "subject": srv.subject})
	conn, err := NewConnection(srv.addr)
	if err != nil {
		return errors.WithMessage(err, "Nats connection error ")
	}

	subscription, err := srv.serve(conn.Conn, srv.logger, srv.subject, timeoutHandler)
	if err != nil {
		return errors.WithMessage(err, "Nats serve error ")
	}

	// Handle SIGINT and SIGTERM.
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	srv.stop = make(chan interface{}, 1)

	CleanUp := func() {
		srv.stop = nil
		srv.logger.Info("Nats server is being closed")
		_ = subscription.Unsubscribe()
		conn.Conn.Close()
	}

	select {
	case <-srv.stop:
		CleanUp()
	case <-done:
		CleanUp()
	}

	srv.logger.Info("Nats server is down now")

	return nil
}

func (srv *Server) Stop() {
	if srv.stop != nil {
		srv.stop <- "stop"
	}
}

func (srv *Server) serve(conn *oNats.EncodedConn, logger logur.Logger, subject string, handler http.Handler) (*oNats.Subscription, error) {
	return conn.Subscribe(subject, func(addr string, rpSubject string, rq *Request) {
		logger.Debug("Nats received message", map[string]interface{}{"url": rq.URL, "subject": subject})

		go func(enc *oNats.EncodedConn) {
			defer handlePanic(conn, logger, rpSubject)
			t1 := time.Now()

			rp := &Response{
				StatusCode: 200, // internal server error as default
				Status:     "",
				Headers:    http.Header{},
			}

			requestID := rq.Headers.Get(XRequestId)
			if requestID == "" {
				requestID = RandomString(6)
			}

			ctx := context.Background()

			// add log
			l := log.WithFields(logger, map[string]interface{}{XHostName: hostname, "Service": srv.serviceName, XRequestId: requestID})
			ctx = context.WithValue(ctx, XLoggerId, l)

			// add request id
			ctx = context.WithValue(ctx, XRequestId, requestID)

			rq, err := requestToHttpRequest(rq, ctx)
			if err != nil {
				replyError(enc, logger, err, rpSubject)
				return
			}

			defer func() {
				l.Info("Request complete", map[string]interface{}{
					"method":     rq.Method,
					"url":        rq.URL,
					"status":     rp.StatusCode,
					"elapsed_ms": float64(time.Since(t1).Nanoseconds()) / 1000000.0},
				)
			}()

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
			err = fmt.Errorf("panic : %v", r)
		}
		replyError(enc, logger, err, rpSubject)
		logger.Info("panic recovered")
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

func requestToHttpRequest(rq *Request, c context.Context) (*http.Request, error) {
	var body io.Reader
	if rq.Body != nil {
		body = bytes.NewReader(rq.Body)
	} else {
		body = bytes.NewReader([]byte{})
	}

	request, err := http.NewRequestWithContext(c, rq.Method, rq.URL, body)
	if err != nil {
		return nil, errors.WithMessage(err, "Nats: Something wrong with creating the request")
	}

	if rq.Headers != nil {
		request.Header = rq.Headers
	}

	return request, nil
}

func RouteParamsMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		queryParams := QueryParams(r.URL.Query())
		ctx := context.WithValue(r.Context(), XQueryParams, queryParams)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
