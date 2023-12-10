package migrations

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm/clause"
)

func (a *Migrations) createSandbox(ctx context.Context, sandboxName, version string) error {
	// create the default sandbox
	sandbox := app.Sandbox{
		Name:        sandboxName,
		Description: "default aws sandbox",
	}
	res := a.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoNothing: true,
		}).
		Create(&sandbox)
	if res.Error != nil {
		return fmt.Errorf("unable to create sandbox: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return nil
	}

	// create sandbox version
	baseURL := a.cfg.SandboxArtifactsBaseURL
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	baseURL += filepath.Join(sandboxName, version) + "/"
	sandboxRelease := app.SandboxRelease{
		Version:                 version,
		ProvisionPolicyURL:      baseURL + "provision.json",
		TrustPolicyURL:          baseURL + "trust.json",
		DeprovisionPolicyURL:    baseURL + "deprovision.json",
		OneClickRoleTemplateURL: baseURL + "install-role.yaml",
	}
	err := a.db.Model(&sandbox).Where("id = ?", sandbox.ID).Association("Releases").Append(&sandboxRelease)
	if err != nil {
		return fmt.Errorf("unable to save sandbox release: %w", err)
	}

	return nil
}

// this function is used to seed the minimal amount of dependencies we need to properly bootstrap the application. It
// should not be used for general dev seeding, `nuonctl api seed` is where we manage that.
func (a *Migrations) migration003Seed(ctx context.Context) error {
	a.l.Info("creating default aws sandbox")
	if err := a.createSandbox(ctx, "aws-eks", "08e7f11"); err != nil {
		return fmt.Errorf("unable to create aws-eks sandbox: %w", err)
	}
	if err := a.createSandbox(ctx, "empty", "08e7f11"); err != nil {
		return fmt.Errorf("unable to create empty sandbox: %w", err)
	}

	return nil
}
