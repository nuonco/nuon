package protos

import (
	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (c *Adapter) ToAzureSettings(install *app.Install) *executors.AzureSettings {
	if install.AzureAccount == nil {
		return nil
	}

	return &executors.AzureSettings{
		Location:                 install.AzureAccount.Location,
		SubscriptionID:           install.AzureAccount.SubscriptionID,
		SubscriptionTenantID:     install.AzureAccount.SubscriptionTenantID,
		ServicePrincipalAppID:    install.AzureAccount.ServicePrincipalAppID,
		ServicePrincipalPassword: install.AzureAccount.ServicePrincipalPassword,
	}
}
