package app

import (
	"gitlab.com/silenteer/go-nats/log"
	"gitlab.com/silenteer/go-nats/nats"
)

type Config struct {
	Logging *log.Config
	Nats    *nats.Config
}

func DefaultConfig() *Config {
	return &Config{
		Logging: log.DefaultConfig(),
		Nats:    nats.DefaultConfig(),
	}
}
