package restful

import (
	"fmt"

	"github.com/spf13/viper"
)

const (
	C_TlsEnable = "tls.enable"
	C_TlsCert   = "tls.cert"
	C_TlsKey    = "tls.key"
	C_Port      = "port"
)

func init() {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.AutomaticEnv() // read in environment variables that match
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// logging
	viper.SetDefault(C_TlsEnable, false)
	viper.SetDefault(C_TlsCert, "")
	viper.SetDefault(C_TlsKey, "")
	viper.SetDefault(C_Port, "")

}

func getDefaultConfig() []Option {
	options := []Option{
		TlsEnable(viper.GetBool(C_TlsEnable)),
		TlsCert(viper.GetString(C_TlsCert)),
		TlsKey(viper.GetString(C_TlsKey)),
		Port(viper.GetString(C_Port)),
	}
	return options
}
