package helm

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"helm.sh/helm/v4/pkg/action"
	release "helm.sh/helm/v4/pkg/release/v1"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/pkg/helm"
)

func (h *handler) execUninstall(
	ctx context.Context,
	l *zap.Logger,
	actionCfg *action.Configuration,
	job *models.AppRunnerJob,
	jobExecution *models.AppRunnerJobExecution,
) error {
	if err := h.uninstall(ctx, l, actionCfg); err != nil {
		h.writeErrorResult(ctx, "uninstall", err)
		return fmt.Errorf("unable to uninstall helm chart: %w", err)
	}

	res := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success: true,
	}
	if _, err := h.apiClient.CreateJobExecutionResult(
		ctx,
		job.ID,
		jobExecution.ID,
		res,
	); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}

// HelmContentDiff represents a difference in the content of a Helm chart resource.
// It is used to capture changes between two states of a resource.
// Fields:
// - ApiPath: The API path of the resource (e.g., "apps/v1").
// - Name: The name of the resource.
// - Namespace: The namespace in which the resource resides.
// - Kind: The kind of the resource (e.g., "Deployment", "Service").
// - Before: The state of the resource before the change, typically serialized as a string.
// - After: The state of the resource after the change, typically serialized as a string.
type HelmContentDiff struct {
	ApiPath   string `json:"api,omitempty"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Kind      string `json:"kind,omitempty"`
	Before    string `json:"before"`
	After     string `json:"after"`
}

func (hc *HelmContentDiff) parseRawName(s string) {
	// Parse format: "namespace, name, kind (apiBase)"
	if s == "" {
		return
	}

	// Use regex to parse the format: "namespace, name, kind (apiBase)"
	re := regexp.MustCompile(`^([^,]+),\s*([^,]+),\s*([^(]+)\s*\(([^)]+)\)$`)
	matches := re.FindStringSubmatch(s)

	if len(matches) == 5 {
		hc.Namespace = strings.TrimSpace(matches[1])
		hc.Name = strings.TrimSpace(matches[2])
		hc.Kind = strings.TrimSpace(matches[3])
		hc.ApiPath = strings.TrimSpace(matches[4])
	}
}

// HelmPlanContents is essentially a light wrapper around an Op
type HelmPlanContents struct {
	Diff        string            `json:"plan"`
	Op          string            `json:"op"`
	ContentDiff []HelmContentDiff `json:"helm_content_diff"`
}

func (h *handler) extractApplyPlanContents(contents string) (HelmPlanContents, error) {
	// base64 decode
	decodedBytes, err := base64.StdEncoding.DecodeString(contents)
	if err != nil {
		return HelmPlanContents{}, errors.Wrap(err, "unable to base64 decode contents")
	}

	// decompress
	contentsBuffer := bytes.NewReader([]byte(decodedBytes))
	reader, err := gzip.NewReader(contentsBuffer)
	if err != nil {
		return HelmPlanContents{}, errors.Wrap(err, "unable to read contents into gzip reader")
	}
	defer reader.Close()

	decompressedBytes, err := io.ReadAll(reader)
	if err != nil {
		return HelmPlanContents{}, errors.Wrap(err, "unable to decompress contents")
	}

	var helmPlan HelmPlanContents
	json.Unmarshal(decompressedBytes, &helmPlan)

	return helmPlan, nil
}

// NOTE: the helm plans are not real plans, they are just diffs
func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("Initializing Helm...",
		zapcore.Field{Key: "base_path", Type: zapcore.StringType, String: h.state.arch.BasePath()},
	)
	actionCfg, kubeCfg, err := h.actionInit(ctx, l)
	if err != nil {
		return fmt.Errorf("unable to initialize helm actions: %w", err)
	}

	// set the release storage backend dynamically
	releaseStore, err := h.getHelmReleaseStore(ctx, kubeCfg)
	if err != nil {
		return errors.Wrap(err, "unable to get release store")
	}

	actionCfg.Releases = releaseStore

	l.Info("Checking for previous Helm release...",
		zapcore.Field{Key: "base_path", Type: zapcore.StringType, String: h.state.arch.BasePath()},
	)
	prevRel, err := helm.GetRelease(actionCfg, h.state.plan.HelmDeployPlan.Name)
	if err != nil {
		return fmt.Errorf("unable to get previous helm release: %w", err)
	}

	var (
		rel      *release.Release
		op       string
		diff     string
		helmPlan HelmPlanContents
	)

	// load helm plan from the plan
	if len(h.state.plan.ApplyPlanContents) > 0 {
		// TODO: use the actual struct and move into a shared pk
		l.Debug("extracting apply plan contents", zap.Int("contents.compressed.length", len(h.state.plan.ApplyPlanContents)))
		helmPlan, err = h.extractApplyPlanContents(h.state.plan.ApplyPlanContents)
		if err != nil {
			return errors.Wrap(err, "unable to decompress and/or marshal apply plan contents into HelmPlanContents")
		}
		l.Debug("extracting apply plan contents", zap.String("plan.op", helmPlan.Op))
	}

	switch job.Operation {
	case models.AppRunnerJobOperationTypeCreateDashApplyDashPlan:
		var contentDiff *[]HelmContentDiff
		var err error
		// in this case, the diff is generated so it is available to the createAPIResult method
		if prevRel == nil {
			diff, contentDiff, err = h.installDiff(ctx, l, actionCfg, kubeCfg)
			helmPlan.Op = "install"
		} else {
			diff, contentDiff, err = h.upgrade_diff(ctx, l, actionCfg, kubeCfg)
			helmPlan.Op = "upgrade"
		}
		if err != nil {
			return err
		}

		if diff == "" {
			diff = "no changes"
		}

		helmPlan.Diff = diff
		helmPlan.ContentDiff = *contentDiff

		l.Debug("calculated helm diff", zap.String("diff", diff))
	case models.AppRunnerJobOperationTypeCreateDashTeardownDashPlan:
		// TODO(fd): figure out the best way to get a plan for this
		l.Info("executing helm uninstall plan")

		diff, contentDiff, err := h.uninstallDiff(ctx, l, actionCfg, prevRel)
		if err != nil {
			return err
		}

		helmPlan.Op = "uninstall"
		helmPlan.Diff = diff
		helmPlan.ContentDiff = *contentDiff
	case models.AppRunnerJobOperationTypeApplyDashPlan:
		l.Info(fmt.Sprintf("executing helm %s", helmPlan.Op))
		switch helmPlan.Op {
		case "install":
			op = "install"
			rel, err = h.install(ctx, l, actionCfg, kubeCfg)
		case "upgrade":
			op = "upgrade"
			rel, err = h.upgrade(ctx, l, actionCfg, kubeCfg)
		case "uninstall":
			op = "uninstall"
			err = h.execUninstall(ctx, l, actionCfg, job, jobExecution)
		default:
			l.Error("plan did not define an Op. this is unexpected.")
		}
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported run type %s", job.Operation)
	}

	// handle error
	if err != nil {
		h.writeErrorResult(ctx, op, err)
		return fmt.Errorf("unable to %s helm chart: %w", op, err)
	}

	var apiRes *models.ServiceCreateRunnerJobExecutionResultRequest
	var planContents HelmPlanContents

	// save plan if its not apply job operation is not apply
	if job.Operation != models.AppRunnerJobOperationTypeApplyDashPlan {
		planContents = helmPlan
	}

	apiRes, err = h.createAPIResultRequest(l, rel, planContents)
	if err != nil {
		h.writeErrorResult(ctx, op, err)
		return fmt.Errorf("unable to create api result from release: %w", err)
	}

	_, err = h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, apiRes)
	if err != nil {
		l.Error("failed to create job executione result")
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}
