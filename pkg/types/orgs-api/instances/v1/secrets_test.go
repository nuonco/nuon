package instancesv1

import (
	"fmt"
	"strings"
	"testing"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
	"github.com/stretchr/testify/assert"
)

func TestAsRef(t *testing.T) {
	secret := Secret{
		OrgId:       shortid.New(),
		AppId:       shortid.New(),
		ComponentId: shortid.New(),
		InstallId:   shortid.New(),
		Id:          shortid.New(),
		Key:         "key1",
		Value:       "value1",
	}
	ref := secret.AsRef()
	assert.Equal(t, secret.OrgId, ref.OrgId)
	assert.Equal(t, secret.AppId, ref.AppId)
	assert.Equal(t, secret.ComponentId, ref.ComponentId)
	assert.Equal(t, secret.InstallId, ref.InstallId)
}

func TestS3Path(t *testing.T) {
	orgID := shortid.New()
	appID := shortid.New()
	compID := shortid.New()
	installID := shortid.New()
	secretID := shortid.New()
	secretRef := SecretRef{
		OrgId:       orgID,
		AppId:       appID,
		ComponentId: compID,
		InstallId:   installID,
		SecretId:    secretID,
	}
	s3Path := secretRef.S3Path()
	assert.True(t, strings.Contains(s3Path, "org="+orgID))
	assert.True(t, strings.Contains(s3Path, "app="+appID))
	assert.True(t, strings.Contains(s3Path, "component="+compID))
	assert.True(t, strings.Contains(s3Path, "install="+installID))
	assert.True(t, strings.Contains(s3Path, fmt.Sprintf("/secret-v1-%s.pb.enc", secretRef.SecretId)))
}
