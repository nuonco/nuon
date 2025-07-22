package credentials

type ServicePrincipalCredentials struct {
	SubscriptionID       string `cty:"subscription_id" json:"subscription_id" temporaljson:"subscription_id" hcl:"subscription_id" mapstructure:"subscription_id,omitempty"`
	SubscriptionTenantID string `cty:"subscription_tenant_id" json:"subscription_tenant_id" temporaljson:"subscription_tenant_id" hcl:"subscription_tenant_id" mapstructure:"subscription_tenant_id,omitempty"`
}

type Config struct {
	ServicePrincipal *ServicePrincipalCredentials `cty:"service_principal,block" hcl:"service_principal,block" mapstructure:"service_principal,omitempty" json:"service_principal" temporaljson:"service_principal"`
	UseDefault       bool                         `cty:"use_default,optional" hcl:"use_default,optional" mapstructure:"use_default,omitempty" json:"use_default" temporaljson:"use_default"`
}

func (c Config) String() string {
	if c.UseDefault {
		return "default credentials"
	}

	return "managed identity"
}
