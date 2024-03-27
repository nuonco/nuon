package azureaks

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/powertoolsdev/mono/pkg/sandboxes"
	"google.golang.org/protobuf/types/known/structpb"
)

type ClusterOutputs struct {
	ID                   string `mapstructure:"id"`
	Name                 string `mapstructure:"name"`
	ClientCertificate    string `mapstructure:"client_certificate"`
	ClientKey            string `mapstructure:"client_key"`
	ClusterCACertificate string `mapstructure:"cluster_ca_certificate"`
	ClusterFQDN          string `mapstructure:"cluster_fqdn"`
	OIDCIssuerURL        string `mapstructure:"oidc_issuer_url"`
	Location             string `mapstructure:"location"`
	KubeConfigRaw        string `mapstructure:"kube_config_raw"`
	KubeAdminConfigRaw   string `mapstructure:"kube_admin_config_raw"`
}

type RunnerOutputs struct{}

type TerraformOutputs struct {
	// domain outputs
	PublicDomain   sandboxes.DomainOutputs `mapstructure:"public_domain"`
	InternalDomain sandboxes.DomainOutputs `mapstructure:"internal_domain"`

	Cluster ClusterOutputs       `mapstructure:"cluster"`
	ACR     sandboxes.ACROutputs `mapstructure:"acr"`
	VPN     sandboxes.VPNOutputs `mapstructure:"vpn"`
	Runner  RunnerOutputs        `mapstructure:"runner"`
}

func (t *TerraformOutputs) Validate() error {
	validate := validator.New()
	return validate.Struct(t)
}

func ParseTerraformOutputs(outputs *structpb.Struct) (TerraformOutputs, error) {
	m := outputs.AsMap()

	var tfOutputs TerraformOutputs
	if err := mapstructure.Decode(m, &tfOutputs); err != nil {
		return tfOutputs, fmt.Errorf("invalid terraform outputs: %w", err)
	}

	err := tfOutputs.Validate()
	if err != nil {
		return tfOutputs, fmt.Errorf("terraform output error: %w", err)
	}

	return tfOutputs, nil
}
