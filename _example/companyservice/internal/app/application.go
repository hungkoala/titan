package app

import (
	"gitlab.com/silenteer/go-nats/log"
	"gitlab.com/silenteer/go-nats/nats"
)

func NewServer() *nats.Server {
	logger := log.DefaultLogger(nil)
	companyRepository := NewCompanyRepository()
	companyService := NewCompanyService(companyRepository)

	return nats.NewServer(
		nats.Subject("company_service"),
		nats.Routes(companyService.Routes),
		nats.Logger(logger),
	)
}
