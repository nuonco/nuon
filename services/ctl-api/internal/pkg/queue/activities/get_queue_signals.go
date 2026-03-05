package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @wrapper-prefix QueueInternal
// @by-field QueueID
func (a *Activities) getQueueSignals(ctx context.Context, queueID string) ([]*app.QueueSignal, error) {
	var queueSignals []*app.QueueSignal

	jdb := generics.NewJSONBQuery(a.db.WithContext(ctx))
	if res := jdb.WhereJSON(generics.JSONBQuery{
		Operator: "=",
		Field:    "status",
		Path:     "status",
		Value:    app.StatusQueued,
	}).Where(app.QueueSignal{
		QueueID: queueID,
	}).Order("created_at asc").
		Find(&queueSignals); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get queue signals")
	}

	return queueSignals, nil
}
