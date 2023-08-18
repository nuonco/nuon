package app

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/deprecated/api/gqlclient"
	"github.com/powertoolsdev/mono/pkg/ui"
)

type GetInstallOpts struct {
	InstallID string
}

func (c *commands) Echo(ctx context.Context) error {
	ui.Line(ctx, "hello world")

	return nil
}

func (c *commands) GetInstall(ctx context.Context, opts *GetInstallOpts) error {
	// TODO: maybe fetch by install name
	if opts.InstallID != "" {
		install, err := c.apiClient.GetInstall(ctx, opts.InstallID)
		if err != nil {
			return fmt.Errorf("unable to get install: %w", err)
		}

		ui.Line(ctx, "%s - %s (%s)", install.Id, install.Name, install.GetSettings())
		return nil
	}

	return fmt.Errorf("no install ID was provided")
}

type CreateInstallOpts struct {
	InstallName   string `validate:"required"`
	InstallRegion string `validate:"required,oneof=US_EAST_1 US_EAST_2 US_WEST_2"`
	InstallARN    string `validate:"required"`
}

func (c *commands) CreateInstall(ctx context.Context, opts *CreateInstallOpts) error {
	validate := validator.New()
	if err := validate.Struct(opts); err != nil {
		return fmt.Errorf("somethings wrong: %w", err)
	}
	awsSettings := gqlclient.AWSSettingsInput{
		Region: "US_EAST_1",
		Role:   opts.InstallARN,
	}
	input := gqlclient.InstallInput{
		Name:        opts.InstallName,
		AppId:       c.appID,
		AwsSettings: &awsSettings,
	}

	ui.Line(ctx, "attempting to create an install: %s", input.Name)

	install, err := c.apiClient.UpsertInstall(ctx, input)
	if err != nil {
		return fmt.Errorf("unable to create this install: %w", err)
	}

	ui.Line(ctx, "successfully created install:  %s - %s (%s)", install.Id, install.Name, install.GetSettings())
	return nil
}

func (c *commands) GetInstallStatus(ctx context.Context, installID string) error {
	if installID == "" {
		return fmt.Errorf("provide an install ID")
	}

	status, err := c.apiClient.GetInstallStatus(ctx, c.orgID, c.appID, installID)
	if err != nil {
		return fmt.Errorf("can't get install status: %w", err)
	}

	ui.Line(ctx, "install %s status: %s", installID, status)
	return nil
}
