package api

import (
	"github.com/pkg/errors"
	"gitlab.com/silenteer/titan"
)

type CompanyClient struct {
	natClient *titan.Client
}

func NewCompanyClient(natClient *titan.Client) *CompanyClient {
	return &CompanyClient{natClient: natClient}
}

func (client *CompanyClient) GetCompanies(ctx *titan.Context) (*[]CompanyDto, error) {
	request, _ := titan.NewReqBuilder().
		Get("/api/service/companies").
		Build()

	var result []CompanyDto
	err := client.natClient.SendAndReceiveJson(ctx, request, &result)
	return &result, err
}

func (client *CompanyClient) GetCompany(ctx *titan.Context, key string) (*CompanyDto, error) {
	request, _ := titan.NewReqBuilder().
		Get("/api/service/companies/" + key).
		Build()

	var result CompanyDto
	err := client.natClient.SendAndReceiveJson(ctx, request, &result)
	if result == (CompanyDto{}) {
		return nil, nil
	}
	return &result, err
}

func (client *CompanyClient) SaveCompany(ctx *titan.Context, company *CompanyDto) (*CompanyDto, error) {
	return nil, errors.New("not implemented")
}

func (client *CompanyClient) UpdateCompany(ctx *titan.Context, company *CompanyDto) (*CompanyDto, error) {
	return nil, errors.New("not implemented")
}

func (client *CompanyClient) DeleteCompany(ctx *titan.Context) (string, error) {
	return "nil", errors.New("not implemented")
}
