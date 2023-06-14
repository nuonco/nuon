package instancesv1

import (
	"fmt"
	"strings"
	"testing"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/stretchr/testify/assert"
)

func TestAsRef(t *testing.T) {
	orgID := domains.NewOrgID()
	appID := domains.NewAppID()
	compID := domains.NewComponentID()
	installID := domains.NewInstallID()
	secretID := domains.NewSecretID()
	secret := Secret{
		OrgId:       orgID,
		AppId:       appID,
		ComponentId: compID,
		InstallId:   installID,
		Id:          secretID,
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
	orgID := domains.NewOrgID()
	appID := domains.NewAppID()
	compID := domains.NewComponentID()
	installID := domains.NewInstallID()
	secretID := domains.NewSecretID()
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
