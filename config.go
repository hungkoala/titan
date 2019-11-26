package titan

import (
	"fmt"
	"os"

	"logur.dev/logur"

	"gitlab.com/silenteer/titan/log"

	"github.com/spf13/viper"
)

var natConfig NatsConfig
var logConfig log.Config

func init() {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.AutomaticEnv() // read in environment variables that match
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// set default value logging
	viper.SetDefault("Logging.Format", "logfmt")
	viper.SetDefault("Logging.Level", "debug")
	viper.SetDefault("Logging.NoColor", "false")

	// nats
	viper.SetDefault("Nats.Servers", "nats://127.0.0.1:4222, nats://localhost:4222")
	viper.SetDefault("Nats.ReadTimeout", 5)

	err := viper.UnmarshalKey("Nats", &natConfig)
	if err != nil {
		fmt.Println(fmt.Sprintf("Unmarshal nats config error %+v", err))
		os.Exit(1)
	}

	err = viper.UnmarshalKey("Logging", &logConfig)
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
	return log.NewLogger(GetLogConfig())
}
