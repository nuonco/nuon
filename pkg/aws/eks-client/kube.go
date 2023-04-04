package eksclient

import (
	"context"
	"encoding/base64"
	"fmt"

	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

func (e *eksClient) GetKubeConfig(ctx context.Context) (*rest.Config, error) {
	cluster, err := e.GetCluster(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get cluster: %w", err)
	}

	kubeCfg, err := e.getKubeConfig(cluster)
	if err != nil {
		return nil, fmt.Errorf("unable to get kube config: %w", err)
	}

	return kubeCfg, nil
}

func (e *eksClient) getKubeConfig(cluster *ekstypes.Cluster) (*rest.Config, error) {
	gen, err := token.NewGenerator(true, false)
	if err != nil {
		return nil, err
	}

	opts := &token.GetTokenOptions{
		Region:        e.Region,
		AssumeRoleARN: e.RoleARN,
		SessionName:   e.RoleSessionName,
		ClusterID:     *cluster.Name,
	}
	tok, err := gen.GetWithOptions(opts)
	if err != nil {
		return nil, err
	}

	ca, err := base64.StdEncoding.DecodeString(*cluster.CertificateAuthority.Data)
	if err != nil {
		return nil, err
	}

	return &rest.Config{
		Host:        *cluster.Endpoint,
		BearerToken: tok.Token,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: ca,
		},
	}, nil
}
