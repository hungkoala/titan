package titan

import "github.com/nats-io/nats.go"

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
