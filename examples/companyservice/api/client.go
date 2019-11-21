package api

import (
	"github.com/pkg/errors"
	"gitlab.com/silenteer/titan/kaka"
)

type CompanyClient struct {
	natClient *kaka.Client
}

func NewCompanyClient(natClient *kaka.Client) *CompanyClient {
	return &CompanyClient{natClient: natClient}
}

func (client *CompanyClient) GetCompanies(ctx *kaka.Context) (*[]CompanyDto, error) {
	request, _ := kaka.NewReqBuilder().
		Get("/api/companies").
		Build()

	var result []CompanyDto
	err := client.natClient.SendAndReceiveJson(ctx, request, &result)
	return &result, err
}

func (client *CompanyClient) GetCompany(ctx *kaka.Context, key string) (*CompanyDto, error) {
	request, _ := kaka.NewReqBuilder().
		Get("/api/companies/" + key).
		Build()

	var result CompanyDto
	err := client.natClient.SendAndReceiveJson(ctx, request, &result)
	if result == (CompanyDto{}) {
		return nil, nil
	}
	return &result, err
}

func (client *CompanyClient) SaveCompany(ctx *kaka.Context, company *CompanyDto) (*CompanyDto, error) {
	return nil, errors.New("not implemented")
}

func (client *CompanyClient) UpdateCompany(ctx *kaka.Context, company *CompanyDto) (*CompanyDto, error) {
	return nil, errors.New("not implemented")
}

func (client *CompanyClient) DeleteCompany(ctx *kaka.Context) (string, error) {
	return "nil", errors.New("not implemented")
}
