package titan

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"

	"logur.dev/logur"

	"gitlab.com/silenteer-oss/titan/log"

	"github.com/spf13/viper"
)

var hostname string
var natConfig NatsConfig
var logConfig log.Config
var defaultClient *Client
var mux sync.Mutex
var logger logur.Logger

const (
	NatsServers     = "Nats.Servers"
	NatsReadTimeout = "Nats.ReadTimeout"
	LoggingFormat   = "Logging.Format"
	LoggingLevel    = "Logging.Level"
	LoggingNoColor  = "Logging.NoColor"
)

func init() {
	var err error
	hostname, err = os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}
	hostname = hostname + "." + RandomString(6)

	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.AutomaticEnv() // read in environment variables that match
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// logging
	viper.SetDefault(LoggingFormat, "logfmt")
	viper.SetDefault(LoggingLevel, "debug")
	viper.SetDefault(LoggingNoColor, false)
	logConfig = log.Config{
		Format:  viper.GetString(LoggingFormat),
		Level:   viper.GetString(LoggingLevel),
		NoColor: viper.GetBool(LoggingNoColor),
	}
	logger = log.WithFields(log.NewLogger(&logConfig), map[string]interface{}{"hostname": hostname})

	// nats
	viper.SetDefault(NatsServers, "nats://127.0.0.1:4222, nats://localhost:4222")
	viper.SetDefault(NatsReadTimeout, 99999)
	natConfig = NatsConfig{
		Servers:     viper.GetString(NatsServers),
		ReadTimeout: viper.GetInt(NatsReadTimeout),
	}

	logger.Debug("NATS Config :", map[string]interface{}{"Servers": natConfig.Servers, "ReadTimeout": natConfig.ReadTimeout})
	logger.Debug("Log Config :", map[string]interface{}{"format": logConfig.Format, "level": logConfig.Level, "NoColor": logConfig.NoColor})
}

type NatsConfig struct {
	Servers     string
	ReadTimeout int
}

func (c NatsConfig) GetReadTimeoutDuration() time.Duration {
	return time.Duration(c.ReadTimeout) * time.Second
}

func GetNatsConfig() *NatsConfig {
	return &natConfig
}

func GetLogConfig() *log.Config {
	return &logConfig
}

func GetLogger() logur.Logger {
	if logger != nil {
		return logger
	}
	mux.Lock()
	logger = log.WithFields(log.NewLogger(GetLogConfig()), map[string]interface{}{"hostname": hostname})
	mux.Unlock()
	return logger
}

func GetDefaultClient() *Client {
	if defaultClient != nil {
		return defaultClient
	}
	config := GetNatsConfig()
	log := GetLogger()
	mux.Lock()
	conn, err := NewConnection(
		config.Servers,
		nats.Timeout(10*time.Second), // connection timeout
		nats.MaxReconnects(-1),       // never give up
		nats.ErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, e error) {
			if e != nil {
				log.Error(fmt.Sprintf("Nats client error %+v", e))
			}
		}),
		nats.DisconnectErrHandler(func(_ *nats.Conn, e error) {
			if e != nil {
				log.Error(fmt.Sprintf("Nats client disconect error %+v", e))
			}
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			log.Debug("Nats client  Reconnect")
		}),
		nats.DiscoveredServersHandler(func(_ *nats.Conn) {
			log.Debug("Nats client  Discovered")
		}),
	)

	if err != nil {
		fmt.Println(fmt.Sprintf("nats client connection error %+v", err))
		os.Exit(1)
	}

	go func() {
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		// close connection on exit
		<-done
		conn.Conn.Close()
	}()

	defaultClient = &Client{conn}
	mux.Unlock()

	return defaultClient
}
