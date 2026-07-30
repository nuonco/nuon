package activities

import (
	"context"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type ListPhoneHomeBackfillInstallsRequest struct {
	// CursorCreatedAt / CursorID form the keyset cursor: the batch selects installs
	// ordered by (created_at, id) ascending that sort after this pair. Zero values
	// start from the beginning.
	CursorCreatedAt time.Time `json:"cursor_created_at"`
	CursorID        string    `json:"cursor_id"`
	Limit           int       `json:"limit"`
}

type ListPhoneHomeBackfillInstallsResponse struct {
	InstallIDs []string `json:"install_ids"`
	// LastCreatedAt / LastID are the cursor of the last install examined; feed them
	// back as the next CursorCreatedAt / CursorID. LastID is empty on an empty batch.
	LastCreatedAt time.Time `json:"last_created_at"`
	LastID        string    `json:"last_id"`
	// Examined is how many installs this batch looked at; a value < Limit means the
	// fleet is drained.
	Examined int `json:"examined"`
}

// backfillInstallRow is a bare projection so scanning does not run Install's
// AfterQuery hook, which would fire a per-row org lookup for data the backfill does
// not need.
type backfillInstallRow struct {
	ID        string
	CreatedAt time.Time
}

// ListPhoneHomeBackfillInstalls returns one keyset-paginated page of installs that
// are candidates for phone-home secret provisioning.
//
// Pre-filtered to orgs with the feature enabled, which is the dominant filter and is
// cheap because the flag is a deliberate per-org admin action. Every other condition
// — AWS-only, a target account, a reachable management account — is left to
// EnsureInstallPhoneHomeSecret's own gate, so that gate stays the single source of
// truth for eligibility. This pre-filter can only reduce work, never change an
// outcome.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) ListPhoneHomeBackfillInstalls(
	ctx context.Context, req ListPhoneHomeBackfillInstallsRequest,
) (*ListPhoneHomeBackfillInstallsResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	var orgIDs []string
	if res := a.db.WithContext(ctx).
		Model(&app.Org{}).
		Where("features->>? = ?", string(app.OrgFeaturePhoneHomeAuth), "true").
		Pluck("id", &orgIDs); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error)
	}

	resp := &ListPhoneHomeBackfillInstallsResponse{InstallIDs: []string{}}
	if len(orgIDs) == 0 {
		return resp, nil
	}

	var rows []backfillInstallRow
	if res := a.db.WithContext(ctx).
		Model(&app.Install{}).
		Select("id", "created_at").
		Where("org_id IN ?", orgIDs).
		Where("(created_at, id) > (?, ?)", req.CursorCreatedAt, req.CursorID).
		Order("created_at asc, id asc").
		Limit(limit).
		Find(&rows); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error)
	}

	for _, row := range rows {
		resp.InstallIDs = append(resp.InstallIDs, row.ID)
	}
	resp.Examined = len(rows)

	if len(rows) > 0 {
		last := rows[len(rows)-1]
		resp.LastCreatedAt = last.CreatedAt
		resp.LastID = last.ID
	}

	return resp, nil
}
