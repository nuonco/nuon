package helpers

import (
	"context"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type InstallTelemetryConfig struct {
	Name                string
	Labels              labels.Labels
	AppID               string
	AppName             string  `gorm:"column:App__name"`
	OrgName             string  `gorm:"column:Org__name"`
	TelemetryEnabled    *bool   `gorm:"column:InstallConfig__telemetry_enabled"`
	OrgTelemetryEnabled bool    `gorm:"column:Org__telemetry_enabled"`
	OrgRelayEndpoint    *string `gorm:"column:Org__telemetry_relay_endpoint"`
}

func (h *Helpers) GetInstallTelemetryConfig(ctx context.Context, orgID, installID string) (*InstallTelemetryConfig, error) {
	var config InstallTelemetryConfig
	table := plugins.TableName(h.db, app.Install{})
	err := h.db.WithContext(ctx).
		Model(&app.Install{}).
		Scopes(scopes.WithDisableViews).
		Select(table+".name", table+".labels", table+".app_id").
		Joins("App", h.db.Select("name")).
		Joins("InstallConfig", h.db.Select("telemetry_enabled")).
		Joins("Org", h.db.Select("name", "telemetry_enabled", "telemetry_relay_endpoint")).
		Where(app.Install{ID: installID, OrgID: orgID}).
		Take(&config).Error
	return &config, err
}

func (c *InstallTelemetryConfig) ResourceAttributes() map[string]string {
	attributes := map[string]string{
		"nuon.org.name":     c.OrgName,
		"nuon.app.name":     c.AppName,
		"nuon.install.name": c.Name,
	}
	for key, value := range c.Labels {
		attributes["nuon.install.labels."+key] = value
	}
	return attributes
}
