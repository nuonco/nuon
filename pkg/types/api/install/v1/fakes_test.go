package installv1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_fakeAPIInstallAWSSettings(t *testing.T) {
	settings, err := fakeAPIInstallAWSSettings(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, settings)

	awsSettings, ok := settings.(*UpsertInstallRequest_AwsSettings)
	assert.True(t, ok)
	assert.NoError(t, err)
	assert.NotNil(t, awsSettings)
}
