package executors

type AzureSettings struct {
	Location                 string `json:"string"`
	SubscriptionID           string `json:"subscription_id"`
	SubscriptionTenantID     string `json:"subscription_tenant_id"`
	ServicePrincipalAppID    string `json:"service_principal_app_id"`
	ServicePrincipalPassword string `json:"service_principal_password"`
}
