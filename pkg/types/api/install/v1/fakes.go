package installv1

import (
	"reflect"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
)

func fakeAPIInstallAWSSettings(v reflect.Value) (interface{}, error) {
	return &UpsertInstallRequest_AwsSettings{
		AwsSettings: &AwsSettings{
			Region: AwsRegion_AWS_REGION_US_WEST_1,
			Role:   uuid.NewString(),
		},
	}, nil
}

func init() {
	// api fakers
	_ = faker.AddProvider("apiInstallAWSSettings", fakeAPIInstallAWSSettings)
}
