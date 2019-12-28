package titan

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"gitlab.com/silenteer/titan/log"
	"logur.dev/logur"
)

// Option is a function on the options for a connection.
type Option func(*Options) error

// Options can be used to create a customized connection.
type Options struct {
	queue  string
	config *NatsConfig
	router Router
}

func Queue(queue string) Option {
	return func(o *Options) error {
		o.queue = queue
		return nil
	}
}

func Config(config *NatsConfig) Option {
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

func NewServer(subject string, options ...Option) *Server {
	logger := log.WithFields(GetLogger(), map[string]interface{}{"subject": subject})
	config := GetNatsConfig()

	r := chi.NewRouter()
	r.Use(RouteParamsMiddleware)

	// set default handlers
	defaultHandlers := &DefaultHandlers{subject: subject}
	optionsWithDefault := append(options, Routes(defaultHandlers.Routes))

	opts := Options{
		config: config,
		router: NewRouter(r),
		queue:  "workers",
	}

	for _, opt := range optionsWithDefault {
		if opt != nil {
			if err := opt(&opts); err != nil {
				logger.Error(fmt.Sprintf("Nats server creation error: %+v\n ", err))
				os.Exit(1)
			}
		}
	}
	return &Server{
		subject: subject,
		queue:   opts.queue,
		config:  opts.config,
		handler: opts.router,
		logger:  log.WithFields(logger, map[string]interface{}{"queue": opts.queue}),
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
	subject string
	queue   string
	config  *NatsConfig
	handler http.Handler // handler to invoke, http.DefaultServeMux if nil
	logger  logur.Logger
	stop    chan interface{}
}

func (srv *Server) start() error {

	if srv.handler == nil {
		return errors.New("nats: Handler not found")
	}

	config := srv.config
	if config == nil {
		return errors.New("nats: NatsConfig can not be nil")
	}

	if srv.subject == "" {
		return errors.New("nats: Subject can not be empty")
	}

	if config.Servers == "" {
		return errors.New("nats: Address can not be empty")
	}

	if srv.logger == nil {
		return errors.New("nats: Logger can not be empty")
	}

	if config.ReadTimeout <= 0 {
		return errors.New("nats: ReadTimeout can not be empty")
	}

	timeoutHandler := http.TimeoutHandler(srv.handler, time.Duration(config.ReadTimeout)*time.Second, "nats handler timeout")

	srv.logger.Info("Connecting to NATS Server at: ", map[string]interface{}{"add": config.Servers})
	conn, err := NewConnection(
		config.Servers,
		nats.Timeout(10*time.Second), // connection timeout
		nats.Name(hostname+srv.subject),
		nats.MaxReconnects(-1), // never give up
		nats.ErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, e error) {
			if e != nil {
				srv.logger.Error(fmt.Sprintf("Nats server error %+v", e))
			}
		}),
		nats.DisconnectErrHandler(func(_ *nats.Conn, e error) {
			if e != nil {
				srv.logger.Error(fmt.Sprintf("Nats server disconect error %+v", e))
			}
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			srv.logger.Debug("Nats server  Reconnect")
		}),
		nats.DiscoveredServersHandler(func(_ *nats.Conn) {
			srv.logger.Debug("Nats server  Discovered")
		}),
	)

	if err != nil {
		return errors.WithMessage(err, "Nats connection error ")
	}

	subscription, err := subscribe(conn.Conn, srv.logger, srv.subject, srv.queue, timeoutHandler)
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

func subscribe(conn *nats.EncodedConn, logger logur.Logger, subject string, queue string, handler http.Handler) (*nats.Subscription, error) {
	return conn.QueueSubscribe(subject, queue, func(addr string, rpSubject string, rq *Request) {
		go func(enc *nats.EncodedConn) {
			t := time.Now()
			if rq.Headers == nil {
				rq.Headers = http.Header{}
			}
			requestID := rq.Headers.Get(XRequestId)
			if requestID == "" {
				requestID = RandomString(6)
				rq.Headers.Set(XRequestId, requestID)
			}
			logWithId := log.WithFields(logger, map[string]interface{}{"id": requestID, "method": rq.Method})
			go func() {
				url := extractLoggablePartsFromUrl(rq.URL)
				logWithId.Debug("Nats server received request", map[string]interface{}{"url": url})
			}()

			defer handlePanic(conn, logWithId, rpSubject)

			rp := &Response{
				Headers: http.Header{},
			}

			ctx := context.Background()

			// add log
			ctx = context.WithValue(ctx, XLoggerId, logWithId)

			// add request id
			ctx = context.WithValue(ctx, XRequestId, requestID)

			userInfoJson := rq.Headers.Get(XUserInfo)

			if userInfoJson != "" {
				var userInfo UserInfo
				jerr := json.Unmarshal([]byte(userInfoJson), &userInfo)
				if jerr != nil {
					logger.Error(fmt.Sprintf("Unmarshal User Info  error: %+v\n ", jerr))
				} else {
					ctx = context.WithValue(ctx, XUserInfo, &userInfo)
				}
			}

			rq, err := natsRequestToHttpRequest(rq, ctx)
			if err != nil {
				replyError(enc, logWithId, err, rpSubject)
				return
			}

			defer func() {
				logWithId.Info("Nats server request complete", map[string]interface{}{
					"method":     rq.Method,
					"url":        rq.URL,
					"status":     rp.StatusCode,
					"elapsed_ms": float64(time.Since(t).Nanoseconds()) / 1000000.0},
				)
			}()

			handler.ServeHTTP(rp, rq)
			err = enc.Publish(rpSubject, rp)

			if err != nil {
				logWithId.Error(fmt.Sprintf("Nats error on publish result back: %+v\n ", err))
			}
		}(conn)
	})
}

func handlePanic(enc *nats.EncodedConn, logger logur.Logger, rpSubject string) {
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

func replyError(enc *nats.EncodedConn, logger logur.Logger, err error, rpSubject string) {
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

func natsRequestToHttpRequest(rq *Request, c context.Context) (*http.Request, error) {
	var body io.Reader
	if rq.Body != nil {
		body = bytes.NewReader(rq.Body)
	} else {
		body = bytes.NewReader([]byte{})
	}

	topic := extractTopicFromHttpUrl(rq.URL)

	request, err := http.NewRequestWithContext(c, rq.Method, topic, body)
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
