package company

import (
	"errors"

	"gitlab.com/silenteer/go-nats/nats"
)

type CompanyService struct {
	repository *CompanyRepository
}

func NewCompanyService(repository *CompanyRepository) *CompanyService {
	return &CompanyService{repository: repository}
}

func (com *CompanyService) Routes(r nats.Router) {
	r.RegisterJson("GET", "/api/companies", com.GetCompanies)
	r.RegisterJson("POST", "/api/companies", com.SaveCompany)
	r.RegisterJson("GET", "/api/companies/{name}", com.GetCompany)
	r.RegisterJson("PUT", "/api/companies/{name}", com.UpdateCompany)
	r.RegisterJson("DELETE", "/api/companies/{name}", com.DeleteCompany)
}

func (com *CompanyService) GetCompanies(ctx *nats.Context) ([]Company, error) {
	return com.repository.FindAll(), nil
}

func (com *CompanyService) GetCompany(ctx *nats.Context) (*Company, error) {
	name := ctx.GetPathParam("name")
	company, ok := com.repository.FindBy(name)

	if !ok {
		return nil, nil
	}
	return &company, nil
}

func (com *CompanyService) SaveCompany(ctx *nats.Context, company *Company) (*Company, error) {
	com.repository.Save(company.Name, company.Company{Name: company.Name, Tel: company.Tel, Email: company.Email})
	return company, nil
}

func (com *CompanyService) UpdateCompany(ctx *nats.Context, company *Company) (*Company, error) {
	com.repository.Save(company.Name, company.Company{Name: company.Name, Tel: company.Tel, Email: company.Email})
	return company, nil
}

func (com *CompanyService) DeleteCompany(ctx *nats.Context) (string, error) {
	name := ctx.GetPathParam("name")
	if name == "" {
		return "", errors.New("missing name param")
	}
	return name, nil
}
