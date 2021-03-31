package titan_test

import (
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/silenteer-oss/titan"
)

var yamlExample = []byte(`
Hacker: true
name: steve
hobbies:
- skateboarding
- snowboarding
- go
clothing:
  jacket: leather
  trousers: denim
age: 35
eyes : brown
beard: true
`)

func TestConsulRemoteConfig(t *testing.T) {
	key := "api.service.config"

	assert.Empty(t, viper.GetString("name"))
	assert.Equal(t, 0, viper.GetInt("age"))
	assert.False(t, viper.GetBool("beard"))
	assert.Nil(t, viper.GetStringMap("clothing")["jacket"])
	assert.Empty(t, viper.GetString("clothing.trousers"))

	setValueToConsul(t, key)

	newServer := titan.NewServer(key)
	go newServer.Start()
	t.Cleanup(func() {
		newServer.Stop()
	})

	time.Sleep(2 * time.Second)
	assert.Equal(t, "steve", viper.GetString("name"))
	assert.Equal(t, 35, viper.GetInt("age"))
	assert.Equal(t, true, viper.GetBool("beard"))
	assert.Equal(t, "leather", viper.GetStringMap("clothing")["jacket"])
	assert.Equal(t, "denim", viper.GetString("clothing.trousers"))
}

func setValueToConsul(t *testing.T, key string) {
	t.Helper()
	config := api.DefaultConfig()

	consul, err := api.NewClient(config)
	if err != nil {
		require.Nil(t, err)
	}

	data, err := consul.KV().Put(&api.KVPair{
		Key:   key,
		Value: yamlExample,
		Flags: 0,
	}, nil)

	require.NotNil(t, data)
	require.Nil(t, err)

	t.Cleanup(func() {
		data, err := consul.KV().Delete(key, &api.WriteOptions{})
		require.NotNil(t, data)
		require.Nil(t, err)
	})
}
