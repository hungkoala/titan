package api

import (
	"gitlab.com/silenteer/titan/kaka"
)

type CompanyDto struct {
	Name  string `json:"name"`
	Tel   string `json:"tel"`
	Email string `json:"email"`
}

type CompanyService interface {
	GetCompanies(ctx *kaka.Context) (*[]CompanyDto, error)

	GetCompany(ctx *kaka.Context, key string) (*CompanyDto, error)

	SaveCompany(ctx *kaka.Context, company *CompanyDto) (*CompanyDto, error)

	UpdateCompany(ctx *kaka.Context, company *CompanyDto) (*CompanyDto, error)

	DeleteCompany(ctx *kaka.Context, key string) (string, error)
}
