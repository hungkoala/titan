package titan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/opentracing/opentracing-go"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"gitlab.com/silenteer-oss/titan/log"
	"gitlab.com/silenteer-oss/titan/tracing"
	"logur.dev/logur"
)

// Option is a function on the options for a connection.
type Option func(*Options) error

// Options can be used to create a customized connection.
type Options struct {
	logger            logur.Logger
	queue             string
	config            *NatsConfig
	router            Router
	messageSubscriber *MessageSubscriber
	tracer            opentracing.Tracer
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

func Subscribe(r func(*MessageSubscriber)) Option {
	return func(o *Options) error {
		r(o.messageSubscriber)
		return nil
	}
}

func NewServer(subject string, options ...Option) *Server {
	tracing.InitTracing(subject)
	// default logger
	logger := log.WithFields(GetLogger(), map[string]interface{}{"subject": subject})

	if err := InitRemoteConfig(subject); err != nil {
		logger.Error(fmt.Sprintf("InitRemoteConfig error: %+v\n ", err))
	}

	natConfig := GetNatsConfig()
	logConfig := GetLogConfig()

	logger.Debug("NATS Config :", map[string]interface{}{"Servers": natConfig.Servers, "ReadTimeout": natConfig.ReadTimeout})
	logger.Debug("Log Config :", map[string]interface{}{"format": logConfig.Format, "level": logConfig.Level, "NoColor": logConfig.NoColor})

	r := chi.NewRouter()
	r.Use(
		NewMiddleware("NATS", subject, logger),
	)

	// set default handlers
	// health check and build info
	defaultHandlers := &DefaultHandlers{Subject: subject}
	withDefaultOptions := append(options, Routes(defaultHandlers.Register), Subscribe(defaultHandlers.Subscribe))

	// default options
	opts := Options{
		logger:            logger,
		config:            GetNatsConfig(),
		router:            NewRouter(r),
		queue:             "workers",
		messageSubscriber: NewMessageSubscriber(logger),
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
		subject:           subject,
		queue:             opts.queue,
		config:            opts.config,
		handler:           opts.router,
		messageSubscriber: opts.messageSubscriber,
		logger:            log.WithFields(opts.logger, map[string]interface{}{"queue": opts.queue}),
		tracer:            opts.tracer,
	}
}

func (srv *Server) Start(started ...chan interface{}) {
	err := srv.start(started...)
	if err != nil {
		srv.logger.Error(fmt.Sprintf("Nats server start error: %+v\n ", err))
		os.Exit(1)
	}
}

// will change this to ServerInterface to make it consistency
type IServer interface {
	Stop()
	Start(started ...chan interface{})
}

type Server struct {
	subject           string
	queue             string
	config            *NatsConfig
	handler           http.Handler // handler to invoke, http.DefaultServeMux if nil
	messageSubscriber *MessageSubscriber
	logger            logur.Logger
	stop              chan interface{} // command that instruct the server should be shutdown
	stopped           chan interface{} // inform client that the server has stop
	//msgNum            int64            // number of processing messages
	tracer opentracing.Tracer
}

func (srv *Server) start(started ...chan interface{}) error {

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

	timeoutHandler := http.TimeoutHandler(srv.handler, config.GetReadTimeoutDuration(), `{"message": "nats handler timeout"}`)

	srv.logger.Info("Connecting to NATS Server at: ", map[string]interface{}{"add": config.Servers})
	conn, err := GetDefaultServer(config, srv.logger, srv.subject)

	if err != nil {
		return errors.WithMessage(err, "Nats connection error ")
	}

	subscription, err := subscribe(conn.Conn, srv.logger, srv.subject, srv.queue, timeoutHandler)
	if err != nil {
		return errors.WithMessage(err, "Nats serve subscribe error ")
	}
	_monitoringSubscription = subscription

	err = subscription.SetPendingLimits(config.PendingLimitMsg, config.PendingLimitByte)
	if err != nil {
		return errors.WithMessage(err, "Nats serve  set pending limits error ")
	}

	err = srv.messageSubscriber.subscribe(conn.Conn)
	if err != nil {
		return errors.WithMessage(err, "Nats serve messageSubscriber error ")
	}

	err = conn.Flush()
	if err != nil {
		srv.logger.Error(fmt.Sprintf("Nats serve flush subscription  error: %+v\n ", err))
	}

	// Handle SIGINT and SIGTERM.
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	// stop command from apps/users
	srv.stop = make(chan interface{}, 1)

	// notified unit test/other apps that the server has been stopped
	srv.stopped = make(chan interface{}, 1)

	srv.logger.Info("Server started")
	for i := range started {
		started[i] <- true
	}

	// wait for stop command  or interrupt (ctr+c)
	select {
	case <-srv.stop:
	case <-done:
	}

	srv.logger.Info("Server is closing")
	er := subscription.Drain()
	if er != nil {
		srv.logger.Error(fmt.Sprintf("Unsubscribe error: %+v\n ", er))
	}
	srv.messageSubscriber.drain()

	er = conn.Flush()
	if er != nil {
		srv.logger.Error(fmt.Sprintf("Flush error: %+v\n ", er))
	}

	close(srv.stop)
	srv.stop = nil

	//check end of waiting messages
	endOfMsg := make(chan struct{})
	go func() {
		var numOfMsg uint64
		ticker := time.NewTicker(100 * time.Millisecond)
		for range ticker.C {
			numOfMsg = MsgCountLoad()
			if numOfMsg <= 0 {
				close(endOfMsg)
				return
			}
		}
	}()

	// wait for all messages processed or timeout
	select {
	case <-endOfMsg:
	case <-time.After(15 * time.Second):
	}

	conn.Drain()

	srv.stopped <- "stopped"
	close(srv.stopped)
	srv.logger.Info("Server Stopped")
	return nil
}

func (srv *Server) Stop() {
	if srv != nil && srv.stop != nil {
		srv.stop <- "stop command"
	}
	// wait for server stop
	<-srv.stopped
}

func subscribe(conn *nats.EncodedConn, logger logur.Logger, subject string, queue string, handler http.Handler) (*nats.Subscription, error) {
	return conn.QueueSubscribe(subject, queue, func(addr string, rpSubject string, r []byte) {
		go func(enc *nats.EncodedConn, msg []byte) {
			//t := time.Now()
			var rq Request
			err := json.Unmarshal(msg, &rq)
			if err != nil {
				logger.Error(fmt.Sprintf("Nats server desrialize body error: %+v", err))
				return
			}

			if rq.Headers == nil {
				rq.Headers = http.Header{}
			}
			requestID := rq.Headers.Get(XRequestId)
			if requestID == "" {
				requestID = RandomString(6)
				rq.Headers.Set(XRequestId, requestID)
			}
			logWithId := log.WithFields(logger, map[string]interface{}{
				"id":      requestID,
				"method":  rq.Method,
				"subject": subject,
			})

			defer handlePanic(conn, logWithId, rpSubject)

			rp := &Response{
				Headers: http.Header{},
			}

			BeginRequest(logWithId, &rq)
			defer EndRequest(logWithId, rp)

			httpReq, err := NatsRequestToHttpRequest(&rq)
			if err != nil {
				replyError(enc, logWithId, err, rpSubject)
				return
			}

			// forward request to Controller
			handler.ServeHTTP(rp, httpReq)

			rp.Headers.Set(XResponeTime, strconv.FormatInt(time.Now().UnixNano(), 10))

			// send response back
			err = enc.Publish(rpSubject, rp)
			if err != nil {
				replyError(enc, logWithId, err, rpSubject)
				logWithId.Error(fmt.Sprintf("Nats error on publish result back: %+v\n ", err))
			}
		}(conn, r)
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
