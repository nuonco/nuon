package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateAppKubernetesContextsConfigRequest struct {
	AppConfigID string `json:"app_config_id" validate:"required"`

	Contexts []AppKubernetesContext `json:"contexts" validate:"dive"`
}

type AppKubernetesContext struct {
	Name string `json:"name" validate:"required"`
	// Component is the name of the peer terraform_module or pulumi component
	// that emits cluster connection details as outputs.
	Component string `json:"component" validate:"required"`
}

func (c *CreateAppKubernetesContextsConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	seen := make(map[string]struct{}, len(c.Contexts))
	for _, ctx := range c.Contexts {
		if _, dup := seen[ctx.Name]; dup {
			return fmt.Errorf("duplicate kubernetes_context name: %s", ctx.Name)
		}
		seen[ctx.Name] = struct{}{}
	}

	return nil
}

// @ID						CreateAppKubernetesContextsConfig
// @Summary				create a kubernetes contexts config
// @Description			Create the named kubernetes_context bindings for an app config version. Each context names a peer terraform_module or pulumi component that emits cluster connection details as outputs.
// @Tags					apps
// @Accept					json
// @Param					req		body	CreateAppKubernetesContextsConfigRequest	true	"Input"
// @Produce				json
// @Param					app_id	path	string	true	"app ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppKubernetesContextsConfig
// @Router					/v1/apps/{app_id}/kubernetes-contexts-configs [post]
func (s *service) CreateAppKubernetesContextsConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req CreateAppKubernetesContextsConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrInvalidRequest{Err: err})
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(err)
		return
	}

	cfg, err := s.createAppKubernetesContextsConfig(ctx, appID, &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) createAppKubernetesContextsConfig(ctx context.Context, appID string, req *CreateAppKubernetesContextsConfigRequest) (*app.AppKubernetesContextsConfig, error) {
	// Resolve component names -> IDs in a single query so each child row has
	// a stable FK back to the source component. The name is also persisted
	// (SourceComponentName) so the binding remains intelligible across
	// AppConfig versions if the underlying component is renamed.
	componentIDByName := map[string]string{}
	if len(req.Contexts) > 0 {
		names := make([]string, 0, len(req.Contexts))
		for _, c := range req.Contexts {
			names = append(names, c.Component)
		}

		var components []app.Component
		if res := s.db.WithContext(ctx).
			Select("id", "name").
			Where(&app.Component{AppID: appID}).
			Where("name IN ?", names).
			Find(&components); res.Error != nil {
			return nil, errors.Wrap(res.Error, "unable to look up source components")
		}
		for _, c := range components {
			componentIDByName[c.Name] = c.ID
		}
	}

	inputs := make([]build.KubernetesContextInput, 0, len(req.Contexts))
	for _, c := range req.Contexts {
		inputs = append(inputs, build.KubernetesContextInput{Name: c.Name, ComponentName: c.Component})
	}

	obj, err := build.KubernetesContextsConfig(inputs, componentIDByName, appID, req.AppConfigID)
	if err != nil {
		return nil, stderr.ErrInvalidRequest{Err: err}
	}

	if res := s.db.WithContext(ctx).Create(obj); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create app kubernetes contexts config")
	}

	return obj, nil
}
