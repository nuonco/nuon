package worker

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Activities struct {
	l *zap.Logger
}

func NewActivities(l *zap.Logger, db *gorm.DB) *Activities {
	return &Activities{
		l: l,
	}
}

type PrintOrgRequest struct {
	OrgID string
}

func (a *Activities) PrintOrg(ctx context.Context, req PrintOrgRequest) error {
	a.l.Info("printing org", zap.String("id", req.OrgID))
	return nil
}
