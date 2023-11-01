package multi

import (
	"context"
	"fmt"

	pb "github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/mono/pkg/waypoint/client/k8s"
)

func (m *multiClient) getClient(ctx context.Context, id string) (pb.WaypointClient, error) {
	provider, err := k8s.New(m.v, k8s.WithConfig(k8s.Config{
		Address:     fmt.Sprintf(m.Config.AddressTemplate, id),
		ClusterInfo: m.Config.ClusterInfo,
		Token: k8s.Token{
			Namespace: fmt.Sprintf(m.Config.SecretNamespace),
			Name:      fmt.Sprintf(m.Config.SecretNameTemplate, id),
			Key:       fmt.Sprintf(m.Config.SecretKey),
		},
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to get provider: %w", err)
	}

	client, err := provider.Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch waypoint client: %w", err)
	}

	return client, nil
}
