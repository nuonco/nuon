package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	pkggenerics "github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
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
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(err)
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

	s.evClient.Send(ctx, installID, &signals.Signal{
		Type:           signals.OperationUpdateInstallStackOutputs,
		InstallStackID: stackVersion.InstallStackID,
	})

	return nil
}
