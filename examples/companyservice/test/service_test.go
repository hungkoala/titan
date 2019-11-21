package test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"gitlab.com/silenteer/titan/examples/companyservice/api"

	"gitlab.com/silenteer/titan/examples/companyservice/internal/app"

	"github.com/stretchr/testify/assert"

	"gitlab.com/silenteer/titan/nats"
)

var config = app.DefaultConfig()
var natsClient = nats.NewClient(config.Nats)
var companyService = api.NewCompanyClient(natsClient)

func TestMain(m *testing.M) {
	server := app.NewServer(config)

	go func() {
		server.Start()
	}()

	time.Sleep(2 * time.Millisecond)

	exitVal := m.Run()

	server.Stop()
	os.Exit(exitVal)
}

func TestGetCompanies(t *testing.T) {
	context := nats.NewContext(context.Background())
	companies, err := companyService.GetCompanies(context)
	if err != nil {
		fmt.Println(fmt.Sprintf("get companies error: %+v\n ", err))
	}

	assert.Nil(t, err)
	assert.Equal(t, len(*companies), 1)

}

func TestGetNotExistCompany(t *testing.T) {
	context := nats.NewContext(context.Background())

	company, err := companyService.GetCompany(context, "not_exist")
	if err != nil {
		fmt.Println(fmt.Sprintf("get company error: %+v\n ", err))
	}

	assert.Nil(t, err)
	assert.Nil(t, company)
}

func TestGetExistCompany(t *testing.T) {
	context := nats.NewContext(context.Background())

	company, err := companyService.GetCompany(context, "hung")
	if err != nil {
		fmt.Println(fmt.Sprintf("Get Company error: %+v\n ", err))
	}

	assert.Nil(t, err)
	assert.NotNil(t, company)
	assert.Equal(t, company.Name, "hung")
}
