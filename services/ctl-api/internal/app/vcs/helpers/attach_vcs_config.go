package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

type AttachVCSConfigsParams struct {
	OwnerID            string
	OwnerType          interface{}
	ConnectedGithubVCS *app.ConnectedGithubVCSConfig
	PublicGitVCS       *app.PublicGitVCSConfig
}

func (h *Helpers) AttachVCSConfigs(ctx context.Context, params AttachVCSConfigsParams) error {
	ownerTableName := plugins.TableName(h.db, params.OwnerType)

	if params.ConnectedGithubVCS != nil {
		params.ConnectedGithubVCS.ComponentConfigID = params.OwnerID
		params.ConnectedGithubVCS.ComponentConfigType = ownerTableName

		if err := h.db.WithContext(ctx).Create(params.ConnectedGithubVCS).Error; err != nil {
			return fmt.Errorf("unable to create connected github vcs config: %w", err)
		}
	}

	if params.PublicGitVCS != nil {
		params.PublicGitVCS.ComponentConfigID = params.OwnerID
		params.PublicGitVCS.ComponentConfigType = ownerTableName

		if err := h.db.WithContext(ctx).Create(params.PublicGitVCS).Error; err != nil {
			return fmt.Errorf("unable to create public git vcs config: %w", err)
		}
	}

	return nil
}

func (h *Helpers) AttachVCSConfigsWithTx(tx *gorm.DB, params AttachVCSConfigsParams) error {
	ownerTableName := plugins.TableName(h.db, params.OwnerType)

	if params.ConnectedGithubVCS != nil {
		params.ConnectedGithubVCS.ComponentConfigID = params.OwnerID
		params.ConnectedGithubVCS.ComponentConfigType = ownerTableName

		if err := tx.Create(params.ConnectedGithubVCS).Error; err != nil {
			return fmt.Errorf("unable to create connected github vcs config: %w", err)
		}
	}

	if params.PublicGitVCS != nil {
		params.PublicGitVCS.ComponentConfigID = params.OwnerID
		params.PublicGitVCS.ComponentConfigType = ownerTableName

		if err := tx.Create(params.PublicGitVCS).Error; err != nil {
			return fmt.Errorf("unable to create public git vcs config: %w", err)
		}
	}

	return nil
}
