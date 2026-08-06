package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/bins/runner/internal/jobs/actions"
	"github.com/nuonco/nuon/bins/runner/internal/jobs/deploy"
	"github.com/nuonco/nuon/bins/runner/internal/jobs/sandbox"
	"github.com/nuonco/nuon/bins/runner/internal/jobs/sync"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/componenthealth"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/drain"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/metrics"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/process"
	"github.com/nuonco/nuon/pkg/runner/airgap"
	"github.com/nuonco/nuon/pkg/runner/airgap/day2"
	"github.com/nuonco/nuon/pkg/runner/airgap/day2run"
	"github.com/nuonco/nuon/pkg/runner/airgap/statestore"
	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	"github.com/nuonco/nuon/pkg/runner/errs"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	"github.com/nuonco/nuon/pkg/runner/oci/bundle"
	ocicopy "github.com/nuonco/nuon/pkg/runner/oci/copy"
	ociresolve "github.com/nuonco/nuon/pkg/runner/oci/resolve"
	"github.com/nuonco/nuon/pkg/runner/settings"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

type airgapOptions struct {
	plan                string
	state               string
	bundle              string
	stateS3             string
	workdir             string
	logLevel            string
	installStackOutputs string
	installInputs       string
	deploymentID        string
}

func (c *cli) registerAirgap() error {
	options := new(airgapOptions)
	cmd := &cobra.Command{
		Use:   "airgap",
		Short: "execute a plan envelope without a control plane",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return c.runAirgap(cmd.Context(), options)
		},
	}
	cmd.Flags().StringVar(&options.plan, "plan", "", "path to the plan envelope")
	cmd.Flags().StringVar(&options.state, "state", "", "directory for persistent runner state")
	cmd.Flags().StringVar(&options.bundle, "bundle", "", "S3 URI or local path to a .oci.tar.zst bundle")
	cmd.Flags().StringVar(&options.stateS3, "state-s3", "", "S3 URI used to restore and sync runner state")
	cmd.Flags().StringVar(&options.workdir, "workdir", "/tmp/airgap", "local directory for bundle and state data")
	cmd.Flags().StringVar(&options.logLevel, "log-level", "info", "log level")
	cmd.Flags().StringVar(&options.installStackOutputs, "install-stack-outputs", "", "path or s3:// URI to a JSON object with this environment's install stack outputs (rebinds plans rendered against the vendor's reference stack); an S3 URI is polled until the phone-home Lambda writes it")
	cmd.Flags().StringVar(&options.installInputs, "install-inputs", "", "path or s3:// URI to a JSON object mapping install input names to values; substituted into step plans in place of publish-time placeholders")
	cmd.Flags().StringVar(&options.deploymentID, "deployment-id", "", "short deployment identifier (1-8 lowercase letters or digits) spliced into the bundle's frozen install ID so a second deployment of the same bundle in one account gets distinct physical resource names")
	rootCmd.AddCommand(cmd)
	return nil
}

func (c *cli) runAirgap(ctx context.Context, options *airgapOptions) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	if (options.plan == "") == (options.bundle == "") {
		return fmt.Errorf("exactly one of --plan or --bundle is required")
	}
	if options.plan != "" && options.state == "" {
		return fmt.Errorf("--state is required with --plan")
	}

	logger, err := newAirgapLogger(options.logLevel)
	if err != nil {
		return err
	}
	defer logger.Sync()

	bundleDir := "/bundle"
	if options.bundle != "" {
		options.plan, options.state, bundleDir, err = prepareAirgapBundle(ctx, options)
		if err != nil {
			return err
		}
	}

	runCtx, cancelRun := context.WithCancel(ctx)
	defer cancelRun()

	var syncer *airgapS3Sync
	var stopSync context.CancelFunc
	var syncDone chan struct{}
	var leaseLost atomic.Bool
	owner := airgapLeaseOwner()
	// ownerCtx gates writes that must only happen while we own the lease
	// (final state upload, DONE marker). It survives SIGTERM but is canceled
	// the moment the lease is lost.
	ownerCtx := context.Background()
	if options.stateS3 != "" {
		syncer, err = newAirgapS3Sync(ctx, options.stateS3)
		if err != nil {
			return err
		}
		lease := newAirgapLease(syncer.client, syncer.bucket, syncer.objectKey(airgapLeaseObject), owner, airgapLeaseTTL)
		if err := lease.Acquire(runCtx); err != nil {
			return err
		}
		// The lease outlives runCtx: on SIGTERM it keeps being renewed through
		// shutdown and the final state sync, and is released only afterwards.
		leaseCtx, cancelLease := context.WithCancel(context.Background())
		var cancelOwner context.CancelFunc
		ownerCtx, cancelOwner = context.WithCancel(context.Background())
		defer cancelOwner()
		renewDone := make(chan struct{})
		onLeaseLost := func() {
			leaseLost.Store(true)
			cancelOwner()
			cancelRun()
			// Cancellation is cooperative; in-flight terraform/helm work may
			// not stop before a successor becomes eligible. Fail-stop rather
			// than keep executing without the lease.
			time.AfterFunc(airgapLeaseFailStopGrace, func() {
				logger.Error("deployment lease lost and shutdown did not finish in time; terminating")
				os.Exit(1)
			})
		}
		go func() {
			defer close(renewDone)
			lease.renewLoop(leaseCtx, logger, onLeaseLost)
		}()
		defer func() {
			cancelLease()
			select {
			case <-renewDone:
			case <-time.After(airgapLeaseRenewTimeout + time.Second):
				logger.Warn("lease renewal did not stop in time; skipping release so the successor waits out the TTL")
				return
			}
			if leaseLost.Load() {
				return
			}
			releaseCtx, releaseCancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer releaseCancel()
			if releaseErr := lease.Release(releaseCtx); releaseErr != nil {
				logger.Warn("unable to release deployment lease; a successor must wait out the TTL", zap.Error(releaseErr))
			}
		}()
		if err := syncer.downloadPrefix(runCtx, options.state); err != nil {
			return fmt.Errorf("restore state: %w", err)
		}
		var syncCtx context.Context
		syncCtx, stopSync = context.WithCancel(runCtx)
		defer stopSync()
		syncDone = make(chan struct{})
		go func() {
			defer close(syncDone)
			syncer.syncLoop(syncCtx, options.state, 30*time.Second)
		}()
	}

	onBootstrapDone := func(cbCtx context.Context) error {
		if syncer == nil {
			return nil
		}
		if err := syncer.uploadDir(cbCtx, options.state); err != nil {
			return fmt.Errorf("upload state: %w", err)
		}
		return syncer.writeDone(cbCtx, "success")
	}

	var day2Deps *airgapDay2Deps
	if syncer != nil {
		day2Deps = &airgapDay2Deps{syncer: syncer, owner: owner}
	}
	err = c.executeAirgap(runCtx, options, bundleDir, logger, onBootstrapDone, day2Deps)
	if syncer != nil {
		stopSync()
		<-syncDone
		if leaseLost.Load() {
			return errors.Join(err, fmt.Errorf("deployment lease lost; skipping final state sync so the new owner's state is not overwritten"))
		}
		finalCtx, finalCancel := context.WithTimeout(ownerCtx, 2*time.Minute)
		finalErr := syncer.uploadDir(finalCtx, options.state)
		finalCancel()
		err = errors.Join(err, finalErr)
	}
	if err == nil && syncer != nil {
		if leaseLost.Load() {
			return fmt.Errorf("deployment lease lost during final state sync; not writing DONE")
		}
		markerCtx, cancel := context.WithTimeout(ownerCtx, 30*time.Second)
		defer cancel()
		if markerErr := syncer.writeDone(markerCtx, "success"); markerErr != nil {
			return fmt.Errorf("write DONE marker: %w", markerErr)
		}
	}
	return err
}

func prepareAirgapBundle(ctx context.Context, options *airgapOptions) (string, string, string, error) {
	if err := os.MkdirAll(options.workdir, 0o700); err != nil {
		return "", "", "", fmt.Errorf("create workdir: %w", err)
	}
	archivePath := filepath.Join(options.workdir, "bundle.oci.tar.zst")
	if err := downloadAirgapBundle(ctx, options.bundle, archivePath); err != nil {
		return "", "", "", err
	}
	extractDir := filepath.Join(options.workdir, "bundle")
	if err := os.RemoveAll(extractDir); err != nil {
		return "", "", "", fmt.Errorf("reset bundle directory: %w", err)
	}
	f, err := os.Open(archivePath)
	if err != nil {
		return "", "", "", fmt.Errorf("open bundle: %w", err)
	}
	_, extractErr := bundle.Extract(extractDir, f)
	closeErr := f.Close()
	if extractErr != nil {
		return "", "", "", fmt.Errorf("extract bundle: %w", extractErr)
	}
	if closeErr != nil {
		return "", "", "", fmt.Errorf("close bundle: %w", closeErr)
	}
	if err := bundle.VerifyBlobs(extractDir); err != nil {
		return "", "", "", fmt.Errorf("verify bundle blobs: %w", err)
	}
	b, err := bundle.Open(ctx, extractDir)
	if err != nil {
		return "", "", "", fmt.Errorf("read bundle manifest: %w", err)
	}
	if len(b.PlanEnvelope) == 0 {
		return "", "", "", fmt.Errorf("bundle has no plan envelope")
	}
	if _, err := airgap.Parse(b.PlanEnvelope); err != nil {
		return "", "", "", fmt.Errorf("parse bundle plan envelope: %w", err)
	}
	planPath := filepath.Join(options.workdir, "plan-envelope.json")
	if err := os.WriteFile(planPath, b.PlanEnvelope, 0o600); err != nil {
		return "", "", "", fmt.Errorf("write plan envelope: %w", err)
	}
	return planPath, filepath.Join(options.workdir, "state"), extractDir, nil
}

func newAirgapLogger(logLevel string) (*zap.Logger, error) {
	logCfg := zap.NewProductionConfig()
	if level, parseErr := zap.ParseAtomicLevel(logLevel); parseErr == nil {
		logCfg.Level = level
	}
	return logCfg.Build()
}

type airgapDay2Deps struct {
	syncer *airgapS3Sync
	owner  string
}

func (c *cli) executeAirgap(ctx context.Context, options *airgapOptions, bundleDir string, logger *zap.Logger, onBootstrapDone func(context.Context) error, day2Deps *airgapDay2Deps) error {
	rawEnvelope, err := os.ReadFile(options.plan)
	if err != nil {
		return fmt.Errorf("read plan envelope: %w", err)
	}
	envelope, err := airgap.Parse(rawEnvelope)
	if err != nil {
		return err
	}
	frozenInstallID := envelope.InstallID
	if options.deploymentID != "" {
		if _, err := envelope.ApplyDeploymentID(options.deploymentID); err != nil {
			return err
		}
	}
	store, err := statestore.NewDisk(options.state)
	if err != nil {
		return err
	}
	if options.deploymentID != "" {
		logger.Info("deployment id applied to plan envelope",
			zap.String("frozen_install_id", frozenInstallID),
			zap.String("deployment_install_id", envelope.InstallID))
	}
	copierProvider := fx.Provide(ocicopy.New)
	archiveSourceProvider := fx.Options()
	if b, openErr := bundle.Open(ctx, bundleDir); openErr == nil {
		source, err := airgap.NewBundleSource(b.Store(), b.Members(), b.Provenance)
		if err != nil {
			return err
		}
		if missing := source.MissingPlanSources(envelope); len(missing) > 0 {
			return fmt.Errorf("bundle does not package these plan sources: %s", strings.Join(missing, "; "))
		}
		copierProvider = fx.Provide(func(params ocicopy.CopierParams) ocicopy.Copier {
			return source.Copier(ocicopy.New(params))
		})
		archiveSourceProvider = fx.Provide(func() ociarchive.Source { return source })
		logger.Info("plan oci sources will be served from the bundle", zap.String("bundle_dir", bundleDir))
	} else if options.bundle != "" {
		return fmt.Errorf("open bundle for oci sources: %w", openErr)
	} else {
		logger.Warn("no bundle available; sync-oci sources will be pulled from their original registries", zap.String("bundle_dir", bundleDir))
	}
	backend, err := airgap.NewTFBackend(store, filepath.Join(options.state, "tfbackend-port"))
	if err != nil {
		return err
	}
	defer backend.Close()
	client, err := airgap.NewClient(envelope, store, logger)
	if err != nil {
		return err
	}
	if options.installStackOutputs != "" {
		var raw []byte
		if strings.HasPrefix(options.installStackOutputs, "s3://") {
			raw, err = fetchInstallStackOutputs(ctx, options.installStackOutputs, logger)
		} else {
			raw, err = os.ReadFile(options.installStackOutputs)
		}
		if err != nil {
			return fmt.Errorf("read install stack outputs: %w", err)
		}
		var outputs map[string]any
		if err := json.Unmarshal(raw, &outputs); err != nil {
			return fmt.Errorf("parse install stack outputs %s: %w", options.installStackOutputs, err)
		}
		client.SetInstallStackOutputs(outputs)
		logger.Info("install stack outputs will be rebound into step plans", zap.String("path", options.installStackOutputs), zap.Int("keys", len(outputs)))
	}
	if options.installInputs != "" {
		var raw []byte
		if strings.HasPrefix(options.installInputs, "s3://") {
			raw, err = fetchInstallStackOutputs(ctx, options.installInputs, logger)
		} else {
			raw, err = os.ReadFile(options.installInputs)
		}
		if err != nil {
			return fmt.Errorf("read install inputs: %w", err)
		}
		var values map[string]string
		if err := json.Unmarshal(raw, &values); err != nil {
			return fmt.Errorf("parse install inputs %s: %w", options.installInputs, err)
		}
		if err := airgap.ValidateInputValues(envelope.Inputs, values); err != nil {
			return err
		}
		client.SetInstallInputs(values)
		logger.Info("install inputs will be substituted into step plans", zap.String("path", options.installInputs), zap.Int("keys", len(values)))
	}
	cfg := &runnerconfig.Config{
		GitRef: "airgap", RunnerAPIURL: "http://" + backend.Addr(), RunnerAPIToken: "airgap",
		RunnerID: "airgap-" + envelope.InstallID, HostIP: "127.0.0.1", LogLevel: options.logLevel,
		JobLogDir: filepath.Join(options.state, "job-logs"),
		BundleDir: bundleDir, RegistryDir: options.state + "/registry", RegistryPort: 5001,
		HealthPort: 9999, SandboxJobDuration: 5 * time.Second, SandboxControlPort: 9095,
	}
	runtimeSettings := &settings.Settings{
		HeartBeatTimeout: 30 * time.Second, JobLoopMinPollPeriod: time.Second, LongPollJobs: false,
		EnableLogging: false, LoggingLevel: slog.LevelInfo, EnableMetrics: false,
		Groups: []string{"sandbox", "deploy", "sync", "actions"}, Metadata: map[string]string{},
		OTELConfiguration: "{}", OtelSchemaURL: cfg.RunnerAPIURL, Platform: "airgap", Cfg: cfg,
	}
	drainer := drain.New()

	providers := []fx.Option{
		fx.Provide(func() *runnerconfig.Config { return cfg }),
		fx.Provide(func() *settings.Settings { return runtimeSettings }),
		fx.Provide(func() nuonrunner.Client { return client }),
		fx.Provide(func() *zap.Logger { return logger }),
		fx.Provide(fx.Annotate(func() *zap.Logger { return logger }, fx.ResultTags(`name:"system"`))),
		fx.Provide(validator.New),
		fx.Provide(errs.NewRecorder),
		fx.Provide(func() *drain.Drainer { return drainer }),
		fx.Provide(metrics.New),
		fx.Provide(process.New),
		fx.Provide(fx.Annotate(func() string { return "install" }, fx.ResultTags(`name:"process"`))),
		fx.Provide(componenthealth.NewClusterProvider),
		fx.Provide(componenthealth.NewTerraformProvider),
		fx.Provide(componenthealth.NewManifestKindsProvider),
		fx.Provide(componenthealth.New),
		fx.Invoke(func(*componenthealth.Engine) {}),
		copierProvider,
		archiveSourceProvider,
		fx.Provide(ociresolve.New),
		fx.Invoke(jobloop.WithJobLoops(func([]jobloop.JobLoop) {})),
	}
	providers = append(providers, sandbox.GetJobs()...)
	providers = append(providers, deploy.GetJobs()...)
	providers = append(providers, sync.GetJobs()...)
	providers = append(providers, actions.GetJobs()...)

	app := fx.New(providers...)
	startCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := app.Start(startCtx); err != nil {
		return fmt.Errorf("start airgap runner: %w", err)
	}
	select {
	case <-ctx.Done():
	case <-client.Done():
		result := client.Result()
		logger.Info("airgap bootstrap complete", zap.Bool("succeeded", result.Succeeded), zap.String("failed_step", result.FailedStep))
		var publishErr error
		if result.Succeeded && onBootstrapDone != nil {
			if publishErr = onBootstrapDone(ctx); publishErr != nil {
				logger.Error("publish bootstrap completion", zap.Error(publishErr))
			}
		}
		if result.Succeeded {
			if day2Deps != nil {
				executor := day2run.NewExecutor(client, envelope, store)
				controller, controllerErr := day2run.NewController(day2run.ControllerConfig{
					Mailbox:  day2run.NewMailbox(day2Deps.syncer.client, day2Deps.syncer.bucket, day2Deps.syncer.prefix, logger),
					Envelope: envelope, Digest: day2.EnvelopeDigest(rawEnvelope), DeploymentID: envelope.InstallID,
					Owner: day2Deps.owner, Executor: executor, Logger: logger, WriteLocal: store.WriteFile,
					FlushRun: func(flushCtx context.Context, runID string) error {
						return day2Deps.syncer.uploadSubdir(flushCtx, filepath.Join(options.state, "runs", runID), day2.RunsPrefix+runID)
					},
				})
				if controllerErr != nil {
					logger.Error("configure day-2 dispatch", zap.Error(controllerErr))
				} else {
					go func() {
						if err := controller.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
							logger.Error("day-2 dispatch stopped", zap.Error(err))
						}
					}()
				}
			} else {
				logger.Info("day-2 dispatch requires --state-s3")
			}
			logger.Info("bootstrap complete; staying resident to report component health; send SIGTERM to stop")
			// The DONE marker is deliberately excluded from bulk state sync, so
			// a resident runner must keep retrying publication itself or the
			// customer never sees the deployment finish.
			for publishErr != nil {
				select {
				case <-ctx.Done():
					publishErr = nil
				case <-time.After(30 * time.Second):
					if publishErr = onBootstrapDone(ctx); publishErr != nil {
						logger.Warn("retrying bootstrap completion publication", zap.Error(publishErr))
					}
				}
			}
			<-ctx.Done()
		}
	}
	drainer.Drain()
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer stopCancel()
	if err := app.Stop(stopCtx); err != nil {
		return fmt.Errorf("stop airgap runner: %w", err)
	}
	result := client.Result()
	if !result.Succeeded {
		return fmt.Errorf("airgap run failed at step %q", result.FailedStep)
	}
	return nil
}
