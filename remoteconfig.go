package titan

import (
	"fmt"
	"runtime/debug"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	_ "github.com/spf13/viper/remote"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func InitRemoteConfig(subject string) error {
	consulHost := viper.GetString(ConsulAddr)

	ctx := NewBackgroundContext()
	err := viper.AddRemoteProvider("consul", consulHost, fmt.Sprintf("/%s", subject))
	if err != nil {
		return errors.WithMessagef(err, "err config consul")
	}

	ctx.Logger().Info(fmt.Sprintf("Connect to consul at: %s, folder: %s\n", consulHost, subject))
	viper.SetConfigType("yaml")
	err = viper.ReadRemoteConfig()
	if err != nil {
		if err.Error() != "Remote Configurations Error: No Files Found" {
			return errors.WithMessagef(err, "err read remote config")
		}

		err = createFolder(subject, consulHost)
		if err != nil {
			return errors.WithMessagef(err, "err create consul folder %s", subject)
		}
	}

	wp, err := watch.Parse(map[string]interface{}{
		"type": "key",
		"key":  subject,
	})
	if err != nil {
		return errors.WithMessagef(err, "err Parse config")
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				ctx.Logger().Error(fmt.Sprintf("Panicking %s \n", debug.Stack()))
			}
		}()
		wp.Handler = func(idx uint64, data interface{}) {
			switch d := data.(type) {
			case *api.KVPair:
				ctx.Logger().Info(fmt.Sprintf("Folder %s changed Value: %s", d.Key, maskLeft(d.Value)))
			}
			err := viper.WatchRemoteConfig()
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("unable to read remote config: %v", err))
			}
		}

		wp.Run(consulHost)
	}()

	return nil
}

func createFolder(subject string, consulHost string) error {
	config := api.DefaultConfig()
	config.Address = consulHost
	consul, err := api.NewClient(config)
	if err != nil {
		return errors.WithMessagef(err, "err connect to consul, %s", consulHost)
	}

	_, err = consul.KV().Put(&api.KVPair{
		Key:   subject,
		Value: []byte(""),
		Flags: 0,
	}, nil)

	if err != nil {
		return errors.WithMessagef(err, "err put to consul %s", consulHost)
	}

	return nil
}

func maskLeft(s []byte) string {
	rs := s
	for i := 0; i < len(rs)-4; i++ {
		rs[i] = 'X'
	}
	return string(rs)
}
