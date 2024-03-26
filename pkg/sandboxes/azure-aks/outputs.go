package azureaks

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/protobuf/types/known/structpb"
)

// NOTE: structpb does not support []string type, so we have to use interface{} here
func ToStringSlice(vals []interface{}) []string {
	strVals := make([]string, len(vals))
	for idx, val := range vals {
		v := val
		strVals[idx] = v.(string)
	}

	return strVals
}

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

type VPNOutputs struct {
	Name      string        `mapstructure:"name" validate:"required"`
	SubnetIDs []interface{} `mapstructure:"subnet_ids" validate:"required" faker:"stringSliceAsInt"`
}

type AccountOutputs struct {
	SubscriptionID string `mapstructure:"subscription_id" validate:"required"`
	Location       string `mapstructure:"location" validate:"required"`
}

type ACROutputs struct {
	ID          string `mapstructure:"id" validate:"required"`
	LoginServer string `mapstructure:"login_server" validate:"required"`
	TokenID     string `mapstructure:"token_id" validate:"required"`
	Password    string `mapstructure:"password" validate:"required"`
}

type RunnerOutputs struct{}

type DomainOutputs struct {
	Nameservers []interface{} `mapstructure:"nameservers" validate:"required" faker:"stringSliceAsInt"`
	Name        string        `mapstructure:"name" validate:"required" faker:"domain"`
	ZoneID      string        `mapstructure:"zone_id" validate:"required"`
}

type TerraformOutputs struct {
	// domain outputs
	//PublicDomain	 DomainOutputs `mapstructure:"public_domain"`
	//InternalDomain DomainOutputs `mapstructure:"internal_domain"`

	Cluster ClusterOutputs `mapstructure:"cluster"`
	ACR     ACROutputs     `mapstructure:"acr"`
	VPN     VPNOutputs     `mapstructure:"vpn"`
	Runner  RunnerOutputs  `mapstructure:"runner"`
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
