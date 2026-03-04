package seed

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

func (s *Seeder) EnsureQueue(ctx context.Context, t *testing.T) *app.Queue {
	q, err := s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:   generics.GetFakeObj[string](),
		OwnerType: "test",
		Namespace: "default",
		MaxDepth:  100,
	})
	require.Nil(t, err)
	require.NotNil(t, q)

	return q
}
