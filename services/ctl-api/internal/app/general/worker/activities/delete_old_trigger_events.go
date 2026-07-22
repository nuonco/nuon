package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	rejectedTriggerEventRetention = 7 * 24 * time.Hour
	ignoredTriggerEventRetention  = 30 * 24 * time.Hour
)

type DeleteOldTriggerEventsRequest struct {
	BatchSize     int                    `json:"batch_size"`
	RoutingStatus app.EventRoutingStatus `json:"routing_status"`
}

type DeleteOldTriggerEventsResponse struct {
	Deleted int64 `json:"deleted"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) DeleteOldTriggerEvents(ctx context.Context, req DeleteOldTriggerEventsRequest) (*DeleteOldTriggerEventsResponse, error) {
	if req.BatchSize <= 0 {
		req.BatchSize = 5000
	}

	retention, err := triggerEventRetention(req.RoutingStatus)
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().Add(-retention)
	var deleted int64
	err = a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		referencedEvents := tx.Model(&app.EventRunbookWaiter{}).
			Select("matched_event_id").
			Where(clause.Neq{Column: "matched_event_id", Value: nil})
		var candidates []app.TriggerEvent
		if err := tx.Unscoped().
			Select("id").
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where(app.TriggerEvent{RoutingStatus: req.RoutingStatus}).
			Where("received_at < ?", cutoff).
			Where("id NOT IN (?)", referencedEvents).
			Order("routing_status, received_at, id").
			Limit(req.BatchSize).
			Find(&candidates).Error; err != nil {
			return err
		}
		if len(candidates) == 0 {
			return nil
		}
		ids := make([]string, len(candidates))
		for i := range candidates {
			ids[i] = candidates[i].ID
		}
		if err := tx.Unscoped().
			Where(map[string]any{"trigger_event_id": ids}).
			Delete(&app.EventDispatch{}).Error; err != nil {
			return err
		}
		res := tx.Unscoped().
			Where(app.TriggerEvent{RoutingStatus: req.RoutingStatus}).
			Where("received_at < ?", cutoff).
			Where("id IN ?", ids).
			Where("id NOT IN (?)", referencedEvents).
			Delete(&app.TriggerEvent{})
		deleted = res.RowsAffected
		return res.Error
	})
	if err != nil {
		return nil, fmt.Errorf("unable to delete old trigger events: %w", err)
	}

	a.mw.Count("general.trigger_event_cleanup.deleted", deleted, []string{"routing_status:" + string(req.RoutingStatus)})
	return &DeleteOldTriggerEventsResponse{Deleted: deleted}, nil
}

func triggerEventRetention(status app.EventRoutingStatus) (time.Duration, error) {
	switch status {
	case app.EventRoutingStatusRejected:
		return rejectedTriggerEventRetention, nil
	case app.EventRoutingStatusIgnored:
		return ignoredTriggerEventRetention, nil
	default:
		return 0, fmt.Errorf("unsupported trigger event cleanup status %q", status)
	}
}
