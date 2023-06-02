package fakers

import (
	"reflect"
	"testing"

	installv1 "github.com/powertoolsdev/mono/pkg/types/api/install/v1"
	"github.com/stretchr/testify/assert"
)

func Test_api(t *testing.T) {
	settings, err := fakeAPIInstallAWSSettings(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, settings)

	awsSettings, ok := settings.(*installv1.UpsertInstallRequest_AwsSettings)
	assert.True(t, ok)
	assert.NoError(t, err)
	assert.NotNil(t, awsSettings)
}
