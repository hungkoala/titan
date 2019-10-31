package nats

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	oNats "github.com/nats-io/nats.go"
	"gitlab.com/silenteer/go-nats/log"
	"logur.dev/logur"
)

type Server struct {
	Addr    string // TCP address to listen on, ":http" if empty
	Subject string

	Handler http.Handler // handler to invoke, http.DefaultServeMux if nil

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	//
	// Because ReadTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout time.Duration

	// ReadHeaderTimeout is the amount of time allowed to read
	// request headers. The connection's read deadline is reset
	// after reading the headers and the Handler can decide what
	// is considered too slow for the body. If ReadHeaderTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, there is no timeout.
	ReadHeaderTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like ReadTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the
	// next request when keep-alives are enabled. If IdleTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, there is no timeout.
	IdleTimeout time.Duration

	// MaxHeaderBytes controls the maximum number of bytes the
	// server will read parsing the request header's keys and
	// values, including the request line. It does not limit the
	// size of the request body.
	// If zero, DefaultMaxHeaderBytes is used.
	MaxHeaderBytes int

	// ErrorLog specifies an optional logger for errors accepting
	// connections, unexpected behavior from handlers, and
	// underlying FileSystem errors.
	// If nil, logging is done via the log package's standard logger.
	//ErrorLog *log.Logger
	Logger logur.Logger

	inShutdown int32 // accessed atomically (non-zero means we're in Shutdown)

	subscription *oNats.Subscription
}

// ListenAndServe listens on the TCP network address addr and then calls
// Serve with handler to handle requests on incoming connections.
// Accepted connections are configured to enable TCP keep-alives.
//
// The handler is typically nil, in which case the DefaultServeMux is used.
//
// ListenAndServe always returns a non-nil error.
func ListenAndServe(addr string, handler http.Handler) error {
	server := &Server{Addr: addr, Handler: handler}
	return server.ListenAndServe()
}

func (srv *Server) ListenAndServe() error {
	if srv.shuttingDown() {
		return ErrServerClosed
	}

	if srv.Handler == nil {
		return errors.New("nats: Handler not found")
	}

	timeoutHandler := http.TimeoutHandler(srv.Handler, 3*time.Second, "timeout")

	addr := srv.Addr
	if addr == "" {
		addr = "nats://127.0.0.1:4222"
	}

	logger := srv.Logger
	if logger == nil {
		logger = log.DefaultLogger(nil)
	}

	conn := NewConnection(addr)
	//srv.Serve(conn.Enc, logger, srv.Subject)
	subscription, err := srv.Serve(conn.Enc, logger, "test", timeoutHandler)

	if err != nil {
		return err
	}

	//srv.subscription = subscription

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println(<-ch)

	// Stop the service gracefully.
	_ = subscription.Unsubscribe()

	return nil
}

var ErrServerClosed = errors.New("nats: Server closed")

func (s *Server) shuttingDown() bool {
	return atomic.LoadInt32(&s.inShutdown) != 0
}

func (srv *Server) Serve(conn *oNats.EncodedConn, logger logur.Logger, subject string, handler http.Handler) (*oNats.Subscription, error) {
	if len(subject) == 0 {
		logger.Error("No addresses to listen to")
	}

	return conn.Subscribe(subject, func(addr string, rpSubject string, rq *Request) {
		logger.Debug("Received message", map[string]interface{}{"url": rq.URL, "subject": subject})

		go func(enc *oNats.EncodedConn) {
			defer handlePanic(conn, logger, rpSubject)

			resp := &Response{
				StatusCode: 200, // internal server error as default
				Status:     "",
				Headers:    http.Header{},
			}

			req, err := natsEventToHttpRequest(rq)
			if err != nil {
				replyError(enc, logger, err, rpSubject)
				return
			}

			handler.ServeHTTP(resp, req)

			_ = enc.Publish(rpSubject, resp)

		}(conn)
	})
}

func (srv *Server) Close() {
	if srv.subscription != nil {
		_ = srv.subscription.Unsubscribe()
	}
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
	logger.Error("Error", map[string]interface{}{"error": err})
	resp := &Response{
		StatusCode: 500, // internal server error as default
		Status:     "",
		Headers:    http.Header{},
	}
	er := enc.Publish(rpSubject, resp)
	if er != nil {
		logger.Error("Error on reply back", map[string]interface{}{"Error": er})
	}
}

func natsEventToHttpRequest(rq *Request) (*http.Request, error) {
	var body io.Reader
	if rq.Body != nil {
		body = bytes.NewReader(rq.Body)
	}
	request, err := http.NewRequest(rq.Method, rq.URL, body)
	if err != nil {
		return nil, errors.New("Nats: Something wrong with creating the request" + err.Error())
	}

	if rq.Headers != nil {
		request.Header = rq.Headers
	}
	request.Header.Add("Connection", "close")

	return request, nil
}
