package waypoint

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/kube"
	"github.com/powertoolsdev/mono/pkg/waypoint/client/multi"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/terraformcloud"
	"go.uber.org/zap"
)

func New(v *validator.Validate, orgsOutputs *terraformcloud.OrgsOutputs, l *zap.Logger) (multi.Client, error) {
	l.Error("test", zap.Any("test", orgsOutputs.Waypoint))
	client, err := multi.New(v, multi.WithConfig(&multi.Config{
		AddressTemplate:    "%s." + orgsOutputs.Waypoint.RootDomain,
		SecretNameTemplate: orgsOutputs.Waypoint.TokenSecretNamespace,
		SecretNamespace:    orgsOutputs.Waypoint.TokenSecretNamespace,
		SecretKey:          "data",
		ClusterInfo: &kube.ClusterInfo{
			ID:             orgsOutputs.Waypoint.ClusterID,
			Endpoint:       orgsOutputs.Waypoint.PublicEndpoint,
			CAData:         orgsOutputs.Waypoint.CAData,
			TrustedRoleARN: orgsOutputs.K8s.AccessRoleARNs["ctl-api"],
		},
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to create multi client: %w", err)
	}

	return client, nil
}
