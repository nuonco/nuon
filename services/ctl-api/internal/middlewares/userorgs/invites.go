package userorgs

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (m *middleware) handleInvites(ctx context.Context, subject, email string) error {
	var invites []app.OrgInvite
	res := m.db.
		WithContext(ctx).
		Where(&app.OrgInvite{
			Email: email,
		}).
		Find(&invites)
	if res.Error != nil {
		return res.Error
	}
	if len(invites) < 1 {
		return nil
	}

	for _, invite := range invites {
		if err := m.acceptInvite(ctx, &invite, subject); err != nil {
			return nil
		}
	}

	return nil
}

func (m *middleware) acceptInvite(ctx context.Context, invite *app.OrgInvite, subject string) error {
	// create org user
	userOrg := &app.UserOrg{
		OrgID:  invite.OrgID,
		UserID: subject,
	}
	err := m.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&userOrg).Error
	if err != nil {
		return fmt.Errorf("unable to add user to org: %w", err)
	}

	// update invite object
	res := m.db.WithContext(ctx).
		Model(&app.OrgInvite{ID: invite.ID}).
		Updates(app.OrgInvite{Status: app.OrgInviteStatusAccepted})
	if res.Error != nil {
		return fmt.Errorf("unable to update invite: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("invite not found %w", gorm.ErrRecordNotFound)
	}

	// send a notification to the correct org event flow that it was accepted
	m.orgsHooks.InviteAccepted(ctx, invite.OrgID, invite.ID)

	return nil
}
