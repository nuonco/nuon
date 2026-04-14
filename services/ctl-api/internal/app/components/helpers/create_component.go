package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/pkg/errors"
)

type CreateComponentParams struct {
	AppID        string
	Name         string
	VarName      string
	Dependencies []string
}

func (h *Helpers) CreateComponent(ctx context.Context, params *CreateComponentParams) (*app.Component, error) {
	component := app.Component{
		AppID:             params.AppID,
		Name:              params.Name,
		VarName:           params.VarName,
		Status:            "queued",
		StatusDescription: "waiting for event loop to start for component",
	}
	res := h.db.WithContext(ctx).Create(&component)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create component: %w", res.Error)
	}

	_, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     component.ID,
		OwnerType:   plugins.TableName(h.db, app.Component{}),
		Namespace:   "components",
		MaxInFlight: 1,
		MaxDepth:    50,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create queue for component: %w", err)
	}

	depIDs, err := h.GetComponentIDs(ctx, params.AppID, params.Dependencies)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component ids")
	}
	if err := h.CreateComponentDependencies(ctx, component.ID, depIDs); err != nil {
		return nil, fmt.Errorf("unable to create component dependencies: %w", err)
	}

	if err := h.EnsureInstallComponents(ctx, params.AppID, nil); err != nil {
		return nil, fmt.Errorf("unable to ensure install components: %w", err)
	}

	return &component, nil
}
