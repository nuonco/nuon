package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

// AttachVCSConfigsParams contains parameters for attaching VCS configs to an owner
type AttachVCSConfigsParams struct {
	OwnerID            string
	OwnerType          interface{} // Used to get the table name for polymorphic relationship
	ConnectedGithubVCS *app.ConnectedGithubVCSConfig
	PublicGitVCS       *app.PublicGitVCSConfig
}

// AttachVCSConfigs attaches VCS configurations to an owner entity (polymorphic relationship)
// This handles creating the VCS config records with proper ownership fields set
func (h *Helpers) AttachVCSConfigs(ctx context.Context, params AttachVCSConfigsParams) error {
	ownerTableName := plugins.TableName(h.db, params.OwnerType)

	// Attach connected GitHub VCS config if provided
	if params.ConnectedGithubVCS != nil {
		params.ConnectedGithubVCS.ComponentConfigID = params.OwnerID
		params.ConnectedGithubVCS.ComponentConfigType = ownerTableName

		if err := h.db.WithContext(ctx).Create(params.ConnectedGithubVCS).Error; err != nil {
			return fmt.Errorf("unable to create connected github vcs config: %w", err)
		}
	}

	// Attach public git VCS config if provided
	if params.PublicGitVCS != nil {
		params.PublicGitVCS.ComponentConfigID = params.OwnerID
		params.PublicGitVCS.ComponentConfigType = ownerTableName

		if err := h.db.WithContext(ctx).Create(params.PublicGitVCS).Error; err != nil {
			return fmt.Errorf("unable to create public git vcs config: %w", err)
		}
	}

	return nil
}

// AttachVCSConfigsWithTx is the same as AttachVCSConfigs but accepts a custom transaction
func (h *Helpers) AttachVCSConfigsWithTx(tx *gorm.DB, params AttachVCSConfigsParams) error {
	ownerTableName := plugins.TableName(h.db, params.OwnerType)

	// Attach connected GitHub VCS config if provided
	if params.ConnectedGithubVCS != nil {
		params.ConnectedGithubVCS.ComponentConfigID = params.OwnerID
		params.ConnectedGithubVCS.ComponentConfigType = ownerTableName

		if err := tx.Create(params.ConnectedGithubVCS).Error; err != nil {
			return fmt.Errorf("unable to create connected github vcs config: %w", err)
		}
	}

	// Attach public git VCS config if provided
	if params.PublicGitVCS != nil {
		params.PublicGitVCS.ComponentConfigID = params.OwnerID
		params.PublicGitVCS.ComponentConfigType = ownerTableName

		if err := tx.Create(params.PublicGitVCS).Error; err != nil {
			return fmt.Errorf("unable to create public git vcs config: %w", err)
		}
	}

	return nil
}
