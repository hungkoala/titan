package titan

import (
	"fmt"
	"os"

	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
)

type Config struct {
	Servers     string
	Subject     string
	Queue       string
	ReadTimeout int
}

func DefaultConfig() *Config {
	return &Config{
		Servers:     nats.DefaultURL,
		ReadTimeout: 5, //second
		Subject:     "test",
		Queue:       "workers",
	}
}

func InitViperConfig(cfgFile string) {
	if cfgFile != "" {
		// Use config file from the flag.
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			fmt.Println(fmt.Sprintf("File `%s` does not exist", cfgFile))
			os.Exit(1)
		}
		viper.SetConfigFile(cfgFile)
	} else {
		// search config file in current dir
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println(fmt.Sprintf("File `%s` does not exist", viper.ConfigFileUsed()))
		os.Exit(1)
	}
}
