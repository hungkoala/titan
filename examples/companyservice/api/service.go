package api

import (
	"gitlab.com/silenteer/titan"
)

type CompanyDto struct {
	Name  string `json:"name"`
	Tel   string `json:"tel"`
	Email string `json:"email"`
}

type CompanyService interface {
	GetCompanies(ctx *titan.Context) (*[]CompanyDto, error)

	GetCompany(ctx *titan.Context, key string) (*CompanyDto, error)

	SaveCompany(ctx *titan.Context, company *CompanyDto) (*CompanyDto, error)

	UpdateCompany(ctx *titan.Context, company *CompanyDto) (*CompanyDto, error)

	DeleteCompany(ctx *titan.Context, key string) (string, error)
}
