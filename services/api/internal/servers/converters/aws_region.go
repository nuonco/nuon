package converters

import (
	"github.com/powertoolsdev/mono/services/api/internal/models"
	installv1 "github.com/powertoolsdev/mono/pkg/protos/api/generated/types/install/v1"
)

func ProtoToAwsRegion(inpRegion installv1.AwsRegion) models.AWSRegion {
	switch inpRegion {
	case installv1.AwsRegion_AWS_REGION_US_EAST_1:
		return models.AWSRegionUsEast1
	case installv1.AwsRegion_AWS_REGION_US_WEST_1:
		return models.AWSRegionUsWest1
	case installv1.AwsRegion_AWS_REGION_US_WEST_2:
		return models.AWSRegionUsWest2
	case installv1.AwsRegion_AWS_REGION_US_EAST_2:
		return models.AWSRegionUsEast2
	}

	return models.AWSRegionUsWest2
}

func AwsRegionToProto(inpRegion models.AWSRegion) installv1.AwsRegion {
	switch inpRegion {
	case models.AWSRegionUsEast1:
		return installv1.AwsRegion_AWS_REGION_US_EAST_1
	case models.AWSRegionUsWest1:
		return installv1.AwsRegion_AWS_REGION_US_WEST_1
	case models.AWSRegionUsWest2:
		return installv1.AwsRegion_AWS_REGION_US_WEST_2
	case models.AWSRegionUsEast2:
		return installv1.AwsRegion_AWS_REGION_US_EAST_2
	}

	return installv1.AwsRegion_AWS_REGION_UNSPECIFIED
}
