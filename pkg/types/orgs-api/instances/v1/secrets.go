package instancesv1

import (
	"fmt"
	"path"

	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
)

func (s *Secret) AsRef() *SecretRef {
	return &SecretRef{
		OrgId:       s.OrgId,
		AppId:       s.AppId,
		ComponentId: s.ComponentId,
		InstallId:   s.InstallId,
		SecretId:    s.Id,
	}
}

func (s *SecretRef) S3Path() string {
	return path.Join(
		prefix.SecretsPath(s.OrgId, s.AppId, s.ComponentId, s.InstallId),
		fmt.Sprintf("secret-v1-%s.pb.enc", s.SecretId))
}
