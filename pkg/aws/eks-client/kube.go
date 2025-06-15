package eksclient

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/kube"
)

func (e *eksClient) GetClusterInfo(ctx context.Context) (*kube.ClusterInfo, error) {
	cluster, err := e.GetCluster(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster info")
	}

	return &kube.ClusterInfo{
		ID:       e.ClusterName,
		Endpoint: generics.FromPtrStr(cluster.Endpoint),
		CAData:   generics.FromPtrStr(cluster.CertificateAuthority.Data),
	}, nil
}
