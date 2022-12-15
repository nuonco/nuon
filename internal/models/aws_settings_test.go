// aws_settings.go
package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAWSRegion_ToRegion(t *testing.T) {
	assert.Equal(t, "us-east-1", AWSRegionUsEast1.ToRegion())
	assert.Equal(t, "us-east-2", AWSRegionUsEast2.ToRegion())
	assert.Equal(t, "us-west-2", AWSRegionUsWest2.ToRegion())
	assert.Equal(t, "us-west-1", AWSRegionUsWest1.ToRegion())
}
