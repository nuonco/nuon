package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	// maxResourcesPerComponent is a defensive server-side cap; the runner
	// truncates first.
	maxResourcesPerComponent = 500
	// maxResourceDetailsBytes bounds the per-resource details JSON blob.
	maxResourceDetailsBytes = 16 * 1024
	// componentHealthEvaluateSignalType doubles as the dedupe key: one pending
	// evaluation per install queue is all that is ever useful.
	componentHealthEvaluateSignalType = "component-health-evaluate"
	// componentHealthInsertBatchSize bounds each ClickHouse insert.
	componentHealthInsertBatchSize = 500

	sourceComponent = "component"
	sourceSandbox   = "sandbox"

	healthHealthy = "healthy"
	healthUnknown = "unknown"
)

// healthSeverity ranks health so Degraded/unhealthy latch over a later
// Progressing/unknown report; Healthy is 0 and clears the latch (in latchHealth).
var healthSeverity = map[string]int{
	healthHealthy: 0,
	healthUnknown: 1,
	"progressing": 1,
	"degraded":    2,
	"unhealthy":   3,
}

type ComponentHealthResource struct {
	Provider     string `json:"provider"`
	APIGroup     string `json:"api_group"`
	Kind         string `json:"kind"`
	Namespace    string `json:"namespace"`
	Name         string `json:"name"`
	Health       string `json:"health"`
	Message      string `json:"message"`
	NativeStatus string `json:"native_status"`
	Details      string `json:"details"`
}

type ComponentHealthComponent struct {
	InstallComponentID string                    `json:"install_component_id" validate:"required"`
	ComponentID        string                    `json:"component_id"`
	ComponentType      string                    `json:"component_type"`
	NativeStatus       string                    `json:"native_status"`
	Truncated          bool                      `json:"truncated"`
	Resources          []ComponentHealthResource `json:"resources"`
}

// ComponentHealthSandboxRelease is a helm release the install's sandbox manages
// (base infra like external-dns, cert-manager) rather than an app component.
type ComponentHealthSandboxRelease struct {
	ReleaseName string                    `json:"release_name" validate:"required"`
	Namespace   string                    `json:"namespace"`
	Resources   []ComponentHealthResource `json:"resources"`
}

type CreateComponentHealthRequest struct {
	InstallID       string                          `json:"install_id" validate:"required"`
	Kind            string                          `json:"kind"`
	ObservedAt      time.Time                       `json:"observed_at"`
	Components      []ComponentHealthComponent      `json:"components"`
	SandboxReleases []ComponentHealthSandboxRelease `json:"sandbox_releases"`
}

type CreateComponentHealthResponse struct {
	Ingested int `json:"ingested"`
}

// @ID					CreateComponentHealth
// @Summary				report component resource health
// @Description			Batch ingest of the resources a runner's install components manage, with per-resource health. Powers the live resource explorer.
// @Param				req			body	CreateComponentHealthRequest	true	"Input"
// @Param				runner_id	path	string							true	"runner ID"
// @Tags				runners/runner
// @Accept				json
// @Produce				json
// @Security			APIKey
// @Security			OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	service.CreateComponentHealthResponse
// @Router				/v1/runners/{runner_id}/component-health [POST]
func (s *service) CreateComponentHealth(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req CreateComponentHealthRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := s.v.Struct(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to resolve org: %w", err))
		return
	}

	resp, err := s.createComponentHealth(ctx, orgID, runnerID, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to ingest component health: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// resourceStateKey matches the latest-state view's partition key so a report's
// resource can be matched to its previously-stored state.
func resourceStateKey(installComponentID, provider, apiGroup, kind, namespace, name string) string {
	return strings.Join([]string{installComponentID, provider, apiGroup, kind, namespace, name}, "\x00")
}

// priorHealth is a resource's last stored health, so a degraded resource doesn't
// fall back to progressing when the k8s Warning event ages out.
type priorHealth struct {
	health       string
	message      string
	nativeStatus string
}

func (s *service) priorResourceHealth(ctx context.Context, orgID, installID string) (map[string]priorHealth, error) {
	var rows []app.InstallComponentResourceState
	err := s.chDB.WithContext(ctx).
		Scopes(scopes.WithOverrideTable(views.CurrentViewName(s.chDB, &app.InstallComponentResourceState{}))).
		Select("install_component_id", "provider", "api_group", "kind", "namespace", "name", "health", "message", "native_status").
		Where(app.InstallComponentResourceState{OrgID: orgID, InstallID: installID}).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make(map[string]priorHealth, len(rows))
	for _, r := range rows {
		out[resourceStateKey(r.InstallComponentID, r.Provider, r.APIGroup, r.Kind, r.Namespace, r.Name)] = priorHealth{
			health:       r.Health,
			message:      r.Message,
			nativeStatus: r.NativeStatus,
		}
	}
	return out, nil
}

// latchHealth applies the sticky-degraded rule: a healthy report clears the latch,
// otherwise the worse of incoming vs prior is kept, with its matching message.
func latchHealth(inHealth, inMessage, inNative string, prior priorHealth, hasPrior bool) (string, string, string) {
	if inHealth == healthHealthy {
		return inHealth, inMessage, inNative
	}
	// unknown is absence of an assessment, so latching would republish a stale
	// diagnosis under a fresh timestamp as if just confirmed.
	if inHealth == healthUnknown {
		return inHealth, inMessage, inNative
	}
	if hasPrior && healthSeverity[prior.health] > healthSeverity[inHealth] {
		return prior.health, prior.message, prior.nativeStatus
	}
	return inHealth, inMessage, inNative
}

func (s *service) createComponentHealth(ctx context.Context, orgID, runnerID string, req CreateComponentHealthRequest) (*CreateComponentHealthResponse, error) {
	observedAt := req.ObservedAt
	if observedAt.IsZero() {
		observedAt = time.Now()
	}

	// Without the feature nothing reads these rows, so writing them bills the
	// org storage for a product it does not have.
	if enabled, _ := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureComponentHealth); !enabled {
		return &CreateComponentHealthResponse{Ingested: 0}, nil
	}

	prior, err := s.priorResourceHealth(ctx, orgID, req.InstallID)
	if err != nil {
		return nil, fmt.Errorf("unable to load prior resource health: %w", err)
	}

	rows := make([]app.InstallComponentResourceState, 0)
	for _, comp := range req.Components {
		resources := comp.Resources
		if len(resources) > maxResourcesPerComponent {
			s.l.Warn("truncating oversized component health resource list",
				zap.String("runner_id", runnerID),
				zap.String("install_component_id", comp.InstallComponentID),
				zap.Int("received", len(resources)),
				zap.Int("cap", maxResourcesPerComponent),
			)
			resources = resources[:maxResourcesPerComponent]
		}

		for _, r := range resources {
			nativeStatus := r.NativeStatus
			if nativeStatus == "" {
				nativeStatus = comp.NativeStatus
			}

			key := resourceStateKey(comp.InstallComponentID, r.Provider, r.APIGroup, r.Kind, r.Namespace, r.Name)
			p, ok := prior[key]
			health, message, nativeStatus := latchHealth(r.Health, r.Message, nativeStatus, p, ok)

			rows = append(rows, app.InstallComponentResourceState{
				OrgID:              orgID,
				InstallID:          req.InstallID,
				InstallComponentID: comp.InstallComponentID,
				ComponentID:        comp.ComponentID,
				RunnerID:           runnerID,
				Source:             sourceComponent,
				Provider:           r.Provider,
				APIGroup:           r.APIGroup,
				Kind:               r.Kind,
				Namespace:          r.Namespace,
				Name:               r.Name,
				Health:             health,
				Message:            message,
				NativeStatus:       nativeStatus,
				Details:            boundDetails(r.Details),
				ObservedAt:         observedAt,
			})
		}
	}

	var sandboxWorst, sandboxWorstMessage string
	for _, rel := range req.SandboxReleases {
		resources := rel.Resources
		if len(resources) > maxResourcesPerComponent {
			s.l.Warn("truncating oversized sandbox release resource list",
				zap.String("runner_id", runnerID),
				zap.String("release_name", rel.ReleaseName),
				zap.Int("received", len(resources)),
				zap.Int("cap", maxResourcesPerComponent),
			)
			resources = resources[:maxResourcesPerComponent]
		}

		for _, r := range resources {
			key := resourceStateKey("", r.Provider, r.APIGroup, r.Kind, r.Namespace, r.Name)
			p, ok := prior[key]
			health, message, nativeStatus := latchHealth(r.Health, r.Message, r.NativeStatus, p, ok)

			if healthSeverity[health] > healthSeverity[sandboxWorst] {
				sandboxWorst = health
				sandboxWorstMessage = message
			}

			rows = append(rows, app.InstallComponentResourceState{
				OrgID:        orgID,
				InstallID:    req.InstallID,
				RunnerID:     runnerID,
				Source:       sourceSandbox,
				OwnerName:    rel.ReleaseName,
				Provider:     r.Provider,
				APIGroup:     r.APIGroup,
				Kind:         r.Kind,
				Namespace:    r.Namespace,
				Name:         r.Name,
				Health:       health,
				Message:      message,
				NativeStatus: nativeStatus,
				Details:      boundDetails(r.Details),
				ObservedAt:   observedAt,
			})
		}
	}

	s.updateInstallSandboxHealth(ctx, req.InstallID, sandboxWorst, sandboxWorstMessage, true)
	s.ensureInstallHealthQueues(ctx, req.InstallID)

	if len(rows) == 0 {
		return &CreateComponentHealthResponse{Ingested: 0}, nil
	}

	res := s.chDB.WithContext(ctx).CreateInBatches(&rows, componentHealthInsertBatchSize)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to write resource states: %w", res.Error)
	}

	// Only after the observations are durable — evaluating before the write
	// lands would read the previous report and draw last cycle's verdict.
	s.triggerHealthEvaluation(ctx, req.InstallID)

	return &CreateComponentHealthResponse{Ingested: len(rows)}, nil
}

// ensureInstallHealthQueues lazily reconciles the install's health queue and its
// evaluator emitter, so installs predating the evaluator don't sit with
// observations flowing but no verdict. Memoized, best-effort.
func (s *service) ensureInstallHealthQueues(ctx context.Context, installID string) {
	if _, done := s.ensuredHealthQueues.Load(installID); done {
		return
	}

	if err := s.installsHelpers.EnsureInstallQueues(ctx, installID); err != nil {
		s.l.Warn("unable to ensure install queues from component health ingest",
			zap.String("install_id", installID), zap.Error(err))
		return
	}
	s.ensuredHealthQueues.Store(installID, struct{}{})
}

// triggerHealthEvaluation evaluates on the report that just arrived rather than
// waiting for a timer, so a verdict lands with the data that justifies it.
//
// Deduped per queue: while an evaluation is still pending, further reports
// collapse into it, so multiple runners on one install cannot pile up. Entirely
// best-effort — a queue problem must never fail a runner's health report.
func (s *service) triggerHealthEvaluation(ctx context.Context, installID string) {
	ownerType := plugins.TableName(s.db, app.Install{})

	var q app.Queue
	if err := s.db.WithContext(ctx).
		Select("id").
		Where(app.Queue{
			OwnerID:   installID,
			OwnerType: ownerType,
			Name:      installshelpers.InstallComponentHealthQueueName,
		}).
		First(&q).Error; err != nil {
		return
	}

	dedupeKey := componentHealthEvaluateSignalType
	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		OwnerID:   installID,
		OwnerType: ownerType,
		Signal: queuesignal.NewRaw(componentHealthEvaluateSignalType, map[string]any{
			"install_id": installID,
		}),
		DedupeKey: &dedupeKey,
	}); err != nil {
		s.l.Warn("unable to trigger component health evaluation",
			zap.String("install_id", installID), zap.Error(err))
	}
}

// updateInstallSandboxHealth denormalizes the worst sandbox-resource health onto
// the install so reads can surface a degraded sandbox without a ClickHouse query.
// Only degraded/unhealthy is recorded; best-effort (never fails the ingest).
func (s *service) updateInstallSandboxHealth(ctx context.Context, installID, worst, message string, enabled bool) {
	status := ""
	msg := ""
	if enabled && healthSeverity[worst] >= healthSeverity["degraded"] {
		status = worst
		msg = message
	}

	now := time.Now()
	if err := s.db.WithContext(ctx).
		Model(&app.Install{ID: installID}).
		Select("sandbox_health_status", "sandbox_health_message", "last_health_report_at").
		Updates(app.Install{
			SandboxHealthStatus:  status,
			SandboxHealthMessage: msg,
			LastHealthReportAt:   &now,
		}).Error; err != nil {
		s.l.Warn("unable to update install sandbox health rollup",
			zap.String("install_id", installID), zap.Error(err))
	}
}

func boundDetails(details string) string {
	if len(details) > maxResourceDetailsBytes {
		// truncating mid-JSON would corrupt the blob for consumers, so drop it
		// for a valid marker instead. The runner is expected to bound this first.
		return `{"_truncated":true}`
	}
	return details
}
