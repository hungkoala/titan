package app

import (
	"gitlab.com/silenteer/go-nats/log"
	"gitlab.com/silenteer/go-nats/nats"
)

func NewServerAndStart() *nats.Server {
	logger := log.DefaultLogger(nil)
	companyRepository := NewCompanyRepository()
	companyService := NewCompanyService(companyRepository)

	return nats.NewServerAndStart(
		nats.Subject("company_service"),
		nats.Routes(companyService.Routes),
		nats.Logger(logger),
	)
}
