package api

import (
	"github.com/pkg/errors"
	"gitlab.com/silenteer/titan/nats"
)

type CompanyClient struct {
	natClient *nats.Client
}

func NewCompanyClient(natClient *nats.Client) *CompanyClient {
	return &CompanyClient{natClient: natClient}
}

func (client *CompanyClient) GetCompanies(ctx *nats.Context) (*[]CompanyDto, error) {
	request, _ := nats.NewReqBuilder().
		Get("/api/companies").
		Build()

	var result []CompanyDto
	err := client.natClient.SendAndReceiveJson(ctx, request, &result)
	return &result, err
}

func (client *CompanyClient) GetCompany(ctx *nats.Context, key string) (*CompanyDto, error) {
	request, _ := nats.NewReqBuilder().
		Get("/api/companies/" + key).
		Build()

	var result CompanyDto
	err := client.natClient.SendAndReceiveJson(ctx, request, &result)
	if result == (CompanyDto{}) {
		return nil, nil
	}
	return &result, err
}

func (client *CompanyClient) SaveCompany(ctx *nats.Context, company *CompanyDto) (*CompanyDto, error) {
	return nil, errors.New("not implemented")
}

func (client *CompanyClient) UpdateCompany(ctx *nats.Context, company *CompanyDto) (*CompanyDto, error) {
	return nil, errors.New("not implemented")
}

func (client *CompanyClient) DeleteCompany(ctx *nats.Context) (string, error) {
	return "nil", errors.New("not implemented")
}
