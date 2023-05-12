// aws_settings.go
package models

import (
	"strings"
	"time"
)

// ToRegion converts the enum to a compatible region for AWS operations
func (a AWSRegion) ToRegion() string {
	r := strings.ToLower(string(a))
	r = strings.ReplaceAll(r, "_", "-")

	return r
}

type AWSSettings struct {
	Model

	InstallID string

	Region     AWSRegion `faker:"-"`
	IamRoleArn string
	AccountID  string
}

func (AWSSettings) IsInstallSettings() {}

func (AWSSettings) IsNode() {}

func (aws AWSSettings) GetID() string {
	return aws.Model.ID
}

func (aws AWSSettings) GetCreatedAt() time.Time {
	return aws.Model.CreatedAt
}

func (aws AWSSettings) GetUpdatedAt() time.Time {
	return aws.Model.UpdatedAt
}
