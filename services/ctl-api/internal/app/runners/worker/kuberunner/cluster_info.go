package runner

import (
	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/kube"
)

func (w *wkflow) getClusterInfo() *kube.ClusterInfo {
	return &kube.ClusterInfo{
		ID:       w.cfg.OrgRunnerK8sClusterID,
		Endpoint: w.cfg.OrgRunnerK8sPublicEndpoint,
		CAData:   w.cfg.OrgRunnerK8sCAData,

		AWSAuth: &awscredentials.Config{
			Region: w.cfg.OrgRunnerRegion,
			AssumeRole: &awscredentials.AssumeRoleConfig{
				RoleARN:                w.cfg.OrgRunnerK8sIAMRoleARN,
				SessionName:            "ctl-api-runner-install",
				SessionDurationSeconds: 60 * 60,
			},
		},
	}
}
