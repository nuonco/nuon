package app

import (
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

type SlackThreadAnchor struct {
	ID        string    `gorm:"primarykey" json:"id,omitzero"`
	CreatedAt time.Time `json:"created_at,omitzero"`
	UpdatedAt time.Time `json:"updated_at,omitzero"`

	TeamID     string `gorm:"notnull;uniqueIndex:idx_slack_thread_anchors_tcw" json:"team_id,omitzero"`
	ChannelID  string `gorm:"notnull;uniqueIndex:idx_slack_thread_anchors_tcw" json:"channel_id,omitzero"`
	WorkflowID string `gorm:"notnull;uniqueIndex:idx_slack_thread_anchors_tcw" json:"workflow_id,omitzero"`

	ParentTS     string `gorm:"notnull" json:"parent_ts,omitzero"`
	OrgID        string `gorm:"notnull;index:idx_slack_thread_anchors_org" json:"org_id,omitzero"`
	WorkflowType string `json:"workflow_type,omitzero"`
}

func (SlackThreadAnchor) TableName() string {
	return "slack_thread_anchors"
}

func (a *SlackThreadAnchor) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewSlackThreadAnchorID()
	}
	return nil
}
