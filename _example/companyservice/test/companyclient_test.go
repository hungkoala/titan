package test

import (
	"context"
	"os"
	"testing"
	"time"

	"gitlab.com/silenteer/go-nats/_example/companyservice/internal/app"

	"github.com/stretchr/testify/assert"

	"gitlab.com/silenteer/go-nats/_example/companyservice/client"
	"gitlab.com/silenteer/go-nats/nats"
)

var natsClient = nats.NewClient("nats://127.0.0.1:4222")
var companyService = client.NewCompanyClient(natsClient)

func TestMain(m *testing.M) {
	var server *nats.Server

	go func() {
		server = app.NewServerAndStart()
	}()

	time.Sleep(2 * time.Millisecond)

	exitVal := m.Run()

	server.Stop()
	os.Exit(exitVal)
}

func TestGetCompanies(t *testing.T) {
	context := nats.NewContext(context.Background())
	companies, err := companyService.GetCompanies(context)

	assert.Nil(t, err)
	assert.Equal(t, len(*companies), 1)

}

func TestGetNotExistCompany(t *testing.T) {
	context := nats.NewContext(context.Background())

	company, err := companyService.GetCompany(context, "not_exist")

	assert.Nil(t, err)
	assert.Nil(t, company)
}

func TestGetExistCompany(t *testing.T) {
	context := nats.NewContext(context.Background())

	company, err := companyService.GetCompany(context, "hung")

	assert.Nil(t, err)
	assert.NotNil(t, company)
	assert.Equal(t, company.Name, "hung")
}
