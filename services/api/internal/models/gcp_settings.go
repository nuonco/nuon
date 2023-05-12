// gcp_settings.go
package models

type GCPSettings struct {
	Model

	InstallID string
	Install   Install
}

func (GCPSettings) IsInstallSettings() {}

func (GCPSettings) IsNode() {}
