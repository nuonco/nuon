package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

type CreateComponentReleaseRequest struct {
	BuildID string `json:"build_id"`

	Strategy struct {
		InstallsPerStep int    `json:"installs_per_step"`
		Delay           string `json:"delay"`
	} `json:"strategy"`
}

func (c *CreateComponentReleaseRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID CreateComponentRelease
// @Summary	create a release
// @Description.markdown	create_component_release.md
// @Param			component_id	path	string	true	"component ID"
// @Tags			releases
// @Accept			json
// @Param			req	body	CreateComponentReleaseRequest	true	"Input"
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		201				{object}	app.ComponentRelease
// @Router			/v1/components/{component_id}/releases [post]
func (s *service) CreateComponentRelease(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	cmpID := ctx.Param("component_id")

	var req CreateComponentReleaseRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	app, err := s.createRelease(ctx, cmpID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create release: %w", err))
		return
	}

	s.hooks.Created(ctx, app.ID, org.SandboxMode)
	ctx.JSON(http.StatusCreated, app)
}

func (s *service) createReleaseSteps(installs []app.Install, req *CreateComponentReleaseRequest) ([]*app.ComponentReleaseStep, error) {
	installIDs := installsToIDSlice(installs)

	installsPerStep := req.Strategy.InstallsPerStep
	if installsPerStep == 0 {
		installsPerStep = len(installs)
	}
	stepInstalls := generics.SliceToGroups(installIDs, installsPerStep)

	steps := make([]*app.ComponentReleaseStep, 0)
	for _, grp := range stepInstalls {
		step := &app.ComponentReleaseStep{
			Status:              "queued",
			StatusDescription:   "queued",
			RequestedInstallIDs: grp,
		}

		delay, err := time.ParseDuration(req.Strategy.Delay)
		if err != nil {
			return nil, fmt.Errorf("invalid delay for component release: %w", err)
		}
		step.Delay = generics.ToPtr(delay.String())
		steps = append(steps, step)
	}

	return steps, nil
}

func (s *service) createRelease(ctx context.Context, cmpID string, req *CreateComponentReleaseRequest) (*app.ComponentRelease, error) {
	// fetch the component, app, installs and the build
	cmp := app.Component{}
	res := s.db.WithContext(ctx).
		Preload("App").
		Preload("App.Installs", "status IN ?", []string{"active", "queued", "provisioning"}).
		First(&cmp, "id = ?", cmpID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	steps, err := s.createReleaseSteps(cmp.App.Installs, req)
	if err != nil {
		return nil, fmt.Errorf("unable to create release steps: %w", err)
	}

	// create the component release
	release := app.ComponentRelease{
		Status:            "queued",
		StatusDescription: "queued and waiting for event loop to process",
	}
	bld := app.ComponentBuild{}
	err = s.db.WithContext(ctx).
		First(&bld, "id = ?", req.BuildID).
		Association("ComponentReleases").
		Append(&release)
	if err != nil {
		return nil, fmt.Errorf("unable to create build for component: %w", err)
	}

	// append the steps into the association
	var rel app.ComponentRelease
	err = s.db.WithContext(ctx).
		First(&rel, "id = ?", release.ID).
		Association("ComponentReleaseSteps").
		Append(generics.ToIntSlice(steps)...)
	if err != nil {
		return nil, fmt.Errorf("unable to create release steps: %w", err)
	}

	// create release and steps, according to the inputs
	return &release, nil
}
