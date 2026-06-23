package activities

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type SaveStateRequest struct {
	State *state.State `validate:"required"`

	InstallID       string                         `validate:"required"`
	TriggeredByID   string                         `validate:"required"`
	TriggeredByType string                         `validate:"required"`
	GeneratedBy     app.InstallStateGenerateSource `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) SaveState(ctx context.Context, req *SaveStateRequest) (*app.InstallState, error) {
	// the blob upload in InstallState's BeforeCreate hook requires org_id on the context
	if keys.OrgIDFromContext(ctx) == "" {
		var install app.Install
		if res := a.db.WithContext(ctx).Select("org_id").First(&install, "id = ?", req.InstallID); res.Error != nil {
			return nil, errors.Wrap(res.Error, "unable to look up install org for state")
		}
		ctx = cctx.SetOrgIDContext(ctx, install.OrgID)
	}

	stateJSON, err := json.Marshal(req.State)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal install state for blob")
	}

	obj := &app.InstallState{
		InstallID:       req.InstallID,
		TriggeredByID:   req.TriggeredByID,
		TriggeredByType: req.TriggeredByType,
		State:           req.State,
		GeneratedBy:     req.GeneratedBy,
		StateBlob:       &blobstore.Blob{},
	}
	obj.StateBlob.Set(string(stateJSON))

	res := a.db.WithContext(ctx).
		Create(&obj)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create install state")
	}
	return obj, nil
}
