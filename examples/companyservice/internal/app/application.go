package app

import (
	"gitlab.com/silenteer/titan/kaka"
	"gitlab.com/silenteer/titan/log"
)

func NewServer(config *Config) *kaka.Server {
	logger := log.NewLogger(config.Logging)
	companyRepository := NewCompanyRepository()
	companyService := NewCompanyService(companyRepository)

	return kaka.NewServer(
		kaka.SetConfig(config.Nats),
		kaka.Routes(companyService.Routes),
		kaka.Logger(logger),
	)
}
