// gcp_settings.go
package models

type GCPSettings struct {
	ModelV2

	InstallID string
	Install   Install
}

func (GCPSettings) IsInstallSettings() {}

func (GCPSettings) IsNode() {}
