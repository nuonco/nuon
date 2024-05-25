package waypoint

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/kube"
	"github.com/powertoolsdev/mono/pkg/waypoint/client/multi"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/terraformcloud"
)

func New(v *validator.Validate, orgsOutputs *terraformcloud.OrgsOutputs) (multi.Client, error) {
	client, err := multi.New(v, multi.WithConfig(&multi.Config{
		AddressTemplate:    "%s." + orgsOutputs.Waypoint.RootDomain + ":9701",
		SecretNameTemplate: orgsOutputs.Waypoint.TokenSecretTemplate,
		SecretNamespace:    orgsOutputs.Waypoint.TokenSecretNamespace,
		SecretKey:          "token",
		ClusterInfo: &kube.ClusterInfo{
			ID:             orgsOutputs.Waypoint.ClusterID,
			Endpoint:       orgsOutputs.Waypoint.PublicEndpoint,
			CAData:         orgsOutputs.Waypoint.CAData,
			TrustedRoleARN: orgsOutputs.K8s.AccessRoleARNs["eks-ctl-api"],
		},
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to create multi client: %w", err)
	}

	return client, nil
}
