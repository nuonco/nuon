package credentials

type ServicePrincipalCredentials struct {
	SubscriptionID           string `json:"subscription_id" gorm:"not null;default null"`
	SubscriptionTenantID     string `json:"subscription_tenant_id" gorm:"not null;default null"`
	ServicePrincipalAppID    string `json:"service_principal_app_id" gorm:"not null;default null"`
	ServicePrincipalPassword string `json:"service_principal_password" gorm:"not null;default null"`
}

type Config struct {
	ServicePrincipal *ServicePrincipalCredentials `cty:"service_principal,block" hcl:"service_principal,block" mapstructure:"service_principal,omitempty"`
	UseDefault       bool                         `cty:"use_default,optional" hcl:"use_default,optional" mapstructure:"use_default,omitempty"`
}
