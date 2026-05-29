package service

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	ddclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/datadog/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/hooks/datadogrender"
)

// TestConnectionResponse reports the outcome of a Test action: did we
// reach DD at all (ValidatedAPIKey), and did the synthetic event land
// (PostedEventID, EventURL). Two distinct booleans rather than a single
// "ok" because they fail under different conditions and the dashboard
// renders different errors for each.
type TestConnectionResponse struct {
	ValidatedAPIKey bool   `json:"validated_api_key"`
	PostedEventID   int64  `json:"posted_event_id,omitempty"`
	EventURL        string `json:"event_url,omitempty"`
}

// @ID						TestDatadogConnection
// @Summary				Test a Datadog connection end-to-end
// @Description			Probes /api/v1/validate against the stored key, then posts a synthetic event into the tenant's event stream so the user can confirm the integration end-to-end in the DD UI. Both calls run live against DD with a 10s timeout. Either step failing returns 400 with the upstream error surfaced inline so the dashboard can render it as a toast. The connection must belong to the calling org.
// @Tags					datadog
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id			path	string	true	"Org ID"
// @Param					connection_id	path	string	true	"Connection ID"
// @Success				200		{object}	TestConnectionResponse
// @Failure				400		{object}	stderr.ErrResponse
// @Failure				404		{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/datadog/connections/{connection_id}/test [POST]
func (s *service) TestConnection(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	connectionID := ctx.Param("connection_id")

	var conn app.DatadogConnection
	if err := s.db.WithContext(ctx).
		Where(app.DatadogConnection{ID: connectionID, OrgID: org.ID}).
		First(&conn).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Error(stderr.ErrNotFound{
				Err:         fmt.Errorf("datadog connection %q not found in org %q", connectionID, org.ID),
				Description: "Datadog connection not found",
			})
			return
		}
		ctx.Error(fmt.Errorf("unable to fetch datadog connection: %w", err))
		return
	}

	// 1. Validate the stored API key. validateAPIKey already maps DD
	// 401/403 to stderr.ErrInvalidRequest so the dashboard renders 400.
	if err := s.validateAPIKey(ctx, conn.Site, conn.APIKey); err != nil {
		ctx.Error(err)
		return
	}

	// 2. Post a synthetic event so the user can verify in DD's UI.
	// Reuses datadogrender.SourceName so the event has the same
	// `source:nuon` tag every real emit carries — if a monitor in DD
	// is already filtering on `source:nuon`, the test event is part
	// of that surface too (intentional — confirms end-to-end including
	// downstream monitor wiring).
	baseURL := ddclient.ResolveSiteURL(conn.Site)
	req := ddclient.PostEventRequest{
		Title:          "Nuon: integration test from " + acct.ID,
		Text:           buildTestEventBody(&conn, acct),
		Tags:           buildTestEventTags(&conn),
		AlertType:      ddclient.EventAlertTypeInfo,
		Priority:       ddclient.EventPriorityLow,
		AggregationKey: "nuon-integration-test-" + conn.ID,
		SourceTypeName: datadogrender.SourceName,
		DateHappened:   time.Now().Unix(),
	}
	posted, err := s.ddClient.PostEvent(ctx, baseURL, conn.APIKey, req)
	if err != nil {
		// DD rejected the post even though it accepted the key —
		// usually means the key has scope restrictions or DD is
		// degraded for the tenant. Surface as ErrInvalidRequest so
		// the dashboard renders 400 with the DD error body inline.
		ctx.Error(stderr.NewInvalidRequest(
			fmt.Errorf("datadog rejected the test event: %w", err)))
		return
	}

	ctx.JSON(http.StatusOK, TestConnectionResponse{
		ValidatedAPIKey: true,
		PostedEventID:   posted.Event.ID,
		EventURL:        posted.Event.URL,
	})
}

func buildTestEventBody(conn *app.DatadogConnection, acct *app.Account) string {
	return "%%%\n" +
		"This is a test event from your Nuon integration.\n\n" +
		"**Connection:** `" + conn.ID + "` (" + conn.Name + ")\n" +
		"**Initiated by:** " + acct.ID + "\n\n" +
		"If you can see this, Nuon can reach your Datadog tenant. " +
		"Real workflow / step / approval / drift events will arrive " +
		"with the same `source:nuon` tag and richer `nuon_*` tags.\n" +
		"%%%"
}

func buildTestEventTags(conn *app.DatadogConnection) []string {
	tags := []string{
		"source:" + datadogrender.SourceName,
		"nuon_kind:integration_test",
		"nuon_org_id:" + conn.OrgID,
		"nuon_connection_id:" + conn.ID,
	}
	tags = append(tags, []string(conn.DefaultTags)...)
	return tags
}
