package restful

import "gitlab.com/silenteer-oss/titan"

func NewRestClient() *titan.Client {
	discovery := NewEnvDiscovery()
	conn := NewConnection(discovery)
	return titan.NewClient(conn)
}
