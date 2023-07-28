package provider

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/api/gqlclient"
)

func stringToAPIRegion(val string) (gqlclient.AWSRegion, error) {
	switch val {
	case "us-east-1":
		return gqlclient.AWSRegionUsEast1, nil
	case "us-east-2":
		return gqlclient.AWSRegionUsEast2, nil
	case "us-west-1":
		return gqlclient.AWSRegionUsWest1, nil
	case "us-west-2":
		return gqlclient.AWSRegionUsWest2, nil
	}

	return gqlclient.AWSRegion(""), fmt.Errorf("region not supported: %s", val)
}
