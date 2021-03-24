package titan

import (
	"fmt"
	"runtime/debug"
	"time"

	consulapi "github.com/hashicorp/consul/api"
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

	viper.SetConfigType("yaml") // Need to explicitly set this to json
	err = viper.ReadRemoteConfig()
	if err != nil && err.Error() == "Remote Configurations Error: No Files Found" {
		config := consulapi.DefaultConfig()
		consul, err := consulapi.NewClient(config)
		if err != nil {
			return errors.WithMessagef(err, "err connect to consul, %s", consulHost)
		}

		_, err = consul.KV().Put(&consulapi.KVPair{
			Key:   subject,
			Value: []byte(""),
			Flags: 0,
		}, nil)

		if err != nil {
			return errors.WithMessagef(err, "err put to consul %s", consulHost)
		}
	}

	if err != nil {
		return errors.WithMessagef(err, "err read config %s", consulHost)
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				ctx.Logger().Error(fmt.Sprintf("Panicking %s \n", debug.Stack()))
			}
		}()
		ticker := time.NewTicker(time.Minute * 1)
		for {
			<-ticker.C
			err := viper.WatchRemoteConfig()
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("unable to read remote config: %v", err))
				continue
			}
		}
	}()

	return nil
}
