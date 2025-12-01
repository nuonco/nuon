package deployv1

import (
	"reflect"
	"time"

	"github.com/go-faker/faker/v4"
	"google.golang.org/protobuf/types/known/durationpb"
)

func fakeDeployConfig(v reflect.Value) (interface{}, error) {
	return &Config{
		Timeout: durationpb.New(time.Second * 10),
		Cfg: &Config_Basic{
			Basic: &BasicConfig{
				InstanceCount: 1,
				ListenerCfg: &ListenerConfig{
					ListenPort:      80,
					HealthCheckPath: "/",
				},
			},
		},
	}, nil
}

//nolint:gochecknoinits
func init() {
	_ = faker.AddProvider("deployConfig", fakeDeployConfig)
}
