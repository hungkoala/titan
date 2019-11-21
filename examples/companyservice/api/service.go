package api

import (
	"gitlab.com/silenteer/titan/nats"
)

type CompanyDto struct {
	Name  string `json:"name"`
	Tel   string `json:"tel"`
	Email string `json:"email"`
}

type CompanyService interface {
	GetCompanies(ctx *nats.Context) (*[]CompanyDto, error)

	GetCompany(ctx *nats.Context, key string) (*CompanyDto, error)

	SaveCompany(ctx *nats.Context, company *CompanyDto) (*CompanyDto, error)

	UpdateCompany(ctx *nats.Context, company *CompanyDto) (*CompanyDto, error)

	DeleteCompany(ctx *nats.Context, key string) (string, error)
}
