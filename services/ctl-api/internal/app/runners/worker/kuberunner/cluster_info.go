package runner

import (
	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"
	"github.com/nuonco/nuon/pkg/kube"
)

func (w *Wkflow) getClusterInfo() *kube.ClusterInfo {
	clusterInfo := &kube.ClusterInfo{
		ID:       w.cfg.OrgRunnerK8sClusterID,
		Endpoint: w.cfg.OrgRunnerK8sPublicEndpoint,
		CAData:   w.cfg.OrgRunnerK8sCAData,
	}

	if w.cfg.IsGCP() {
		clusterInfo.GCPAuth = &gcpcredentials.Config{
			ProjectID: w.cfg.ManagementAccountID,
			Region:    w.cfg.OrgRunnerRegion,
		}
	} else if w.cfg.IsAzure() {
		clusterInfo.AzureAuth = &azurecredentials.Config{
			UseDefault: true,
			ServicePrincipal: &azurecredentials.ServicePrincipalCredentials{
				SubscriptionTenantID: w.cfg.ManagementAzureTenantID,
			},
		}
	} else if w.cfg.OrgRunnerK8sUseDefaultCreds {
		clusterInfo.AWSAuth = &awscredentials.Config{
			Region:     w.cfg.OrgRunnerRegion,
			UseDefault: true,
		}
	} else {
		clusterInfo.AWSAuth = &awscredentials.Config{
			Region: w.cfg.OrgRunnerRegion,
			AssumeRole: &awscredentials.AssumeRoleConfig{
				RoleARN:                w.cfg.OrgRunnerK8sIAMRoleARN,
				SessionName:            "ctl-api-runner-install",
				SessionDurationSeconds: 60 * 60,
			},
		}
	}

	return clusterInfo
}
