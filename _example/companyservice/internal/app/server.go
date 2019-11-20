package app

import (
	"gitlab.com/silenteer/go-nats/_example/companyservice/internal/app/company"
	"gitlab.com/silenteer/go-nats/log"
	"gitlab.com/silenteer/go-nats/nats"
)

func NewServer() *nats.Server {
	logger := log.DefaultLogger(nil)
	companyRepository := company.NewCompanyRepository()
	companyService := company.NewCompanyService(companyRepository)

	return nats.NewServer(
		nats.Subject("company_service"),
		nats.Routes(companyService.Routes),
		nats.Logger(logger),
	)
}
