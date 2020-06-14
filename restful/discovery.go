package restful

import "github.com/spf13/viper"

type Discovery interface {
	LookupService(serviceName string) (string, error)
}

type EnvDiscovery struct {
}

func NewEnvDiscovery() *EnvDiscovery {
	return &EnvDiscovery{}
}

//get value from environment variables
func (d *EnvDiscovery) LookupService(serviceName string) (string, error) {
	add := viper.GetString(serviceName)
	return add, nil
}

// consul service discovery, not implemented yet
type ConsulDiscovery struct {
}
