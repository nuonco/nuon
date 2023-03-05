package fakers

import (
	"reflect"
	"time"

	deployv1 "github.com/powertoolsdev/protos/components/generated/types/deploy/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

func fakeDeployConfig(v reflect.Value) (interface{}, error) {
	return &deployv1.Config{
		//nolint:all
		Timeout: durationpb.New(time.Second * 10),
		Cfg: &deployv1.Config_Basic{
			Basic: &deployv1.BasicConfig{
				InstanceCount: 1,
				ListenerCfg: &deployv1.ListenerConfig{
					ListenPort:      80,
					HealthCheckPath: "/",
				},
			},
		},
	}, nil
}
