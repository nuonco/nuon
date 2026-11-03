package app

import (
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

type OrgMemberStatus string

const (
	OrgMemberStatusActive  OrgMemberStatus = "active"
	OrgMemberStatusInvited OrgMemberStatus = "invited"
)

type OrgMember struct {
	ID           string             `json:"id"`
	AccountID    string             `json:"account_id,omitzero"`
	InviteID     string             `json:"invite_id,omitzero"`
	Email        string             `json:"email"`
	Name         string             `json:"name,omitzero"`
	Status       OrgMemberStatus    `json:"status"`
	RoleType     RoleType           `json:"role_type,omitzero"`
	RoleBindings []RoleBinding      `json:"role_bindings" gorm:"-"`
	InviteScope  *permissions.Scope `json:"invite_scope,omitempty" gorm:"serializer:json"`
	JoinedAt     *time.Time         `json:"joined_at,omitzero"`
	CreatedAt    time.Time          `json:"created_at"`
}
