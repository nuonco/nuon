package protos

import (
	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (a *Adapter) ToAWSSettings(install *app.Install) *executors.AWSSettings {
	if install.AWSAccount == nil {
		return nil
	}

	settings := &executors.AWSSettings{
		Region:                    install.AWSAccount.Region,
		IAMRoleARN:                install.AWSAccount.IAMRoleARN,
		AWSRoleDelegationSettings: &executors.AWSRoleDelegationSettings{},
	}
	if install.AppSandboxConfig.AWSDelegationConfig != nil {
		settings.AWSRoleDelegationSettings = &executors.AWSRoleDelegationSettings{
			IAMRoleARN:      install.AppSandboxConfig.AWSDelegationConfig.IAMRoleARN,
			AccessKeyID:     install.AppSandboxConfig.AWSDelegationConfig.AccessKeyID,
			SecretAccessKey: install.AppSandboxConfig.AWSDelegationConfig.SecretAccessKey,
		}
	}

	return settings
}
