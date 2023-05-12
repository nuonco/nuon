package models

import "time"

type SandboxVersion struct {
	Model
	SandboxName    string
	SandboxVersion string
	TfVersion      string
}

func (s SandboxVersion) GetID() string {
	return s.Model.ID
}

func (s SandboxVersion) GetCreatedAt() time.Time {
	return s.Model.CreatedAt
}

func (s SandboxVersion) GetUpdatedAt() time.Time {
	return s.Model.UpdatedAt
}
