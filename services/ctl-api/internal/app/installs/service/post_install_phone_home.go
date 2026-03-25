package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type InstallPhoneHomeRequest map[string]any

type phoneHomeRequestType string

const (
	phoneHomeRequestTypeUpdate = "Update"
	phoneHomeRequestTypeDelete = "Delete"
	phoneHomeRequestTypeCreate = "Create"
)

// @ID PhoneHome
// @Summary				phone home for an install
// @Description.markdown phone_home.md
// @Param					install_id	path	string		true	"install ID"
// @Param					phone_home_id	path	string		true	"phone home ID"
// @Param					req		body	InstallPhoneHomeRequest	true	"Input"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{string}	ok
// @Router					/v1/installs/{install_id}/phone-home/{phone_home_id} [post]
func (s *service) InstallPhoneHome(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	phoneHomeID := ctx.Param("phone_home_id")

	var req InstallPhoneHomeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	var requestType string
	if v, ok := req["request_type"]; ok {
		requestType = v.(string)
	} else {
		ctx.Error(fmt.Errorf("request type param not present"))
		return
	}

	switch requestType {
	case phoneHomeRequestTypeCreate, phoneHomeRequestTypeUpdate, phoneHomeRequestTypeDelete:
	default:
		ctx.Error(fmt.Errorf("request type param not present"))
		return
	}

	if requestType == phoneHomeRequestTypeDelete {
		ctx.JSON(http.StatusOK, "ok")
		return
	}

	if err := s.updateInstallPhoneHome(ctx, installID, phoneHomeID, &req); err != nil {
		ctx.Error(errors.Wrap(err, "unable to update install phone home"))
		return
	}

	ctx.JSON(http.StatusCreated, "ok")
}

func (s *service) updateInstallPhoneHome(ctx context.Context, installID, phoneHomeID string, req *InstallPhoneHomeRequest) error {
	var stackVersion app.InstallStackVersion
	if res := s.db.WithContext(ctx).
		Where(app.InstallStackVersion{
			InstallID:   installID,
			PhoneHomeID: phoneHomeID,
		}).
		First(&stackVersion); res.Error != nil {
		return errors.Wrap(res.Error, "unable to find cloudformation stack")
	}

	data, err := pkggenerics.ToMapstructureWithJSONTag(req)
	if err != nil {
		return errors.Wrap(err, "unable to convert to mapstructure")
	}

	// now make updates
	updatedStack := app.InstallStackVersion{
		ID: stackVersion.ID,
	}
	res := s.db.WithContext(ctx).
		Model(&updatedStack).
		Updates(app.InstallStackVersion{
			Status: app.NewCompositeStatus(ctx, app.InstallStackVersionStatusActive),
			Runs: []app.InstallStackVersionRun{
				{
					Data: generics.ToHstore(pkggenerics.ToStringMap(data)),
				},
			},
		})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update stack version")
	}
	if res.RowsAffected != 1 {
		return errors.Wrap(gorm.ErrRecordNotFound, "cloudformation stack not found")
	}

	run := app.InstallStackVersionRun{
		OrgID:                 stackVersion.OrgID,
		CreatedByID:           stackVersion.CreatedByID,
		InstallStackVersionID: stackVersion.ID,
		Data:                  generics.ToHstore(pkggenerics.ToStringMap(data)),
	}
	if res = s.db.WithContext(ctx).
		Create(&run); res.Error != nil {
		return errors.Wrap(res.Error, "unable to create install stack version run")
	}

	// Only send UpdateInstallStackOutputs signal if this is a stack param update (existing runs present meaning,
	// its an stack update on existing install and not part of provision / reprovision flow).
	// For the first phone home during provisioning, the provision/reprovision workflow step handles this.
	var existingRunCount int64
	if res = s.db.WithContext(ctx).
		Model(&app.InstallStackVersionRun{}).
		Where("install_stack_version_id = ?", stackVersion.ID).
		Count(&existingRunCount); res.Error != nil {
		return errors.Wrap(res.Error, "unable to count existing runs")
	}

	if existingRunCount > 1 {
		s.evClient.Send(ctx, installID, &signals.Signal{
			Type:           signals.OperationUpdateInstallStackOutputs,
			InstallStackID: stackVersion.InstallStackID,
		})
	}

	return nil
}
