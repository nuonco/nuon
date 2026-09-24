package app

import "time"

type OrgMemberStatus string

const (
	OrgMemberStatusActive  OrgMemberStatus = "active"
	OrgMemberStatusInvited OrgMemberStatus = "invited"
)

type OrgMember struct {
	ID        string          `json:"id"`
	AccountID string          `json:"account_id,omitzero"`
	InviteID  string          `json:"invite_id,omitzero"`
	Email     string          `json:"email"`
	Name      string          `json:"name,omitzero"`
	Status    OrgMemberStatus `json:"status"`
	RoleType  RoleType        `json:"role_type,omitzero"`
	JoinedAt  *time.Time      `json:"joined_at,omitzero"`
	CreatedAt time.Time       `json:"created_at"`
}
