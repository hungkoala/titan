package nats

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
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
}

// ListenAndServe always returns a non-nil error.
func ListenAndServe(subject string, handler http.Handler) error {
	server := &Server{Subject: subject, Handler: handler}
	return server.Start()
}

func (srv *Server) Start() error {

	if srv.Handler == nil {
		return errors.New("nats: Handler not found")
	}

	if srv.Subject == "" {
		return errors.New("nats: Subject can not be empty")
	}

	timeout := srv.ReadTimeout
	if timeout == 0 {
		timeout = 3 * time.Second
	}

	timeoutHandler := http.TimeoutHandler(srv.Handler, timeout, "timeout")

	addr := srv.Addr
	if addr == "" {
		addr = oNats.DefaultURL
	}

	logger := log.DefaultLogger(nil)

	conn, err := NewConnection(addr)
	if err != nil {
		return err
	}
	logger.Info("Connecting to NATS Server at: ", map[string]interface{}{"add": addr, "subject": srv.Subject})

	subscription, err := srv.serve(conn.Conn, logger, srv.Subject, timeoutHandler)
	if err != nil {
		return err
	}

	// Handle SIGINT and SIGTERM.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	logger.Info("Server is being closed")
	_ = subscription.Unsubscribe()
	conn.Conn.Close()

	return nil
}

func (srv *Server) serve(conn *oNats.EncodedConn, logger logur.Logger, subject string, handler http.Handler) (*oNats.Subscription, error) {
	return conn.Subscribe(subject, func(addr string, rpSubject string, rq *Request) {
		logger.Debug("Received message", map[string]interface{}{"url": rq.URL, "subject": subject})

		go func(enc *oNats.EncodedConn) {
			defer handlePanic(conn, logger, rpSubject)

			resp := &Response{
				StatusCode: 200, // internal server error as default
				Status:     "",
				Headers:    http.Header{},
			}

			req, err := natsRequestToHttpRequest(rq)
			if err != nil {
				replyError(enc, logger, err, rpSubject)
				return
			}

			handler.ServeHTTP(resp, req)

			_ = enc.Publish(rpSubject, resp)

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

func natsRequestToHttpRequest(rq *Request) (*http.Request, error) {
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
