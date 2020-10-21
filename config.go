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
}

type NatsConfig struct {
	Servers     string
	ReadTimeout int
}

func (c NatsConfig) GetReadTimeoutDuration() time.Duration {
	return time.Duration(c.ReadTimeout) * time.Second
}

func GetNatsConfig() *NatsConfig {
	natConfigOnce.Do(func() { // <-- atomic, does not allow repeating
		natConfig = &NatsConfig{
			Servers:     viper.GetString(NatsServers),
			ReadTimeout: viper.GetInt(NatsReadTimeout),
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
