package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateAppSecretsConfigRequest struct {
	AppConfigID string `json:"app_config_id" validate:"required"`

	Secrets []AppSecretConfig `json:"secrets" validate:"dive"`
}

func (c *CreateAppSecretsConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

func (c CreateAppSecretsConfigRequest) buildInputs() []build.SecretInput {
	objs := make([]build.SecretInput, 0, len(c.Secrets))
	for _, secr := range c.Secrets {
		targets := make([]config.KubernetesSyncTarget, 0, len(secr.KubernetesSyncTargets))
		for _, t := range secr.KubernetesSyncTargets {
			targets = append(targets, config.KubernetesSyncTarget{
				Namespaces: t.Namespaces,
				Name:       t.Name,
				Key:        t.Key,
			})
		}

		objs = append(objs, build.SecretInput{
			Name:                      secr.Name,
			DisplayName:               secr.DisplayName,
			Description:               secr.Description,
			Required:                  secr.Required,
			AutoGenerate:              secr.AutoGenerate,
			Format:                    secr.Format,
			Default:                   secr.Default,
			KubernetesSync:            secr.KubernetesSync || len(targets) > 0,
			KubernetesSecretNamespace: secr.KubernetesSecretNamespace,
			KubernetesSecretName:      secr.KubernetesSecretName,
			KubernetesSyncTargets:     targets,
		})
	}
	return objs
}

type AppSecretConfig struct {
	Name        string `json:"name" validate:"interpolated_name,required"`
	DisplayName string `json:"display_name" validate:"required"`
	Description string `json:"description" validate:"required"`

	Required     bool   `json:"required"`
	AutoGenerate bool   `json:"auto_generate"`
	Format       string `json:"format"`
	Default      string `json:"default"`

	KubernetesSync            bool   `json:"kubernetes_sync"`
	KubernetesSecretNamespace string `json:"kubernetes_secret_namespace"`
	KubernetesSecretName      string `json:"kubernetes_secret_name" validate:"omitempty,hostname_rfc1123"`

	KubernetesSyncTargets []KubernetesSyncTarget `json:"kubernetes_sync_targets" validate:"dive"`
}

type KubernetesSyncTarget struct {
	Namespaces []string `json:"namespaces" validate:"required,min=1,dive,hostname_rfc1123"`
	Name       string   `json:"name" validate:"required,hostname_rfc1123"`
	Key        string   `json:"key" validate:"required"`
}

func (c AppSecretConfig) getKubernetesSyncTargets(appID string) []app.AppSecretKubernetesSyncTarget {
	targets := make([]app.AppSecretKubernetesSyncTarget, 0, len(c.KubernetesSyncTargets))
	for _, t := range c.KubernetesSyncTargets {
		targets = append(targets, app.AppSecretKubernetesSyncTarget{
			Namespaces: t.Namespaces,
			Name:       t.Name,
			Key:        t.Key,
		})
	}
	return targets
}

// @ID						CreateAppSecretsConfig
// @Description.markdown	create_app_secrets_config.md
// @Tags					apps
// @Accept					json
// @Param					req	body	CreateAppSecretsConfigRequest	true	"Input"
// @Produce				json
// @Param					app_id	path	string	true	"app ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppSecretsConfig
// @Router /v1/apps/{app_id}/secrets-configs [post]
func (s *service) CreateAppSecretsConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req CreateAppSecretsConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: err,
		})
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(err)
		return
	}

	cfg, err := s.createAppSecretsConfig(ctx, appID, &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) createAppSecretsConfig(ctx context.Context, appID string, req *CreateAppSecretsConfigRequest) (*app.AppSecretsConfig, error) {
	obj, err := build.SecretsConfig(req.buildInputs(), appID, req.AppConfigID)
	if err != nil {
		return nil, err
	}

	res := s.db.WithContext(ctx).Create(obj)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create app secrets config")
	}

	return obj, nil
}
