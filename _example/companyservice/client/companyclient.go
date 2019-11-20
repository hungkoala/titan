package client

import (
	"github.com/pkg/errors"
	"gitlab.com/silenteer/go-nats/_example/companyservice/api"
	"gitlab.com/silenteer/go-nats/nats"
)

type CompanyClient struct {
	natClient *nats.Client
}

func NewCompanyClient(natClient *nats.Client) *CompanyClient {
	return &CompanyClient{natClient: natClient}
}

func (client *CompanyClient) GetCompanies(ctx *nats.Context) (*[]api.CompanyDto, error) {
	request, _ := nats.NewReqBuilder().
		Get("/api/companies").
		Subject("company_service").
		Build()

	var result []api.CompanyDto
	err := client.natClient.SendAndReceiveJson(ctx, request, &result)
	return &result, err
}

func (client *CompanyClient) GetCompany(ctx *nats.Context, key string) (*api.CompanyDto, error) {
	request, _ := nats.NewReqBuilder().
		Get("/api/companies/" + key).
		Subject("company_service").
		Build()

	var result api.CompanyDto
	err := client.natClient.SendAndReceiveJson(ctx, request, &result)
	if result == (api.CompanyDto{}) {
		return nil, nil
	}
	return &result, err
}

func (client *CompanyClient) SaveCompany(ctx *nats.Context, company *api.CompanyDto) (*api.CompanyDto, error) {
	return nil, errors.New("not implemented")
}

func (client *CompanyClient) UpdateCompany(ctx *nats.Context, company *api.CompanyDto) (*api.CompanyDto, error) {
	return nil, errors.New("not implemented")
}

func (client *CompanyClient) DeleteCompany(ctx *nats.Context) (string, error) {
	return "nil", errors.New("not implemented")
}
