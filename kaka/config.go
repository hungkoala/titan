package kaka

import oNats "github.com/nats-io/nats.go"

// Config holds details necessary for logging.
type Config struct {
	Servers     string
	Subject     string
	Queue       string
	ReadTimeout int
}

func DefaultConfig() *Config {
	return &Config{
		Servers:     oNats.DefaultURL,
		ReadTimeout: 5, //second
		Subject:     "test",
		Queue:       "workers",
	}
}
