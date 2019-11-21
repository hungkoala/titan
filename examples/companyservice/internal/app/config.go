package app

import (
	"gitlab.com/silenteer/titan"
	"gitlab.com/silenteer/titan/log"
)

type Config struct {
	Logging *log.Config
	Nats    *titan.Config
}

func DefaultConfig() *Config {
	return &Config{
		Logging: log.DefaultConfig(),
		Nats:    titan.DefaultConfig(),
	}
}
