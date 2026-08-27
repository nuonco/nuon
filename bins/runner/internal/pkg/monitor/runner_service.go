package monitor

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/fidiego/systemctl"
	"github.com/fidiego/systemctl/properties"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	RunnerServiceDir  = "/etc/systemd/system"
	RunnerServiceName = "nuon-runner.service"

	runnerServiceRestartPendingFilename = "/opt/nuon/runner/runner-service-restart-pending"
	runnerServiceSystemctlTimeout       = time.Minute
	runnerServiceRestartRetryInterval   = 2 * time.Minute
)

var defaultSystemctlOpts = systemctl.Options{UserMode: false}

type runnerServiceRestartState struct {
	PreviousInvocationID string    `json:"previous_invocation_id"`
	RestartRequestedAt   time.Time `json:"restart_requested_at"`
}

type runnerServiceState struct {
	ActiveState  string
	SubState     string
	InvocationID string
}

type runnerServiceRestartAction int

const (
	runnerServiceRestartWait runnerServiceRestartAction = iota
	runnerServiceRestartRequest
	runnerServiceRestartComplete
)

//go:embed templates/runner-service.aws.service
var runnerServiceAWS string

//go:embed templates/runner-service.gcp.service
var runnerServiceGCP string

//go:embed templates/runner-service.azure.service
var runnerServiceAzure string

func (h *Monitor) ensureRunnerSystemdService(ctx context.Context) error {
	restartState, err := h.ensureRunnerServiceDefinition()
	if err != nil {
		return err
	}
	if restartState != nil {
		return h.reconcileRunnerServiceRestart(ctx, restartState)
	}
	return h.ensureRunnerServiceIsActive(ctx)
}

func (h *Monitor) ensureRunnerServiceDefinition() (*runnerServiceRestartState, error) {
	path := filepath.Join(RunnerServiceDir, RunnerServiceName)
	h.l.Debug(fmt.Sprintf("ensuring runner unit file is current: %s", path))

	var serviceTemplate string
	switch h.settings.Platform {
	case "aws", "":
		serviceTemplate = runnerServiceAWS
	case "gcp":
		serviceTemplate = runnerServiceGCP
	case "azure":
		serviceTemplate = runnerServiceAzure
	default:
		serviceTemplate = runnerServiceAWS
	}
	tmpl := template.Must(template.New("").Parse(serviceTemplate))
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, h.settings); err != nil {
		return nil, errors.Wrap(err, "unable to render runner service definition")
	}

	definitionCurrent, err := runnerServiceDefinitionCurrent(path, rendered.Bytes())
	if err != nil {
		return nil, err
	}
	if !definitionCurrent {
		restartState := runnerServiceRestartState{}
		if err := writeNuonRunnerService(path, runnerServiceRestartPendingFilename, rendered.Bytes(), restartState); err != nil {
			return nil, err
		}
		h.l.Info("runner service definition changed", zap.String("path", path))
	}

	restartState, err := readRunnerServiceRestartState(runnerServiceRestartPendingFilename)
	if err != nil {
		return nil, err
	}
	if restartState == nil {
		h.l.Debug(fmt.Sprintf("%s - everything is in order", RunnerServiceName))
	}
	return restartState, nil
}

func runnerServiceDefinitionCurrent(path string, contents []byte) (bool, error) {
	current, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("unable to read %s", path))
	}
	info, err := os.Stat(path)
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("unable to stat %s", path))
	}
	return bytes.Equal(current, contents) && info.Mode().Perm() == 0o644, nil
}

func writeNuonRunnerService(path string, restartPendingPath string, contents []byte, restartState runnerServiceRestartState) error {
	if err := writeRunnerServiceRestartState(restartPendingPath, restartState); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to replace %s", path))
	}
	if err := f.Chmod(0o644); err != nil {
		f.Close()
		return errors.Wrap(err, fmt.Sprintf("unable to set permissions on %s", path))
	}
	if err := f.Truncate(0); err != nil {
		f.Close()
		return errors.Wrap(err, fmt.Sprintf("unable to truncate %s", path))
	}
	if _, err := f.Write(contents); err != nil {
		f.Close()
		return errors.Wrap(err, fmt.Sprintf("unable to write %s", path))
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return errors.Wrap(err, fmt.Sprintf("unable to sync %s", path))
	}
	if err := f.Close(); err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to close %s", path))
	}
	return nil
}

func writeRunnerServiceRestartState(path string, state runnerServiceRestartState) error {
	contents, err := json.Marshal(state)
	if err != nil {
		return errors.Wrap(err, "unable to marshal pending runner service restart")
	}
	if err := writeFileAtomically(path, contents, 0o600); err != nil {
		return errors.Wrap(err, "unable to write pending runner service restart")
	}
	return nil
}

func readRunnerServiceRestartState(path string) (*runnerServiceRestartState, error) {
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to read pending runner service restart")
	}
	var state runnerServiceRestartState
	if err := json.Unmarshal(contents, &state); err != nil {
		return nil, errors.Wrap(err, "unable to parse pending runner service restart")
	}
	return &state, nil
}

func writeFileAtomically(path string, contents []byte, mode os.FileMode) error {
	f, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*")
	if err != nil {
		return errors.Wrap(err, "unable to create temporary file")
	}
	tempPath := f.Name()
	defer os.Remove(tempPath)

	if err := f.Chmod(mode); err != nil {
		f.Close()
		return errors.Wrap(err, "unable to set temporary file permissions")
	}
	if _, err := f.Write(contents); err != nil {
		f.Close()
		return errors.Wrap(err, "unable to write temporary file")
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return errors.Wrap(err, "unable to sync temporary file")
	}
	if err := f.Close(); err != nil {
		return errors.Wrap(err, "unable to close temporary file")
	}
	if err := os.Rename(tempPath, path); err != nil {
		return errors.Wrap(err, "unable to replace file")
	}
	return nil
}

func (h *Monitor) reconcileRunnerServiceRestart(ctx context.Context, restartState *runnerServiceRestartState) error {
	statusCtx, cancel := context.WithTimeout(ctx, runnerServiceSystemctlTimeout)
	state, err := getRunnerServiceState(statusCtx)
	cancel()
	if err != nil {
		return err
	}

	action := getRunnerServiceRestartAction(*restartState, state, time.Now())
	switch action {
	case runnerServiceRestartComplete:
		if err := os.Remove(runnerServiceRestartPendingFilename); err != nil && !errors.Is(err, os.ErrNotExist) {
			return errors.Wrap(err, "unable to clear pending runner service restart")
		}
		h.l.Info("runner service definition update activated",
			zap.String("service", RunnerServiceName),
			zap.String("invocation_id", state.InvocationID))
	case runnerServiceRestartRequest:
		if err := h.requestRunnerServiceRestart(ctx, restartState); err != nil {
			return err
		}
	case runnerServiceRestartWait:
		h.l.Debug("waiting for runner service definition update to activate",
			zap.String("service", RunnerServiceName),
			zap.String("active_state", state.ActiveState),
			zap.String("sub_state", state.SubState))
	}
	return nil
}

func (h *Monitor) requestRunnerServiceRestart(ctx context.Context, restartState *runnerServiceRestartState) error {
	reloadCtx, cancelReload := context.WithTimeout(ctx, runnerServiceSystemctlTimeout)
	err := systemctl.DaemonReload(reloadCtx, defaultSystemctlOpts)
	cancelReload()
	if err != nil {
		return errors.Wrap(err, "unable to reload the systemd daemon")
	}

	statusCtx, cancelStatus := context.WithTimeout(ctx, runnerServiceSystemctlTimeout)
	previousInvocationID, err := runnerServiceInvocationID(statusCtx)
	cancelStatus()
	if err != nil {
		return err
	}

	restartState.PreviousInvocationID = previousInvocationID
	restartState.RestartRequestedAt = time.Now().UTC()
	if err := writeRunnerServiceRestartState(runnerServiceRestartPendingFilename, *restartState); err != nil {
		return err
	}

	restartCtx, cancelRestart := context.WithTimeout(ctx, runnerServiceSystemctlTimeout)
	err = requestRunnerServiceOperation(restartCtx, "restart")
	cancelRestart()
	if err != nil {
		return err
	}

	h.l.Info("runner service restart requested",
		zap.String("service", RunnerServiceName))
	return nil
}

func getRunnerServiceRestartAction(restartState runnerServiceRestartState, state runnerServiceState, now time.Time) runnerServiceRestartAction {
	if !restartState.RestartRequestedAt.IsZero() && state.ActiveState == "active" && state.InvocationID != "" && state.InvocationID != restartState.PreviousInvocationID {
		return runnerServiceRestartComplete
	}
	if state.ActiveState == "activating" || state.ActiveState == "deactivating" || state.ActiveState == "reloading" {
		return runnerServiceRestartWait
	}
	if restartState.RestartRequestedAt.IsZero() {
		return runnerServiceRestartRequest
	}
	if now.Sub(restartState.RestartRequestedAt) >= runnerServiceRestartRetryInterval {
		return runnerServiceRestartRequest
	}
	return runnerServiceRestartWait
}

func runnerServiceInvocationID(ctx context.Context) (string, error) {
	invocationID, err := systemctl.Show(ctx, RunnerServiceName, properties.InvocationID, defaultSystemctlOpts)
	if errors.Is(err, systemctl.ErrDoesNotExist) {
		return "", nil
	}
	if err != nil {
		return "", errors.Wrap(err, "unable to determine runner service invocation ID")
	}
	return invocationID, nil
}

func getRunnerServiceState(ctx context.Context) (runnerServiceState, error) {
	activeState, err := systemctl.Show(ctx, RunnerServiceName, properties.ActiveState, defaultSystemctlOpts)
	if errors.Is(err, systemctl.ErrDoesNotExist) {
		return runnerServiceState{ActiveState: "inactive"}, nil
	}
	if err != nil {
		return runnerServiceState{}, errors.Wrap(err, "unable to determine runner service active state")
	}
	subState, err := systemctl.Show(ctx, RunnerServiceName, properties.SubState, defaultSystemctlOpts)
	if err != nil {
		return runnerServiceState{}, errors.Wrap(err, "unable to determine runner service sub-state")
	}
	invocationID, err := runnerServiceInvocationID(ctx)
	if err != nil {
		return runnerServiceState{}, err
	}
	return runnerServiceState{
		ActiveState:  activeState,
		SubState:     subState,
		InvocationID: invocationID,
	}, nil
}

func requestRunnerServiceOperation(ctx context.Context, operation string) error {
	output, err := exec.CommandContext(ctx, "systemctl", "--system", "--no-block", operation, RunnerServiceName).CombinedOutput()
	if err == nil {
		return nil
	}
	message := strings.TrimSpace(string(output))
	if message == "" {
		return errors.Wrapf(err, "unable to request runner service %s", operation)
	}
	return errors.Wrapf(err, "unable to request runner service %s: %s", operation, message)
}

// this method encapsulates all of the logic to ensure the service is running.
// NOTE: we use start instead of enable
func (h *Monitor) ensureRunnerServiceIsActive(ctx context.Context) error {
	h.l.Debug("ensuring runner service is active")
	statusCtx, cancelStatus := context.WithTimeout(ctx, runnerServiceSystemctlTimeout)
	state, err := getRunnerServiceState(statusCtx)
	cancelStatus()
	if err != nil {
		return errors.Wrap(err, "unable to determine if unit is active")
	}
	switch state.ActiveState {
	case "active":
		startTimeCtx, cancelStartTime := context.WithTimeout(ctx, runnerServiceSystemctlTimeout)
		startTime, err := systemctl.GetStartTime(startTimeCtx, RunnerServiceName, defaultSystemctlOpts)
		cancelStartTime()
		if err != nil {
			return errors.Wrap(err, "unable to determine start time")
		}
		h.l.Info(fmt.Sprintf("service is up and running - uptime: %s", startTime))
	case "inactive", "failed":
		startCtx, cancelStart := context.WithTimeout(ctx, runnerServiceSystemctlTimeout)
		err := requestRunnerServiceOperation(startCtx, "start")
		cancelStart()
		if err != nil {
			return err
		}
		h.l.Info("runner service start requested", zap.String("service", RunnerServiceName))
	default:
		h.l.Debug("runner service is transitioning",
			zap.String("active_state", state.ActiveState),
			zap.String("sub_state", state.SubState))
	}
	return nil
}
