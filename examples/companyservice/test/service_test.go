package test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gitlab.com/silenteer-oss/titan"

	"gitlab.com/silenteer-oss/titan/examples/companyservice/api"

	"gitlab.com/silenteer-oss/titan/examples/companyservice/internal/app"

	"github.com/stretchr/testify/assert"
)

var natsClient = titan.GetDefaultClient()
var companyService = api.NewCompanyClient(natsClient)

func TestMain(m *testing.M) {
	server := app.NewServer()

	go func() {
		server.Start()
	}()

	time.Sleep(1 * time.Second)

	exitVal := m.Run()

	server.Stop()
	os.Exit(exitVal)
}

func TestGetCompanies(t *testing.T) {
	context := titan.NewContext(context.Background())
	companies, err := companyService.GetCompanies(context)

	require.NoError(t, err, fmt.Sprintf("Get Companies error: %+v\n ", err))

	assert.Nil(t, err)
	assert.Equal(t, len(*companies), 1)

}

func TestGetNotExistCompany(t *testing.T) {
	context := titan.NewContext(context.Background())

	company, err := companyService.GetCompany(context, "not_exist")

	require.NoError(t, err, fmt.Sprintf("Get Company error: %+v\n ", err))

	assert.Nil(t, company)
}

func TestGetExistCompany(t *testing.T) {
	context := titan.NewContext(context.Background())

	company, err := companyService.GetCompany(context, "hung")
	require.NoError(t, err, fmt.Sprintf("Get Company error: %+v\n ", err))
	require.NotNil(t, company)
	assert.Equal(t, company.Name, "hung")
}
