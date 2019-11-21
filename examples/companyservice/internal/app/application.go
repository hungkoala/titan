package app

import (
	"gitlab.com/silenteer/titan"
	"gitlab.com/silenteer/titan/log"
)

func NewServer(config *Config) titan.Server {
	logger := log.NewLogger(config.Logging)
	companyRepository := NewCompanyRepository()
	companyService := NewCompanyService(companyRepository)

	return titan.NewServer(
		titan.SetConfig(config.Nats),
		titan.Routes(companyService.Routes),
		titan.Logger(logger),
	)
}
