package app

import (
	"errors"

	"gitlab.com/silenteer-oss/titan"
)

type CompanyService struct {
	repository *CompanyRepository
}

func NewCompanyService(repository *CompanyRepository) *CompanyService {
	return &CompanyService{repository: repository}
}

func (com *CompanyService) Routes(r titan.Router) {
	r.RegisterJson("GET", "/api/service/companies", com.GetCompanies)
	r.RegisterJson("POST", "/api/service/companies", com.SaveCompany)
	r.RegisterJson("GET", "/api/service/companies/{name}", com.GetCompany)
	r.RegisterJson("PUT", "/api/service/companies/{name}", com.UpdateCompany)
	r.RegisterJson("DELETE", "/api/service/companies/{name}", com.DeleteCompany)
}

func (com *CompanyService) GetCompanies(ctx *titan.Context) ([]Company, error) {
	return com.repository.FindAll(), nil
}

func (com *CompanyService) GetCompany(ctx *titan.Context) (*Company, error) {
	name := ctx.GetPathParam("name")
	company, ok := com.repository.FindBy(name)

	if !ok {
		return nil, nil
	}
	return &company, nil
}

func (com *CompanyService) SaveCompany(ctx *titan.Context, company *Company) (*Company, error) {
	com.repository.Save(company.Name, Company{Name: company.Name, Tel: company.Tel, Email: company.Email})
	return company, nil
}

func (com *CompanyService) UpdateCompany(ctx *titan.Context, company *Company) (*Company, error) {
	com.repository.Save(company.Name, Company{Name: company.Name, Tel: company.Tel, Email: company.Email})
	return company, nil
}

func (com *CompanyService) DeleteCompany(ctx *titan.Context) (string, error) {
	name := ctx.GetPathParam("name")
	if name == "" {
		return "", errors.New("missing name param")
	}
	return name, nil
}
