package restful

import (
	"gitlab.com/silenteer-oss/titan"
)

func NewClient(add string) *titan.Client {
	conn := NewConnection(add)
	return titan.NewClient(conn)
}
