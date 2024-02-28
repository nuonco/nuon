package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

const (
	appConfigSyncTimeout time.Duration = time.Minute * 2
	appConfigSyncSleep   time.Duration = time.Second * 2
	orgName              string        = "customers-shared-sandbox"
)

const appConfigTemplate string = `
version = "v1"

[installer]
name = "{{.Name}}"
description = "{{.Description}}"
slug = "{{.Slug}}"
documentation_url = "{{.Links.Documentation}}"
community_url = "{{.Links.Community}}"
homepage_url = "{{.Links.Homepage}}"
github_url = "{{.Links.Github}}"
logo_url = "{{.Links.Logo}}"
demo_url = "{{.Links.Demo}}"

[runner]
runner_type = "aws-ecs"

[sandbox]
terraform_version = "1.5.4"

[sandbox.public_repo]
directory = "aws-ecs-byo-vpc"
repo = "nuonco/sandboxes"
branch = "main"
`

type AdminCreateDemoInstallerRequest struct {
	Slug        string `validate:"required" json:"slug"`
	Name        string `validate:"required" json:"name"`
	Description string `validate:"required" json:"description"`

	Links struct {
		Documentation string `validate:"required" json:"documentation"`
		Logo          string `validate:"required" json:"logo"`
		Github        string `validate:"required" json:"github"`
		Homepage      string `validate:"required" json:"homepage"`
		Community     string `validate:"required" json:"community"`
		Demo          string `json:"demo"`
	} `json:"links"`
}

func (c *AdminCreateDemoInstallerRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

const ()

// @ID AdminCreateDemoInstaller
// @Description.markdown admin_create_demo_installer.md
// @Tags			installers/admin
// @Accept			json
// @Param			req	body	AdminCreateDemoInstallerRequest	true	"Input"
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/installers/admin-demo-installer [POST]
func (s *service) AdminCreateDemoInstaller(ctx *gin.Context) {
	org, err := s.getOrg(ctx, orgName)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req AdminCreateDemoInstallerRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	app, err := s.ensureApp(ctx, org, req.Name)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create or upsert app: %w", err))
		return
	}

	appCfg, err := s.renderTemplate(ctx, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to render template config: %w", err))
		return
	}

	cfgObj, err := s.createAppConfig(ctx, app, appCfg)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app config: %w", err))
		return
	}

	if err := s.pollAppConfig(ctx, cfgObj.ID); err != nil {
		ctx.Error(fmt.Errorf("error polling app config to be generated: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) ensureApp(ctx context.Context, org *app.Org, name string) (*app.App, error) {
	// fetch the app, because on an upsert, the created-by-id will be incorrect.
	var resolvedApp app.App
	res := s.db.WithContext(ctx).Where(app.App{
		OrgID: org.ID,
		Name:  name,
	}).First(&resolvedApp)
	if res.Error == nil {
		return &resolvedApp, nil
	}
	if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("unable to fetch app: %w", res.Error)
	}

	ap := app.App{
		Name:        name,
		OrgID:       org.ID,
		CreatedByID: org.CreatedByID,
	}
	res = s.db.WithContext(ctx).Create(&ap)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create app: %w", res.Error)
	}

	s.appHooks.Created(ctx, ap.ID, org.OrgType)
	return &ap, nil
}

func (s *service) renderTemplate(ctx context.Context, req *AdminCreateDemoInstallerRequest) ([]byte, error) {
	temp, err := template.New("config").Parse(appConfigTemplate)
	if err != nil {
		return nil, fmt.Errorf("unable to render template: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := temp.Execute(buf, req); err != nil {
		return nil, fmt.Errorf("unable to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *service) createAppConfig(ctx context.Context, ap *app.App, cfg []byte) (*app.AppConfig, error) {
	appCfg := app.AppConfig{
		CreatedByID:       ap.CreatedByID,
		OrgID:             ap.OrgID,
		AppID:             ap.ID,
		Format:            app.AppConfigFmtToml,
		Content:           string(cfg),
		Status:            app.AppConfigStatusPending,
		StatusDescription: "waiting to be synced",
	}

	res := s.db.WithContext(ctx).Create(&appCfg)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create app config: %w", res.Error)
	}

	s.appHooks.ConfigCreated(ctx, ap.ID, appCfg.ID)

	return &appCfg, nil
}

func (s *service) pollAppConfig(ctx context.Context, appConfigID string) error {
	ctx, cancel := context.WithTimeout(ctx, appConfigSyncTimeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		cfg, err := s.getAppConfig(ctx, appConfigID)
		if err != nil {
			return err
		}

		if cfg.Status == app.AppConfigStatusActive {
			return nil
		}
		if cfg.Status == app.AppConfigStatusError {
			return fmt.Errorf("unable to sync app config")
		}

		time.Sleep(appConfigSyncSleep)
	}

	return nil
}

func (s *service) getAppConfig(ctx context.Context, appCfgID string) (*app.AppConfig, error) {
	appCfg := app.AppConfig{}
	res := s.db.WithContext(ctx).
		Preload("CreatedBy").
		First(&appCfg, "id = ?", appCfgID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app config: %w", res.Error)
	}

	return &appCfg, nil
}

func (s *service) getOrg(ctx context.Context, nameOrID string) (*app.Org, error) {
	org := app.Org{}
	res := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Where("name LIKE ?", nameOrID).
		Or("id = ?", nameOrID).
		First(&org)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org: %w", res.Error)
	}

	return &org, nil
}
