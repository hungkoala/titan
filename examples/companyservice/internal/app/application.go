package app

import (
	"gitlab.com/silenteer/titan"
)

func NewServer() *titan.Server {
	companyRepository := NewCompanyRepository()
	companyService := NewCompanyService(companyRepository)

	return titan.NewServer(
		"api.service.companies",
		titan.Routes(companyService.Routes),
	)
}
