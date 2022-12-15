// gcp_settings.go
package models

import "github.com/google/uuid"

type GCPSettings struct {
	Model

	InstallID uuid.UUID
	Install   Install
}

func (GCPSettings) IsInstallSettings() {}

func (GCPSettings) IsNode() {}
