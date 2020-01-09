package titan

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/nats-io/nats.go"

	"logur.dev/logur"

	"gitlab.com/silenteer/titan/log"

	"github.com/spf13/viper"
)

var hostname string
var natConfig NatsConfig
var logConfig log.Config
var defaultClient *Client
var mux sync.Mutex
var logger logur.Logger

func init() {
	var err error
	hostname, err = os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}

	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.AutomaticEnv() // read in environment variables that match
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// set default value logging
	viper.SetDefault("Logging.Format", "logfmt")
	viper.SetDefault("Logging.Level", "debug")
	viper.SetDefault("Logging.NoColor", false)

	// nats
	viper.SetDefault("Nats.Servers", "nats://127.0.0.1:4222, nats://localhost:4222")
	viper.SetDefault("Nats.ReadTimeout", 500)

	// map environment variables to settings
	AutoLoadEnvironmentVariables()
	settings := viper.AllSettings()
	err = mapstructure.Decode(settings["nats"], &natConfig)
	fmt.Println("nats config = ", natConfig)
	if err != nil {
		fmt.Println(fmt.Sprintf("Unmarshal nats config error %+v", err))
		os.Exit(1)
	}

	err = mapstructure.Decode(settings["logging"], &logConfig)
	if err != nil {
		fmt.Println(fmt.Sprintf("Unmarshal logging config error %+v", err))
	}
}

type NatsConfig struct {
	Servers     string
	ReadTimeout int
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

func AutoLoadEnvironmentVariables() {
	// map environment variables to settings
	allKeys := viper.AllKeys()
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		key := strings.ToLower(pair[0])
		val := pair[1]
		for _, k := range allKeys {
			if key == k {
				viper.Set(k, val)
			}
		}
	}
}
