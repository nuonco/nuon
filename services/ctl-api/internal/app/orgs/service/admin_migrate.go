package service

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	sigs "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
	orgreprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/v2/reprovision"
	orgrestart "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/v2/restart"
	runnersigs "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type AdminMigrateOrg struct{}

// @ID						AdminMigrateOrg
// @Summary				migrate an org
// @Description.markdown	admin_migrate_org.md
// @Param					org_id	path	string	true	"org ID or name to update"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req	body	AdminMigrateOrg	true	"Input"
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/orgs/{org_id}/admin-migrate [POST]
func (s *service) AdminMigrateOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	org, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	cctx.SetOrgGinContext(ctx, org)
	cctx.SetAccountGinContext(ctx, &org.CreatedBy)

	var req AdminMigrateOrg
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if err := s.adminMigrateOrg(ctx, org); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, true)
}

func (s *service) adminMigrateOrg(ctx context.Context, org *app.Org) error {
	// create runner role
	role := app.Role{
		OrgID:    generics.NewNullString(org.ID),
		RoleType: app.RoleTypeRunner,
		Policies: []app.Policy{
			{
				OrgID: generics.NewNullString(org.ID),
				Name:  "admin",
				Permissions: pgtype.Hstore(map[string]*string{
					org.ID: permissions.PermissionAll.ToStrPtr(),
				}),
			},
		},
	}
	res := s.db.
		WithContext(ctx).
		Create(&role)
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to create role")
	}

	// update org type to default
	res = s.db.WithContext(ctx).Model(org).Updates(app.Org{
		OrgType: app.OrgTypeDefault,
	})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update org")
	}
	if res.RowsAffected != 1 {
		return errors.Wrap(gorm.ErrRecordNotFound, "org not found")
	}

	// create org runner group
	rg, err := s.runnersHelpers.CreateOrgRunnerGroup(ctx, org)
	if err != nil {
		return errors.Wrap(err, "unable to create org runner group")
	}

	// Runner signals stay v1 - runners don't have v2 queues
	s.evClient.Send(ctx, rg.Runners[0].ID, &runnersigs.Signal{
		Type: sigs.OperationCreated,
	})
	s.evClient.Send(ctx, rg.Runners[0].ID, &runnersigs.Signal{
		Type: runnersigs.OperationProvision,
	})

	// Org signals use v2 queues if enabled
	useQueues, err := s.useOrgQueues(ctx, org.ID)
	if err != nil {
		return fmt.Errorf("checking features: %w", err)
	}
	if useQueues {
		queueID, err := s.getOrgSignalsQueueID(ctx, org.ID)
		if err != nil {
			return fmt.Errorf("unable to get org signals queue: %w", err)
		}
		if err := s.enqueueOrgSignal(ctx, queueID, &orgrestart.Signal{OrgID: org.ID}, org.ID); err != nil {
			return fmt.Errorf("enqueue restart signal: %w", err)
		}
		if err := s.enqueueOrgSignal(ctx, queueID, &orgreprovision.Signal{OrgID: org.ID}, org.ID); err != nil {
			return fmt.Errorf("enqueue reprovision signal: %w", err)
		}
	} else {
		s.evClient.Send(ctx, org.ID, &sigs.Signal{
			Type: sigs.OperationRestart,
		})
		s.evClient.Send(ctx, org.ID, &sigs.Signal{
			Type: sigs.OperationReprovision,
		})
	}
	return nil
}
