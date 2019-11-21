package app

import (
	"gitlab.com/silenteer/titan/log"
	"gitlab.com/silenteer/titan/nats"
)

func NewServer(config *Config) *nats.Server {
	logger := log.NewLogger(config.Logging)
	companyRepository := NewCompanyRepository()
	companyService := NewCompanyService(companyRepository)

	return nats.NewServer(
		nats.SetConfig(config.Nats),
		nats.Routes(companyService.Routes),
		nats.Logger(logger),
	)
}
