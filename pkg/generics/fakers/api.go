package fakers

import (
	"reflect"

	"github.com/google/uuid"
	installv1 "github.com/powertoolsdev/mono/pkg/types/api/install/v1"
)

func fakeAPIInstallAWSSettings(v reflect.Value) (interface{}, error) {
	return &installv1.UpsertInstallRequest_AwsSettings{
		AwsSettings: &installv1.AwsSettings{
			Region: installv1.AwsRegion_AWS_REGION_US_WEST_1,
			Role:   uuid.NewString(),
		},
	}, nil
}
