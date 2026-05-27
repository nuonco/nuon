package processhealthcheck

import (
	"fmt"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/v2/oninactive"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "process_healthcheck"

const (
	offlineTimeout  = 1 * time.Minute
	inactiveTimeout = 5 * time.Minute
)

type Signal struct {
	RunnerID  string `json:"runner_id"`
	ProcessID string `json:"process_id"`

	mw metrics.Writer
	v  *validator.Validate
}

var (
	_ signal.Signal           = (*Signal)(nil)
	_ signal.SleepAfter       = (*Signal)(nil)
	_ signal.SignalWithParams = (*Signal)(nil)
)

func (s *Signal) WithParams(params *signal.Params) {
	s.mw = params.MW
	s.v = params.V
}

func (s *Signal) SleepAfter() time.Duration {
	return 0
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) MaxInFlightAge() time.Duration {
	return 2 * time.Minute
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}
	if s.ProcessID == "" {
		return errors.New("process_id is required")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) (err error) {
	tmw, err := tmetrics.New(s.v, tmetrics.WithMetricsWriter(s.mw))
	if err != nil {
		return errors.Wrap(err, "unable to create temporal metrics writer")
	}

	// tags is the running set of dimensions emitted on every metric in this
	// signal. Mutate it as runner / install / process facts resolve; the
	// deferred latency emit captures the final state. Owner labels are added
	// with collision guards so standard tags always win.
	tags := map[string]string{
		"runner_id":    s.RunnerID,
		"process_type": "unknown",
	}
	addLabels := func(labels map[string]string) {
		for k, v := range labels {
			if _, ok := tags[k]; !ok {
				tags[k] = v
			}
		}
	}
	start := workflow.Now(ctx)
	defer func() {
		tmw.Timing(ctx, "runner.health_check.latency", workflow.Now(ctx).Sub(start), metrics.ToTags(tags)...)
	}()

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get logger")
	}

	runner, err := activities.AwaitGet(ctx, activities.GetRequest{RunnerID: s.RunnerID})
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}
	tags["runner_type"] = string(runner.RunnerGroup.Type)
	tags["runner_status"] = string(runner.Status)
	tags["org_id"] = runner.OrgID
	tags["org_name"] = runner.Org.Name
	switch runner.RunnerGroup.OwnerType {
	case "installs":
		tags["install_id"] = runner.RunnerGroup.OwnerID
		install, err := activities.AwaitGetRunnerInstallByInstallID(ctx, runner.RunnerGroup.OwnerID)
		if err != nil {
			return errors.Wrap(err, "unable to get install for runner")
		}
		tags["install_name"] = install.Name
		addLabels(install.Labels)
	case "orgs":
		addLabels(runner.Org.Labels)
	}

	process, err := activities.AwaitGetRunnerProcessByProcessID(ctx, s.ProcessID)
	if err != nil {
		return dbgenerics.TemporalGormError(err, "runner process not found")
	}
	tags["process_type"] = string(process.Type)

	// Only run health checks for active or offline processes; noop for any other status
	switch process.ProcessStatus() {
	case app.RunnerProcessStatusActive, app.RunnerProcessStatusOffline:
	default:
		return nil
	}

	if handled, err := s.handleShutdownRequested(ctx, l, process); handled || err != nil {
		return err
	}

	heartbeat, err := activities.AwaitGetMostRecentHeartBeatByProcess(ctx, activities.GetMostRecentHeartBeatByProcessRequest{
		RunnerID:  s.RunnerID,
		ProcessID: s.ProcessID,
	})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Wrap(err, "unable to get heartbeat")
	}
	if heartbeat == nil {
		// No heartbeat received yet — cron health checks are a no-op until
		// MaybeEnqueueInitialHealthCheck fires on the first heartbeat.
		return nil
	}
	tags["runner_version"] = heartbeat.Version

	switch heartbeatAge := s.heartbeatAge(workflow.Now(ctx), heartbeat); {
	case heartbeatAge >= inactiveTimeout:
		return s.handleInactive(ctx, tmw, l, tags, process)
	case heartbeatAge >= offlineTimeout:
		return s.handleOffline(ctx, l, process)
	}

	// Heartbeat is fresh — ensure process is active, record a green health
	// check, then run the version-mismatch check.
	if err := s.handleActive(ctx, l, process); err != nil {
		return err
	}
	return s.checkVersionMismatch(ctx, l, runner, process, heartbeat, tags)
}

// handleShutdownRequested checks for the shutdown_requested metadata flag
// (set by a promotion) and creates a graceful shutdown record if present.
// The runner's shutdown poller will pick it up. Health check emitters run
// per-process on a 1-minute cron, so shutdowns are staggered naturally.
//
// Returns handled=true when a shutdown was triggered; the caller should
// stop processing. Failure to clear the flag is logged but not propagated
// because the next tick will retry.
func (s *Signal) handleShutdownRequested(ctx workflow.Context, l *zap.Logger, process *app.RunnerProcess) (bool, error) {
	val, ok := process.CompositeStatus.Metadata["shutdown_requested"]
	if !ok || val == nil {
		return false, nil
	}

	l.Info("shutdown requested via metadata, creating shutdown record",
		zap.String("runner_id", s.RunnerID),
		zap.String("process_id", s.ProcessID),
	)

	if _, err := activities.AwaitCreateRunnerProcessShutdown(ctx, activities.CreateRunnerProcessShutdownRequest{
		RunnerProcessID: s.ProcessID,
		Type:            app.RunnerProcessShutdownTypeGraceful,
		CompositeStatus: app.CompositeStatus{
			Status:                 app.Status(app.RunnerProcessShutdownStatusRequested),
			StatusHumanDescription: "shutdown requested via promotion",
			CreatedAtTS:            workflow.Now(ctx).Unix(),
		},
	}); err != nil {
		return true, errors.Wrap(err, "unable to create shutdown for process")
	}

	if err := activities.AwaitClearProcessShutdownRequested(ctx, activities.ClearProcessShutdownRequestedRequest{
		ProcessID: s.ProcessID,
	}); err != nil {
		l.Warn("unable to clear shutdown_requested metadata",
			zap.String("process_id", s.ProcessID),
			zap.Error(err),
		)
	}

	return true, nil
}

// handleInactive runs the Tier 1 transition when no heartbeat has arrived
// for inactiveTimeout: marks the process inactive, emits stop metrics,
// enqueues the on_inactive signal, and stops the per-process queue (which
// terminates the cron emitter). Sub-errors after the status update are
// logged but not propagated.
func (s *Signal) handleInactive(ctx workflow.Context, tmw tmetrics.Writer, l *zap.Logger, tags map[string]string, process *app.RunnerProcess) error {
	l.Warn("process inactive - no heartbeat for 5 minutes, stopping queue",
		zap.String("runner_id", s.RunnerID),
		zap.String("process_id", s.ProcessID),
	)

	if _, err := activities.AwaitUpdateRunnerProcessStatus(ctx, activities.UpdateRunnerProcessStatusRequest{
		ProcessID:         s.ProcessID,
		Status:            app.RunnerProcessStatusInactive,
		StatusDescription: "no heartbeat received for 5 minutes",
	}); err != nil {
		return errors.Wrap(err, "unable to update process status to inactive")
	}

	stopTags := metrics.ToTags(tags)
	tmw.Incr(ctx, "runner.process.stop", stopTags...)
	if process.StartedAt != nil {
		tmw.Timing(ctx, "runner.process.latency", workflow.Now(ctx).Sub(*process.StartedAt), stopTags...)
	}

	if _, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   s.RunnerID,
		OwnerType: "runners",
		QueueName: fmt.Sprintf("runner-process-%s", s.ProcessID),
		Signal: &oninactive.Signal{
			RunnerID:  s.RunnerID,
			ProcessID: s.ProcessID,
			Reason:    "offline",
		},
	}); err != nil {
		l.Warn("unable to enqueue on_inactive signal",
			zap.String("process_id", s.ProcessID),
			zap.Error(err),
		)
	}

	if err := activities.AwaitStopProcessQueue(ctx, activities.StopProcessQueueRequest{
		RunnerID:  s.RunnerID,
		ProcessID: s.ProcessID,
	}); err != nil {
		l.Warn("unable to stop process queue",
			zap.String("process_id", s.ProcessID),
			zap.Error(err),
		)
	}

	return nil
}

// handleOffline runs the Tier 2 transition when no heartbeat has arrived
// for offlineTimeout: marks the process offline (if not already) and
// records a red health check. Failure to record the health check is
// logged but not propagated — the next tick will retry.
func (s *Signal) handleOffline(ctx workflow.Context, l *zap.Logger, process *app.RunnerProcess) error {
	if process.ProcessStatus() != app.RunnerProcessStatusOffline {
		l.Warn("process offline - no heartbeat for 1 minute",
			zap.String("runner_id", s.RunnerID),
			zap.String("process_id", s.ProcessID),
		)

		if _, err := activities.AwaitUpdateRunnerProcessStatus(ctx, activities.UpdateRunnerProcessStatusRequest{
			ProcessID:         s.ProcessID,
			Status:            app.RunnerProcessStatusOffline,
			StatusDescription: "Runner is offline and will be marked inactive in 5 minutes",
		}); err != nil {
			return errors.Wrap(err, "unable to update process status to offline")
		}
	}

	if _, err := activities.AwaitCreateHealthCheck(ctx, activities.CreateHealthCheckRequest{
		RunnerID:  s.RunnerID,
		ProcessID: s.ProcessID,
		Status:    app.RunnerStatusError,
	}); err != nil {
		l.Warn("unable to create offline health check",
			zap.String("process_id", s.ProcessID),
			zap.Error(err),
		)
	}

	return nil
}

// handleActive runs the happy-path transition when the heartbeat is
// fresh: flips a previously-offline process back to active and records a
// green health check.
func (s *Signal) handleActive(ctx workflow.Context, l *zap.Logger, process *app.RunnerProcess) error {
	if process.ProcessStatus() == app.RunnerProcessStatusOffline {
		l.Info("process back online",
			zap.String("runner_id", s.RunnerID),
			zap.String("process_id", s.ProcessID),
		)
		if _, err := activities.AwaitUpdateRunnerProcessStatus(ctx, activities.UpdateRunnerProcessStatusRequest{
			ProcessID:         s.ProcessID,
			Status:            app.RunnerProcessStatusActive,
			StatusDescription: "heartbeat received",
		}); err != nil {
			return errors.Wrap(err, "unable to update process status to active")
		}
	}

	if _, err := activities.AwaitCreateHealthCheck(ctx, activities.CreateHealthCheckRequest{
		RunnerID:  s.RunnerID,
		ProcessID: s.ProcessID,
		Status:    app.RunnerStatusActive,
	}); err != nil {
		return errors.Wrap(err, "unable to create process health check")
	}

	return nil
}

// checkVersionMismatch compares the configured runner version to the
// heartbeat-reported version, emits a Datadog event when 'latest' is in
// use, and writes the corresponding version_warning into process metadata
// (empty string when versions agree).
func (s *Signal) checkVersionMismatch(ctx workflow.Context, l *zap.Logger, runner *app.Runner, process *app.RunnerProcess, heartbeat *app.RunnerHeartBeat, tags map[string]string) error {
	if heartbeat == nil || heartbeat.Version == "" {
		return nil
	}

	settings := runner.RunnerGroup.Settings
	var configuredVersion string
	switch process.Type {
	case app.RunnerProcessTypeMng:
		configuredVersion = settings.BinaryVersion
	default:
		configuredVersion = settings.ContainerImageTag
	}

	if configuredVersion == "latest" || heartbeat.Version == "latest" {
		eventTags := metrics.ToTags(tags,
			metrics.ToTag("configured_version", configuredVersion),
			metrics.ToTag("heartbeat_version", heartbeat.Version),
		)
		s.mw.Event(&statsd.Event{
			Title: "Runner is using 'latest' version tag",
			Text: fmt.Sprintf(
				"Runner %s (org: %s) is using the 'latest' tag. configured_version=%q, reported_version=%q. Pin to a specific version to avoid drift.",
				s.RunnerID, runner.Org.Name, configuredVersion, heartbeat.Version,
			),
			Tags:           eventTags,
			SourceTypeName: "nuon-runner",
			Priority:       statsd.Normal,
			AlertType:      statsd.Error,
			AggregationKey: "runner-version-latest",
		})
	}

	isAliasTag := configuredVersion != "" && func() bool {
		_, err := semver.NewVersion(configuredVersion)
		return err != nil
	}()

	var versionWarning string
	switch {
	case configuredVersion == "" || configuredVersion == heartbeat.Version:
		// No warning needed.
	case configuredVersion == "cloud":
		// "cloud" means "track the API version" — this is expected for
		// Nuon-managed cloud runners, so no mismatch warning.
	case isAliasTag:
		versionWarning = fmt.Sprintf(
			"Runner is configured with alias tag (%s). We recommend pinning a specific version to avoid drift.",
			configuredVersion,
		)
	default:
		versionWarning = fmt.Sprintf(
			"Reported runner version (%s) does not match configured version (%s). Please update the runner to the correct version.",
			heartbeat.Version, configuredVersion,
		)
	}

	if _, err := activities.AwaitUpdateRunnerProcessStatus(ctx, activities.UpdateRunnerProcessStatusRequest{
		ProcessID:         s.ProcessID,
		Status:            app.RunnerProcessStatusActive,
		StatusDescription: "",
		Metadata: map[string]any{
			"version_warning": versionWarning,
		},
	}); err != nil {
		l.Warn("unable to update version warning metadata",
			zap.String("process_id", s.ProcessID),
			zap.Error(err),
		)
	}

	return nil
}

func (s *Signal) heartbeatAge(now time.Time, heartbeat *app.RunnerHeartBeat) time.Duration {
	if heartbeat == nil {
		return inactiveTimeout
	}
	return now.Sub(heartbeat.CreatedAt)
}
