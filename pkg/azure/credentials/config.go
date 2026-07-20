package credentials

type ServicePrincipalCredentials struct {
	SubscriptionID       string `cty:"subscription_id" json:"subscription_id" temporaljson:"subscription_id" hcl:"subscription_id" mapstructure:"subscription_id,omitempty"`
	SubscriptionTenantID string `cty:"subscription_tenant_id" json:"subscription_tenant_id" temporaljson:"subscription_tenant_id" hcl:"subscription_tenant_id" mapstructure:"subscription_tenant_id,omitempty"`
}

type Config struct {
	ServicePrincipal *ServicePrincipalCredentials `cty:"service_principal,block" hcl:"service_principal,block" mapstructure:"service_principal,omitempty" json:"service_principal" temporaljson:"service_principal"`
	UseDefault       bool                         `cty:"use_default,optional" hcl:"use_default,optional" mapstructure:"use_default,omitempty" json:"use_default" temporaljson:"use_default"`

	// ManagedIdentityClientID runs the operation as a specific user-assigned
	// managed identity instead of the VM's system identity.
	ManagedIdentityClientID string `cty:"managed_identity_client_id,optional" hcl:"managed_identity_client_id,optional" mapstructure:"managed_identity_client_id,omitempty" json:"managed_identity_client_id" temporaljson:"managed_identity_client_id"`
}

func (c Config) String() string {
	if c.ManagedIdentityClientID != "" {
		return "user-assigned managed identity " + c.ManagedIdentityClientID
	}
	if c.UseDefault {
		return "default credentials"
	}

	return "managed identity"
}
