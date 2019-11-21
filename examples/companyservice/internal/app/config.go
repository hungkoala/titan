package app

import (
	"gitlab.com/silenteer/titan/kaka"
	"gitlab.com/silenteer/titan/log"
)

type Config struct {
	Logging *log.Config
	Nats    *kaka.Config
}

func DefaultConfig() *Config {
	return &Config{
		Logging: log.DefaultConfig(),
		Nats:    kaka.DefaultConfig(),
	}
}
