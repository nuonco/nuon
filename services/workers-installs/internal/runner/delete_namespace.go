package runner

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/kube"
	awseks "github.com/powertoolsdev/mono/pkg/sandboxes/aws-eks"
	"github.com/powertoolsdev/mono/pkg/workflows/dal"
	"go.temporal.io/sdk/activity"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type DeleteNamespaceRequest struct {
	Namespace string `validate:"required"`
	OrgID     string `validate:"required"`
	AppID     string `validate:"required"`
	InstallID string `validate:"required"`
}

func (r DeleteNamespaceRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type DeleteNamespaceResponse struct{}

func (a *Activities) getKubeConfig(info *kube.ClusterInfo) (*rest.Config, error) {
	kCfg, err := kube.ConfigForCluster(info)
	if err != nil {
		return nil, fmt.Errorf("failed to get config for cluster: %w", err)
	}

	return kCfg, nil
}

func (a *Activities) DeleteNamespace(ctx context.Context, req DeleteNamespaceRequest) (DeleteNamespaceResponse, error) {
	resp := DeleteNamespaceResponse{}
	l := activity.GetLogger(ctx)

	tfOutputs, err := a.getSandboxOutputs(ctx, req.OrgID, req.AppID, req.InstallID)
	if err != nil {
		return resp, fmt.Errorf("unable to get sandbox outputs: %w", err)
	}

	kubeCfg, err := a.getKubeConfig(&kube.ClusterInfo{
		ID:             tfOutputs.Cluster.Name,
		Endpoint:       tfOutputs.Cluster.Endpoint,
		CAData:         tfOutputs.Cluster.CertificateAuthorityData,
		TrustedRoleARN: a.cfg.NuonAccessRoleArn,
	})
	if err != nil {
		return resp, fmt.Errorf("unable to get kube config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return resp, fmt.Errorf("failed to create kube client: %w", err)
	}

	err = a.deleteNamespace(ctx, clientset.CoreV1().Namespaces(), req.Namespace)
	if err != nil {
		return resp, fmt.Errorf("failed to destroy namespace: %w", err)
	}

	l.Debug("finished destroying namespace", "response", resp)
	return resp, nil
}

func (a *Activities) getSandboxOutputs(ctx context.Context, orgID, appID, installID string) (*awseks.TerraformOutputs, error) {
	dalClient, err := dal.New(a.v,
		dal.WithOrgID(orgID),
		dal.WithSettings(dal.Settings{
			InstallsBucket:                a.cfg.InstallationsBucket,
			InstallsBucketIAMRoleTemplate: a.cfg.OrgInstallationsRoleTemplate,
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to get dal client: %w", err)
	}

	response, err := dalClient.GetInstallSandboxOutputs(ctx, orgID, appID, installID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install provision response: %w", err)
	}

	tfOutputs, err := awseks.ParseTerraformOutputs(response)
	if err != nil {
		return nil, fmt.Errorf("unable to parse terraform outputs: %w", err)
	}
	return &tfOutputs, nil
}

func (a *Activities) deleteNamespace(ctx context.Context, api corev1.NamespaceInterface, name string) error {
	err := api.Delete(ctx, name, apimetav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace ns: %s: %w", name, err)
	}

	return nil
}
