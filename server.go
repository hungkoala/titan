package titan

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
	"gitlab.com/silenteer/titan/log"
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
	config      *Config
	router      Router
	readTimeout time.Duration
	logger      logur.Logger
}

func GetDefaultOptions() Options {
	r := chi.NewRouter()
	r.Use(RouteParamsMiddleware)

	return Options{
		config:      &Config{},
		router:      NewRouter(r),
		readTimeout: 15 * time.Second,
		logger:      log.DefaultLogger(nil),
	}
}

func SetConfig(config *Config) Option {
	return func(o *Options) error {
		o.config = config
		return nil
	}
}

func Routes(r func(Router)) Option {
	return func(o *Options) error {
		r(o.router)
		return nil
	}
}

func ReadTimeout(timeout time.Duration) Option {
	return func(o *Options) error {
		o.readTimeout = timeout
		return nil
	}
}

func Logger(logger logur.Logger) Option {
	return func(o *Options) error {
		o.logger = logger
		return nil
	}
}

func NewServer(options ...Option) *Server {
	opts := GetDefaultOptions()
	logger := log.DefaultLogger(nil)
	for _, opt := range options {
		if opt != nil {
			if err := opt(&opts); err != nil {
				logger.Error(fmt.Sprintf("Nats server creation error: %+v\n ", err))
				os.Exit(1)
			}
		}
	}
	return &Server{
		config:      opts.config,
		handler:     opts.router,
		readTimeout: opts.readTimeout,
		logger:      opts.logger,
	}
}

func (srv *Server) Start() {
	err := srv.start()
	if err != nil {
		srv.logger.Error(fmt.Sprintf("Nats server start error: %+v\n ", err))
		os.Exit(1)
	}
}

type Server struct {
	config      *Config
	handler     http.Handler // handler to invoke, http.DefaultServeMux if nil
	readTimeout time.Duration
	logger      logur.Logger
	stop        chan interface{}
}

func (srv *Server) start() error {

	if srv.handler == nil {
		return errors.New("nats: Handler not found")
	}

	config := srv.config
	if config == nil {
		return errors.New("nats: Config can not be nil")
	}

	if config.Subject == "" {
		return errors.New("nats: Subject can not be empty")
	}

	fmt.Println()
	if config.Servers == "" {
		return errors.New("nats: Address can not be empty")
	}

	if srv.logger == nil {
		return errors.New("nats: Logger can not be empty")
	}

	if srv.readTimeout == 0 {
		return errors.New("nats: ReadTimeout can not be empty")
	}

	timeoutHandler := http.TimeoutHandler(srv.handler, srv.readTimeout, "nats handler  timeout")

	srv.logger.Info("Connecting to NATS Server at: ", map[string]interface{}{"add": config.Servers, "subject": config.Subject})
	conn, err := NewConnection(config.Servers)
	if err != nil {
		return errors.WithMessage(err, "Nats connection error ")
	}

	subscription, err := subscribe(conn.Conn, srv.logger, config.Subject, config.Queue, timeoutHandler)
	if err != nil {
		return errors.WithMessage(err, "Nats serve error ")
	}

	// Handle SIGINT and SIGTERM.
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	srv.stop = make(chan interface{}, 1)

	cleanUp := func() {
		close(srv.stop)
		srv.stop = nil
		srv.logger.Info("Server is closing")
		er := subscription.Unsubscribe()
		if er != nil {
			srv.logger.Error(fmt.Sprintf("Unsubscribe error: %+v\n ", er))
		}
		conn.Conn.Close()
	}

	srv.logger.Info("Server started")

	select {
	case <-srv.stop:
		cleanUp()
	case <-done:
		cleanUp()
	}

	srv.logger.Info("Server Stopped")

	return nil
}

func (srv *Server) Stop() {
	if srv != nil && srv.stop != nil {
		srv.stop <- "stop"
	}
}

func subscribe(conn *oNats.EncodedConn, logger logur.Logger, subject string, queue string, handler http.Handler) (*oNats.Subscription, error) {
	return conn.QueueSubscribe(subject, queue, func(addr string, rpSubject string, rq *Request) {
		logger.Debug("Nats received message", map[string]interface{}{"url": rq.URL, "subject": subject})

		go func(enc *oNats.EncodedConn) {
			defer handlePanic(conn, logger, rpSubject)
			t1 := time.Now()

			rp := &Response{
				Headers: http.Header{},
			}

			requestID := rq.Headers.Get(XRequestId)
			if requestID == "" {
				requestID = RandomString(6)
			}

			ctx := context.Background()

			// add log
			l := log.WithFields(logger, map[string]interface{}{XHostName: hostname, "Subject": subject, "Queue": queue, XRequestId: requestID})
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
			logger.Debug("write response ", map[string]interface{}{"status": rp.StatusCode})
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
