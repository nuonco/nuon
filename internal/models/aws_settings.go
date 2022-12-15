// aws_settings.go
package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// ToRegion converts the enum to a compatible region for AWS operations
func (a AWSRegion) ToRegion() string {
	r := strings.ToLower(string(a))
	r = strings.ReplaceAll(r, "_", "-")

	return r
}

type AWSSettings struct {
	Model

	InstallID uuid.UUID

	Region           AWSRegion `fake:"skip"`
	IamRoleArn       string
	AccountID        string
	NotificationsURL string
}

func (AWSSettings) IsInstallSettings() {}

func (AWSSettings) IsNode() {}

func (aws AWSSettings) GetID() string {
	return aws.Model.ID.String()
}

func (aws AWSSettings) GetCreatedAt() time.Time {
	return aws.Model.CreatedAt
}

func (aws AWSSettings) GetUpdatedAt() time.Time {
	return aws.Model.UpdatedAt
}
