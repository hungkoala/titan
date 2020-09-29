package restful

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gitlab.com/silenteer-oss/titan/socket"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"

	"gitlab.com/silenteer-oss/titan"

	"github.com/go-chi/chi"

	"github.com/pkg/errors"
	"logur.dev/logur"
)

// Option is a function on the options for a connection.
type Option func(*Options) error

// Options can be used to create a customized connection.
type Options struct {
	logger        logur.Logger
	router        titan.Router
	tlsEnable     bool   // base64 encoding of DER format
	tlsKey        string // base64 encoding of DER format
	tlsCert       string
	port          string
	socketEnable  bool // accept socket connection
	socketHandler map[string]socket.HandlerFunc
	statics       map[string]string // Serve static files
	corsDomain    []string          // allow cors by domains name
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

// Cores allow cors with domains name
func Cores(domains []string) Option {
	return func(o *Options) error {
		o.corsDomain = domains
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

func SocketRoute(path string, h socket.HandlerFunc) Option {
	return func(o *Options) error {
		o.socketHandler[path] = h
		return nil
	}
}

func SocketEnable(v bool) Option {
	return func(o *Options) error {
		o.socketEnable = v
		return nil
	}
}

func Static(path, directory string) Option {
	return func(o *Options) error {
		if o.statics == nil {
			o.statics = make(map[string]string)
		}
		o.statics[path] = directory
		return nil
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
	tlsCert       string
	port          string
	handler       http.Handler // handler to invoke, http.DefaultServeMux if nil
	logger        logur.Logger
	stop          chan interface{}
	socketManager *socket.SocketManager
	socketHandler map[string]socket.HandlerFunc
	statics       map[string]string // Serve static files
}

func NewServer(options ...Option) *Server {
	// default logger
	logger := titan.GetLogger()

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(titan.NewMiddleware("Http", logger))

	corsDomain := extractCorsDomain(options...)

	if len(corsDomain) != 0 {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   corsDomain,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}))
	}

	// set default handlers - health check and build info
	defaultHandlers := &titan.DefaultHandlers{Subject: ""}
	defaultRouters := Routes(defaultHandlers.Routes)

	withDefaultOptions := append(append(getDefaultConfig(), options...), defaultRouters)
	// default options
	opts := Options{
		logger:        logger,
		router:        titan.NewRouter(r),
		socketHandler: make(map[string]socket.HandlerFunc),
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

	logConfig := titan.GetLogConfig()
	logger.Debug("Server Log Config :", map[string]interface{}{
		"format":  logConfig.Format,
		"level":   logConfig.Level,
		"NoColor": logConfig.NoColor,
	})

	logger.Debug("Server Http Config :", map[string]interface{}{
		"tlsEnable":    opts.tlsEnable,
		"port":         opts.port,
		"socketEnable": opts.socketEnable,
		//"cert":         opts.tlsCert,
		//"key":          opts.tlsKey,
	})

	srv := &Server{
		tlsEnable:     opts.tlsEnable,
		tlsKey:        opts.tlsKey,
		tlsCert:       opts.tlsCert,
		port:          opts.port,
		handler:       opts.router,
		logger:        opts.logger,
		socketHandler: opts.socketHandler,
		statics:       opts.statics,
	}

	if opts.socketEnable {
		srv.socketManager = socket.InitSocketManager(opts.logger)
	}

	//register statics
	for path, directory := range opts.statics {
		fmt.Println(strings.TrimRight(fmt.Sprintf("/%s/", path), "/"))
		fmt.Println("directory=", directory)
		if path != "" && directory != "" {
			r.Handle(
				fmt.Sprintf("/%s/", directory),
				http.StripPrefix(strings.TrimRight(fmt.Sprintf("/%s/", path), "/"), http.FileServer(http.Dir(directory))))
		}
	}

	return srv
}

func (srv *Server) Start(started ...chan interface{}) {
	err := srv.start(started...)
	if err != nil {
		srv.logger.Error(fmt.Sprintf("Nats server start error: %+v\n ", err))
		os.Exit(1)
	}
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
		defer func() {
			if r := recover(); r != nil {
				var ok bool
				var err error
				err, ok = r.(error)
				if !ok {
					err = fmt.Errorf("panic : %v", r)
				}
				srv.logger.Error(fmt.Sprintf("&&&&  HTTP Sockket panic error %+v\n ", err))
			}
		}()

		server = &http.Server{
			Handler: srv.handler,
			// Other options
		}

		// use proxy instead
		//if srv.socketManager != nil {
		server.Handler = srv
		//}

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

	if srv.socketManager != nil {
		srv.socketManager.Start()
	}

	// Handle SIGINT and SIGTERM.
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	srv.stop = make(chan interface{}, 1)

	// wait for signal
	select {
	case <-srv.stop:
	case <-done:
	}

	if srv.socketManager != nil {
		srv.socketManager.Stop()
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

// the socket proxy
func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			var err error
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("panic : %v", r)
			}
			srv.logger.Error(fmt.Sprintf("&&&&  HTTP Sockket panic error %+v\n ", err))
		}
	}()

	// is this socket request
	if f, ok := srv.socketHandler[strings.ToLower(r.RequestURI)]; ok {
		f(srv.socketManager, srv.logger, w, r)
		return
	}

	//normal http request
	srv.handler.ServeHTTP(w, r)
}

func extractCorsDomain(options ...Option) []string {
	r := chi.NewRouter()
	optionsWithCORS := Options{
		router: titan.NewRouter(r),
	}
	for _, o := range options {
		o(&optionsWithCORS)
	}
	return optionsWithCORS.corsDomain
}
