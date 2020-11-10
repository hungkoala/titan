package titan

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"

	"logur.dev/logur"

	"gitlab.com/silenteer-oss/titan/log"

	"github.com/spf13/viper"
)

var hostname string

var natConfigOnce sync.Once
var natConfig *NatsConfig

var logConfigOnce sync.Once
var logConfig *log.Config

var mux sync.Mutex
var defaultClient *Client

var loggerOnce sync.Once
var logger logur.Logger

const (
	NatsServers     = "Nats.Servers"
	NatsReadTimeout = "Nats.ReadTimeout"

	// see https://docs.nats.io/developing-with-nats/connecting/pingpong
	NatsPingInterval        = "Nats.PingInterval"
	NatsMaxPingsOutstanding = "Nats.MaxPingsOutstanding"

	//see https://docs.nats.io/developing-with-nats/events/slow
	NatsPendingLimitByte = "Nats.PendingLimitByte"
	NatsPendingLimitMsg  = "Nats.PendingLimitMsg"

	LoggingFormat  = "Logging.Format"
	LoggingLevel   = "Logging.Level"
	LoggingNoColor = "Logging.NoColor"
)

func init() {
	var err error
	hostname, err = os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}
	//hostname = hostname + "." + RandomString(6)

	viper.AddConfigPath(".")
	viper.SetConfigName("config")

	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// logging
	viper.SetDefault(LoggingFormat, "logfmt")
	viper.SetDefault(LoggingLevel, "debug")
	viper.SetDefault(LoggingNoColor, false)

	// nats
	viper.SetDefault(NatsServers, "nats://127.0.0.1:4222, nats://localhost:4222")
	viper.SetDefault(NatsReadTimeout, 99999)
	// see https://docs.nats.io/developing-with-nats/connecting/pingpong
	viper.SetDefault(NatsPingInterval, 20)
	viper.SetDefault(NatsMaxPingsOutstanding, 10)
	viper.SetDefault(NatsPendingLimitByte, -1)
	viper.SetDefault(NatsPendingLimitMsg, -1)

}

type NatsConfig struct {
	Servers             string
	ReadTimeout         int
	PingInterval        int
	MaxPingsOutstanding int
	PendingLimitMsg     int
	PendingLimitByte    int
}

func (c NatsConfig) GetReadTimeoutDuration() time.Duration {
	return time.Duration(c.ReadTimeout) * time.Second
}

func GetNatsConfig() *NatsConfig {
	natConfigOnce.Do(func() { // <-- atomic, does not allow repeating
		natConfig = &NatsConfig{
			Servers:             viper.GetString(NatsServers),
			ReadTimeout:         viper.GetInt(NatsReadTimeout),
			PingInterval:        viper.GetInt(NatsPingInterval),
			MaxPingsOutstanding: viper.GetInt(NatsMaxPingsOutstanding),
			PendingLimitMsg:     viper.GetInt(NatsPendingLimitMsg),
			PendingLimitByte:    viper.GetInt(NatsPendingLimitByte),
		}
	})
	return natConfig
}

func GetLogConfig() *log.Config {
	logConfigOnce.Do(func() { // <-- atomic, does not allow repeating
		logConfig = &log.Config{
			Format:  viper.GetString(LoggingFormat),
			Level:   viper.GetString(LoggingLevel),
			NoColor: viper.GetBool(LoggingNoColor),
		}
	})
	return logConfig
}

func GetLogger() logur.Logger {
	loggerOnce.Do(func() { // <-- atomic, does not allow repeating
		logger = log.WithFields(log.NewLogger(GetLogConfig()), map[string]interface{}{"hostname": hostname})
	})
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
		nats.Name(fmt.Sprintf("%s_%s", hostname, "client")),
		nats.Timeout(10*time.Second), // connection timeout
		nats.MaxReconnects(-1),       // never give up
		nats.ErrorHandler(func(_ *nats.Conn, s *nats.Subscription, e error) {
			if e != nil {
				pendingMsg, _, _ := s.Pending()
				log.Error(fmt.Sprintf("Nats client subject=%s, queue name=%s, pending messages=%d, error %+v", s.Subject, s.Queue, pendingMsg, e))
			}
		}),
		nats.DisconnectErrHandler(func(_ *nats.Conn, e error) {
			if e != nil {
				log.Error(fmt.Sprintf("Nats client disconect error %+v", e))
			}
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			log.Debug("Nats client Reconnect")
		}),
		nats.DiscoveredServersHandler(func(_ *nats.Conn) {
			log.Debug("Nats client Discovered")
		}),
		nats.PingInterval(time.Duration(config.PingInterval)*time.Second),
		nats.MaxPingsOutstanding(config.MaxPingsOutstanding),
	)

	if err != nil {
		fmt.Printf("nats client connection error %+v\n", err)
		os.Exit(1)
	}

	go func() {
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
		// close connection on exit
		<-done
		conn.Conn.Close()
	}()

	defaultClient = &Client{conn}
	mux.Unlock()

	return defaultClient
}

func GetDefaultServer(config *NatsConfig, logger logur.Logger, subject string) (*Connection, error) {
	return NewConnection(
		config.Servers,
		nats.Timeout(10*time.Second), // connection timeout
		nats.Name(fmt.Sprintf("%s_%s", subject, hostname)),
		nats.MaxReconnects(-1), // never give up
		nats.ErrorHandler(func(_ *nats.Conn, s *nats.Subscription, e error) {
			if e != nil {
				if s != nil {
					pendingMsg, _, _ := s.Pending()
					logger.Error(fmt.Sprintf("Nats server subject=%s, queue name=%s, pending messages=%d, error %+v", s.Subject, s.Queue, pendingMsg, e))
				} else {
					logger.Error(fmt.Sprintf("Nats server error %+v", e))
				}
			}
		}),
		nats.DisconnectErrHandler(func(s *nats.Conn, e error) {
			if e != nil {
				logger.Error(fmt.Sprintf("Nats server disconect error %+v", e))
			}
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			logger.Debug("Nats server  Reconnect")
		}),
		nats.DiscoveredServersHandler(func(_ *nats.Conn) {
			logger.Debug("Nats server  Discovered")
		}),
		nats.PingInterval(time.Duration(config.PingInterval)*time.Second),
		nats.MaxPingsOutstanding(config.MaxPingsOutstanding),
	)
}
