package app

import (
	"gitlab.com/silenteer/titan/log"
	"gitlab.com/silenteer/titan/nats"
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
