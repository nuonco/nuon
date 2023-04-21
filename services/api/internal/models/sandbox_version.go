package models

import "time"

type SandboxVersion struct {
	ModelV2
	SandboxName    string
	SandboxVersion string
	TfVersion      string
}

func (s SandboxVersion) GetID() string {
	return s.ModelV2.ID
}

func (s SandboxVersion) GetCreatedAt() time.Time {
	return s.ModelV2.CreatedAt
}

func (s SandboxVersion) GetUpdatedAt() time.Time {
	return s.ModelV2.UpdatedAt
}
