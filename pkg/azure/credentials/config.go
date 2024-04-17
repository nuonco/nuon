package credentials

type ServicePrincipalCredentials struct {
	SubscriptionID           string `cty:"subscription_id" json:"subscription_id" hcl:"subscription_id"`
	SubscriptionTenantID     string `cty:"subscription_tenant_id" json:"subscription_tenant_id" hcl:"subscription_tenant_id"`
	ServicePrincipalAppID    string `cty:"service_principal_app_id" json:"service_principal_app_id" hcl:"service_principal_id"`
	ServicePrincipalPassword string `cty:"service_principal_password" json:"service_principal_password" hcl:"service_principal_password"`
}

type Config struct {
	ServicePrincipal *ServicePrincipalCredentials `cty:"service_principal,block" hcl:"service_principal,block" mapstructure:"service_principal,omitempty"`
	UseDefault       bool                         `cty:"use_default,optional" hcl:"use_default,optional" mapstructure:"use_default,omitempty"`
}
