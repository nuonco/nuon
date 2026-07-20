package app

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

type AWSAccountConnectionVerificationStatus string

const (
	AWSAccountConnectionVerificationPending  AWSAccountConnectionVerificationStatus = "pending"
	AWSAccountConnectionVerificationVerified AWSAccountConnectionVerificationStatus = "verified"
	AWSAccountConnectionVerificationError    AWSAccountConnectionVerificationStatus = "error"
)

type AWSAccountConnection struct {
	ID                   string                                 `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitempty"`
	CreatedByID          string                                 `gorm:"not null;default:null" json:"created_by_id" temporaljson:"created_by_id,omitempty"`
	CreatedBy            Account                                `json:"-" temporaljson:"created_by,omitempty"`
	CreatedAt            time.Time                              `gorm:"notnull" json:"created_at" temporaljson:"created_at,omitempty"`
	UpdatedAt            time.Time                              `gorm:"notnull" json:"updated_at" temporaljson:"updated_at,omitempty"`
	DeletedAt            soft_delete.DeletedAt                  `gorm:"uniqueIndex:idx_aws_account_connections_org_account_deleted" json:"-" temporaljson:"deleted_at,omitempty"`
	OrgID                string                                 `gorm:"notnull;uniqueIndex:idx_aws_account_connections_org_account_deleted" json:"org_id" temporaljson:"org_id,omitempty"`
	Org                  Org                                    `json:"-" temporaljson:"org,omitempty"`
	Name                 string                                 `gorm:"notnull" json:"name" temporaljson:"name,omitempty"`
	AccountID            string                                 `gorm:"notnull;uniqueIndex:idx_aws_account_connections_org_account_deleted" json:"account_id" temporaljson:"account_id,omitempty"`
	DefaultRegion        string                                 `gorm:"notnull" json:"default_region" temporaljson:"default_region,omitempty"`
	RoleARN              string                                 `gorm:"notnull;default:''" json:"role_arn" temporaljson:"role_arn,omitempty"`
	ExternalID           string                                 `gorm:"notnull;uniqueIndex" json:"external_id" temporaljson:"-"`
	VerificationStatus   AWSAccountConnectionVerificationStatus `gorm:"notnull" json:"verification_status" temporaljson:"verification_status,omitempty"`
	VerificationCode     string                                 `gorm:"notnull;default:''" json:"verification_code" temporaljson:"verification_code,omitempty"`
	VerificationMessage  string                                 `gorm:"notnull;default:''" json:"verification_message" temporaljson:"verification_message,omitempty"`
	LastCheckedAt        *time.Time                             `json:"last_checked_at,omitempty" temporaljson:"last_checked_at,omitempty"`
	VerifiedAt           *time.Time                             `json:"verified_at,omitempty" temporaljson:"verified_at,omitempty"`
	VerifiedPrincipalARN string                                 `gorm:"notnull;default:''" json:"verified_principal_arn" temporaljson:"verified_principal_arn,omitempty"`
}

func (a *AWSAccountConnection) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAWSAccountConnectionID()
	}
	if a.ExternalID == "" {
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			return err
		}
		a.ExternalID = base64.RawURLEncoding.EncodeToString(buf)
	}
	if a.VerificationStatus == "" {
		a.VerificationStatus = AWSAccountConnectionVerificationPending
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}
