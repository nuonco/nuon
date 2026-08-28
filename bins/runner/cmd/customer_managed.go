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
	"github.com/google/uuid"
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
	"github.com/nuonco/nuon/bins/runner/internal/pkg/telemetryexport"
	bundle "github.com/nuonco/nuon/pkg/customer_managed/bundle"
	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operationrun"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/statestore"
	"github.com/nuonco/nuon/pkg/runner/errs"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	ocicopy "github.com/nuonco/nuon/pkg/runner/oci/copy"
	ociresolve "github.com/nuonco/nuon/pkg/runner/oci/resolve"
	"github.com/nuonco/nuon/pkg/runner/settings"
	"github.com/nuonco/nuon/pkg/runner/version"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

type customerManagedOptions struct {
	plan                string
	state               string
	bundle              string
	stateS3             string
	workdir             string
	logLevel            string
	installStackOutputs string
	installInputs       string
	deploymentID        string

	archiveDigest string
}

func (c *cli) registerCustomerManaged() error {
	options := new(customerManagedOptions)
	cmd := &cobra.Command{
		Use:   "customer-managed",
		Short: "execute a plan envelope without a control plane",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return c.runCustomerManaged(cmd.Context(), options)
		},
	}
	cmd.Flags().StringVar(&options.plan, "plan", "", "path to the plan envelope")
	cmd.Flags().StringVar(&options.state, "state", "", "directory for persistent runner state")
	cmd.Flags().StringVar(&options.bundle, "bundle", "", "S3 URI or local path to a .oci.tar.zst bundle")
	cmd.Flags().StringVar(&options.stateS3, "state-s3", "", "S3 URI used to restore and sync runner state")
	cmd.Flags().StringVar(&options.workdir, "workdir", "/tmp/customer_managed", "local directory for bundle and state data")
	cmd.Flags().StringVar(&options.logLevel, "log-level", "info", "log level")
	cmd.Flags().StringVar(&options.installStackOutputs, "install-stack-outputs", "", "path or s3:// URI to a JSON object with this environment's install stack outputs (rebinds plans rendered against the vendor's reference stack); an S3 URI is polled until the phone-home Lambda writes it")
	cmd.Flags().StringVar(&options.installInputs, "install-inputs", "", "path or s3:// URI to a JSON object mapping install input names to values; substituted into step plans in place of publish-time placeholders")
	cmd.Flags().StringVar(&options.deploymentID, "deployment-id", "", "short deployment identifier (1-8 lowercase letters or digits) spliced into the bundle's frozen install ID so a second deployment of the same bundle in one account gets distinct physical resource names")
	rootCmd.AddCommand(cmd)
	return nil
}

func (c *cli) runCustomerManaged(ctx context.Context, options *customerManagedOptions) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	if (options.plan == "") == (options.bundle == "") {
		return fmt.Errorf("exactly one of --plan or --bundle is required")
	}
	if options.plan != "" && options.state == "" {
		return fmt.Errorf("--state is required with --plan")
	}

	logger, err := newCustomerManagedLogger(options.logLevel)
	if err != nil {
		return err
	}
	defer logger.Sync()

	bundleDir := "/bundle"
	if options.bundle != "" {
		options.plan, options.state, bundleDir, err = prepareCustomerManagedBundle(ctx, options)
		if err != nil {
			return err
		}
	}

	runCtx, cancelRun := context.WithCancel(ctx)
	defer cancelRun()

	var syncer *customerManagedS3Sync
	var stopSync context.CancelFunc
	var syncDone chan struct{}
	var leaseLost atomic.Bool
	owner := customerManagedLeaseOwner()
	runnerStateDir := options.state
	// ownerCtx gates writes that must only happen while we own the lease
	// (final state upload, DONE marker). It survives SIGTERM but is canceled
	// the moment the lease is lost.
	ownerCtx := context.Background()
	if options.stateS3 != "" {
		runnerStateDir = filepath.Join(options.state, "runner")
		if err := migrateLegacyLocalRunnerState(options.state, runnerStateDir); err != nil {
			return fmt.Errorf("migrate local runner state: %w", err)
		}
		syncer, err = newCustomerManagedS3Sync(ctx, options.stateS3)
		if err != nil {
			return err
		}
		lease := newCustomerManagedLease(syncer.client, syncer.bucket, syncer.objectKey(customerManagedLeaseObject), owner, customerManagedLeaseTTL)
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
			time.AfterFunc(customerManagedLeaseFailStopGrace, func() {
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
			case <-time.After(customerManagedLeaseRenewTimeout + time.Second):
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
		if err := syncer.restoreRunnerState(runCtx, runnerStateDir); err != nil {
			return fmt.Errorf("restore state: %w", err)
		}
		var syncCtx context.Context
		syncCtx, stopSync = context.WithCancel(runCtx)
		defer stopSync()
		syncDone = make(chan struct{})
		go func() {
			defer close(syncDone)
			syncer.syncLoop(syncCtx, runnerStateDir, 30*time.Second)
		}()
	}

	onBootstrapDone := func(cbCtx context.Context) error {
		if syncer == nil {
			return nil
		}
		if err := syncer.uploadDir(cbCtx, runnerStateDir); err != nil {
			return fmt.Errorf("upload state: %w", err)
		}
		return syncer.writeDone(cbCtx, "success")
	}

	var operationDeps *customerManagedOperationDeps
	if syncer != nil {
		operationDeps = &customerManagedOperationDeps{syncer: syncer, owner: owner}
	}
	err = c.executeCustomerManaged(runCtx, options, runnerStateDir, bundleDir, logger, onBootstrapDone, operationDeps)
	if syncer != nil {
		stopSync()
		<-syncDone
		if leaseLost.Load() {
			return errors.Join(err, fmt.Errorf("deployment lease lost; skipping final state sync so the new owner's state is not overwritten"))
		}
		finalCtx, finalCancel := context.WithTimeout(ownerCtx, 2*time.Minute)
		finalErr := syncer.uploadDir(finalCtx, runnerStateDir)
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

func prepareCustomerManagedBundle(ctx context.Context, options *customerManagedOptions) (string, string, string, error) {
	if err := os.MkdirAll(options.workdir, 0o700); err != nil {
		return "", "", "", fmt.Errorf("create workdir: %w", err)
	}
	archivePath := filepath.Join(options.workdir, "bundle.oci.tar.zst")
	if err := downloadCustomerManagedBundle(ctx, options.bundle, archivePath); err != nil {
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
	checksum, extractErr := bundle.Extract(extractDir, f)
	closeErr := f.Close()
	if extractErr != nil {
		return "", "", "", fmt.Errorf("extract bundle: %w", extractErr)
	}
	options.archiveDigest = "sha256:" + checksum
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
	if _, err := customermanaged.Parse(b.PlanEnvelope); err != nil {
		return "", "", "", fmt.Errorf("parse bundle plan envelope: %w", err)
	}
	planPath := filepath.Join(options.workdir, "plan-envelope.json")
	if err := os.WriteFile(planPath, b.PlanEnvelope, 0o600); err != nil {
		return "", "", "", fmt.Errorf("write plan envelope: %w", err)
	}
	return planPath, filepath.Join(options.workdir, "state"), extractDir, nil
}

func newCustomerManagedLogger(logLevel string) (*zap.Logger, error) {
	logCfg := zap.NewProductionConfig()
	if level, parseErr := zap.ParseAtomicLevel(logLevel); parseErr == nil {
		logCfg.Level = level
	}
	return logCfg.Build()
}

type customerManagedOperationDeps struct {
	syncer *customerManagedS3Sync
	owner  string
}

func (c *cli) executeCustomerManaged(ctx context.Context, options *customerManagedOptions, stateDir, bundleDir string, logger *zap.Logger, onBootstrapDone func(context.Context) error, operationDeps *customerManagedOperationDeps) error {
	rawEnvelope, err := os.ReadFile(options.plan)
	if err != nil {
		return fmt.Errorf("read plan envelope: %w", err)
	}
	envelope, err := customermanaged.Parse(rawEnvelope)
	if err != nil {
		return err
	}
	frozenInstallID := envelope.InstallID
	if options.deploymentID != "" {
		if _, err := customermanaged.ApplyDeploymentID(envelope, options.deploymentID); err != nil {
			return err
		}
	}
	store, err := statestore.NewDisk(stateDir)
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
	var bundleSource *customermanaged.BundleSource
	var bundleInfo *operation.BundleInfo
	bundleDigest := operation.EnvelopeDigest(rawEnvelope)
	if b, openErr := bundle.Open(ctx, bundleDir); openErr == nil {
		bundleInfo = operationrun.BundleInfoFromManifest(envelope.InstallID, bundleDigest, b.Manifest, time.Now())
		bundleInfo.ArchiveDigest = options.archiveDigest
		source, err := customermanaged.NewBundleSource(b.Store(), b.Members(), b.Provenance)
		if err != nil {
			return err
		}
		bundleSource = source
		if missing := source.MissingPlanSources(envelope); len(missing) > 0 {
			return fmt.Errorf("bundle does not package these plan sources: %s", strings.Join(missing, "; "))
		}
		if err := source.AddPlanAliases(envelope); err != nil {
			return fmt.Errorf("resolve bundle plan sources: %w", err)
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
	backend, err := customermanaged.NewTFBackend(store, filepath.Join(options.state, "tfbackend-port"))
	if err != nil {
		return err
	}
	defer backend.Close()
	clientOptions := customermanaged.ClientOptions{BundleDigest: bundleDigest}
	var activeBundle operation.BundleInfo
	readLocalOperation := func(key string) ([]byte, bool, error) {
		raw, found, readErr := store.ReadFile(key)
		if readErr != nil || found {
			return raw, found, readErr
		}
		legacyKey, ok := operation.LegacyKey(key)
		if !ok {
			return raw, false, nil
		}
		return store.ReadFile(legacyKey)
	}
	if raw, found, readErr := readLocalOperation(operation.BundleKey); readErr != nil {
		return readErr
	} else if found {
		if err := json.Unmarshal(raw, &activeBundle); err != nil {
			return fmt.Errorf("parse active bundle inventory: %w", err)
		}
	}
	var candidate *operation.BundleCandidate
	readCandidate := store.ReadFile
	if operationDeps != nil {
		readCandidate = func(key string) ([]byte, bool, error) {
			return operationDeps.syncer.readControlObject(ctx, key)
		}
	}
	readOperationCandidate := func(key string) ([]byte, bool, error) {
		raw, found, readErr := readCandidate(key)
		if readErr != nil || found {
			return raw, found, readErr
		}
		legacyKey, ok := operation.LegacyKey(key)
		if !ok {
			return raw, false, nil
		}
		return readCandidate(legacyKey)
	}
	if raw, found, readErr := readOperationCandidate(operation.CandidateKey); readErr != nil {
		return readErr
	} else if found {
		var parsed operation.BundleCandidate
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return fmt.Errorf("parse bundle candidate: %w", err)
		}
		candidate = &parsed
		if candidate.Bundle.BundleDigest == bundleDigest && activeBundle.BundleDigest != bundleDigest {
			if candidate.PreviousDigest != activeBundle.BundleDigest {
				return fmt.Errorf("bundle candidate was staged against %q but active bundle is %q", candidate.PreviousDigest, activeBundle.BundleDigest)
			}
			if candidate.Bundle.DeploymentID != envelope.InstallID {
				return fmt.Errorf("bundle candidate targets deployment %q but runner is %q", candidate.Bundle.DeploymentID, envelope.InstallID)
			}
			clientOptions.BundleUpgrade = true
			clientOptions.ApprovalPhase = "components"
			for _, change := range candidate.Changes {
				if change.Kind == operation.BundleContentKindSandbox && change.Change != operation.BundleChangeUnchanged {
					if change.Change != operation.BundleChangeChanged || change.Name != "terraform" {
						return fmt.Errorf("bundle candidate sandbox transition %q for %q is not supported", change.Change, change.Name)
					}
					clientOptions.UpgradeSandbox = true
					clientOptions.ApprovalPhase = "sandbox"
				}
				if change.Kind == operation.BundleContentKindComponent {
					switch change.Change {
					case operation.BundleChangeAdded, operation.BundleChangeChanged:
						clientOptions.UpgradeComponents = append(clientOptions.UpgradeComponents, change.Name)
					case operation.BundleChangeRemoved:
						return fmt.Errorf("bundle candidate removes component %q; component removal is not supported", change.Name)
					}
				}
			}
			if status, err := store.ReadStatus(); err != nil {
				return err
			} else if status != nil && status.BundleDigest == bundleDigest && status.ApprovalPhase != "" {
				clientOptions.ApprovalPhase = status.ApprovalPhase
			}
			approvalKey := operation.CandidateApprovalKey(bundleDigest)
			if clientOptions.ApprovalPhase == "sandbox" {
				approvalKey = operation.CandidateSandboxApprovalKey(bundleDigest)
			}
			_, approved, err := readOperationCandidate(approvalKey)
			if err != nil {
				return err
			}
			clientOptions.RequireApproval = !approved
		}
	}
	client, err := customermanaged.NewClientWithOptions(envelope, store, logger, clientOptions)
	if err != nil {
		return err
	}
	if bundleInfo != nil {
		bundleInfo.OperationID = client.RunID()
	}
	if clientOptions.BundleUpgrade && operationDeps != nil {
		go waitForCandidateApproval(ctx, operationDeps.syncer, bundleDigest, client, logger)
	}
	go waitForInstallControls(ctx, store, operationDeps, client, logger)
	installStackOutputs := map[string]any{}
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
		if err := json.Unmarshal(raw, &installStackOutputs); err != nil {
			return fmt.Errorf("parse install stack outputs %s: %w", options.installStackOutputs, err)
		}
		client.SetInstallStackOutputs(installStackOutputs)
		logger.Info("install stack outputs will be rebound into step plans", zap.String("path", options.installStackOutputs), zap.Int("keys", len(installStackOutputs)))
	}
	installInputValues := map[string]string{}
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
		if err := json.Unmarshal(raw, &installInputValues); err != nil {
			return fmt.Errorf("parse install inputs %s: %w", options.installInputs, err)
		}
		if err := customermanaged.ValidateInputValues(envelope.Inputs, installInputValues); err != nil {
			return err
		}
		client.SetInstallInputs(installInputValues)
		logger.Info("install inputs will be substituted into step plans", zap.String("path", options.installInputs), zap.Int("keys", len(installInputValues)))
	}
	metadataObservedAt := time.Now().UTC()
	inputsRaw, err := json.Marshal(customermanaged.CaptureInputs(envelope.Inputs, installInputValues, metadataObservedAt))
	if err != nil {
		return fmt.Errorf("encode captured install inputs: %w", err)
	}
	if err := store.WriteFile(customermanaged.CapturedInputsKey, inputsRaw); err != nil {
		return fmt.Errorf("persist captured install inputs: %w", err)
	}
	rolesRaw, err := json.Marshal(customermanaged.CaptureRoles(installStackOutputs, metadataObservedAt))
	if err != nil {
		return fmt.Errorf("encode captured install roles: %w", err)
	}
	if err := store.WriteFile(customermanaged.CapturedRolesKey, rolesRaw); err != nil {
		return fmt.Errorf("persist captured install roles: %w", err)
	}
	cfg := &runnerconfig.Config{
		GitRef: "customer-managed", RunnerAPIURL: "http://" + backend.Addr(), RunnerAPIToken: "customer-managed",
		RunnerID: "customer-managed-" + envelope.InstallID, HostIP: "127.0.0.1", LogLevel: options.logLevel,
		OTLPLogsEndpoint: telemetryexport.CustomerManagedOTLPLogsEndpoint,
		OTELLogDir:       filepath.Join(stateDir, statestore.JobLogsPrefix),
		BundleDir:        bundleDir, RegistryDir: options.state + "/registry", RegistryPort: 5001,
		HealthPort: 9999, SandboxJobDuration: 5 * time.Second, SandboxControlPort: 9095,
	}
	runtimeSettings := &settings.Settings{
		HeartBeatTimeout: 30 * time.Second, JobLoopMinPollPeriod: time.Second, LongPollJobs: false,
		EnableLogging: true, LoggingLevel: slog.LevelInfo, EnableMetrics: false,
		Groups: []string{"sandbox", "deploy", "sync", "actions"}, Metadata: map[string]string{"install.id": envelope.InstallID, "runner.id": cfg.RunnerID},
		OTELConfiguration: "{}", OtelSchemaURL: cfg.RunnerAPIURL, Platform: "customer-managed", Cfg: cfg,
	}
	drainer := drain.New()
	heartbeatCtx, stopHeartbeat := context.WithCancel(ctx)
	heartbeatDone := make(chan struct{})
	go func() {
		defer close(heartbeatDone)
		runCustomerManagedHeartbeat(heartbeatCtx, customermanaged.RunnerHeartbeat{
			RunnerID: cfg.RunnerID, SessionID: uuid.NewString(), Version: version.Version,
			BundleDigest: operation.EnvelopeDigest(rawEnvelope), Capabilities: []string{customermanaged.RunnerCapabilityCandidateArtifactPlans}, StartedAt: time.Now().UTC(),
		}, time.Minute, func(writeCtx context.Context, raw []byte) error {
			if err := store.WriteFile(customermanaged.RunnerHeartbeatKey, raw); err != nil {
				return err
			}
			if operationDeps != nil {
				return operationDeps.syncer.writeRunnerHeartbeat(writeCtx, raw)
			}
			return nil
		}, logger)
	}()
	defer func() {
		stopHeartbeat()
		<-heartbeatDone
	}()

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
	providers = append(providers, telemetryexport.CustomerManagedModule)

	app := fx.New(providers...)
	startCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := app.Start(startCtx); err != nil {
		return fmt.Errorf("start customer-managed runner: %w", err)
	}
	select {
	case <-ctx.Done():
	case <-client.Done():
		result := client.Result()
		logger.Info("customer-managed bootstrap complete", zap.Bool("succeeded", result.Succeeded), zap.String("failed_step", result.FailedStep))
		var publishErr error
		if result.Succeeded && onBootstrapDone != nil {
			if publishErr = onBootstrapDone(ctx); publishErr != nil {
				logger.Error("publish bootstrap completion", zap.Error(publishErr))
			}
		}
		if result.Succeeded {
			if bundleInfo != nil {
				completed, statusErr := store.ReadStatus()
				if statusErr != nil {
					return fmt.Errorf("read completed customer-managed status for bundle activation: %w", statusErr)
				}
				if completed == nil || completed.FinishedAt == nil {
					return fmt.Errorf("completed customer-managed status has no finish time for bundle activation")
				}
				bundleInfo.ActivatedAt = completed.FinishedAt.UTC()
			}
			if operationDeps != nil {
				loader := &s3CandidateBundleLoader{syncer: operationDeps.syncer, deploymentID: options.deploymentID}
				executor := operationrun.NewExecutorWithCandidateLoader(client, envelope, store, bundleSource, loader)
				controller, controllerErr := operationrun.NewController(operationrun.ControllerConfig{
					Mailbox:  operationrun.NewNamespacedMailbox(operationDeps.syncer.client, operationDeps.syncer.bucket, operationDeps.syncer.prefix, logger),
					Envelope: envelope, Digest: operation.EnvelopeDigest(rawEnvelope), DeploymentID: envelope.InstallID,
					Owner: operationDeps.owner, Executor: executor, Logger: logger, WriteLocal: store.WriteFile,
					Bundle: bundleInfo,
					FlushRun: func(flushCtx context.Context, runID string) error {
						return operationDeps.syncer.uploadSubdir(flushCtx, filepath.Join(stateDir, "runs", runID), operation.RunsPrefix+runID)
					},
				})
				if controllerErr != nil {
					logger.Error("configure operation dispatch", zap.Error(controllerErr))
				} else {
					go func() {
						if err := controller.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
							logger.Error("operation dispatch stopped", zap.Error(err))
						}
					}()
				}
			} else {
				logger.Info("operation dispatch requires --state-s3")
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
		return fmt.Errorf("stop customer-managed runner: %w", err)
	}
	result := client.Result()
	if !result.Succeeded {
		return fmt.Errorf("customer-managed run failed at step %q", result.FailedStep)
	}
	return nil
}
