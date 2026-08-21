package actions

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type AdHocParams struct {
	InstallID        string
	Command          string
	ScriptPath       string
	Env              []string
	EnvFile          string
	Timeout          time.Duration
	Name             string
	Role             string
	EnableKubeConfig bool
	Wait             bool
}

// Kafka-backed log ingestion can trail stream closure by up to five seconds.
const adHocClosedLogDrainPolls = 4

func (s *Service) CreateAdHocRun(ctx context.Context, params AdHocParams, asJSON bool) error {
	view := ui.NewGetView()

	if params.InstallID == "" {
		params.InstallID = s.cfg.GetString("install_id")
	}
	installID, err := lookup.InstallID(ctx, s.api, params.InstallID)
	if err != nil {
		return view.Error(err)
	}

	inlineContents, err := readAdHocInput(params.Command, params.ScriptPath)
	if err != nil {
		return view.Error(err)
	}

	envVars, err := parseAdHocEnv(params.EnvFile, params.Env)
	if err != nil {
		return view.Error(err)
	}

	if params.Timeout < time.Second || params.Timeout > time.Hour || params.Timeout%time.Second != 0 {
		return view.Error(fmt.Errorf("timeout must be a whole number of seconds between 1s and 1h"))
	}

	req := &models.ServiceCreateAdHocActionRequest{
		Command:          params.Command,
		InlineContents:   inlineContents,
		EnvVars:          envVars,
		Timeout:          int64(params.Timeout / time.Second),
		Name:             params.Name,
		Role:             params.Role,
		EnableKubeConfig: &params.EnableKubeConfig,
	}

	run, err := s.api.CreateAdHocAction(ctx, installID, req)
	if err != nil {
		return view.Error(err)
	}
	if params.Wait {
		logWriter := io.Writer(os.Stdout)
		if asJSON {
			logWriter = os.Stderr
		}
		completedRun, err := s.waitForAdHocRun(ctx, installID, run.ID, logWriter)
		if err != nil {
			return printAdHocWaitError(err, asJSON)
		}

		run.Status = models.AppInstallActionWorkflowRunStatus(completedRun.Status)
		run.StatusDescription = completedRun.StatusDescription
		if completedRun.Status != string(models.AppInstallActionWorkflowRunStatusFinished) {
			message := fmt.Sprintf("ad-hoc action %s ended with status %s", completedRun.ID, completedRun.Status)
			if completedRun.StatusDescription != "" {
				message += ": " + completedRun.StatusDescription
			}
			return printAdHocWaitError(&ui.CLIUserError{
				Msg: message,
			}, asJSON)
		}
		if asJSON {
			ui.PrintJSON(run)
		}
		return nil
	}

	if asJSON {
		ui.PrintJSON(run)
		return nil
	}

	ui.PrintSuccess("ad-hoc action queued")
	view.Render([][]string{
		{"run id", run.ID},
		{"workflow id", run.WorkflowID},
		{"install id", run.InstallID},
		{"status", string(run.Status)},
		{"status description", run.StatusDescription},
	})

	return nil
}

func (s *Service) waitForAdHocRun(ctx context.Context, installID, runID string, writer io.Writer) (*models.AppInstallActionWorkflowRun, error) {
	streamedLogs := false
	for {
		run, err := s.api.GetInstallActionWorkflowRun(ctx, installID, runID)
		if err != nil {
			return nil, fmt.Errorf("get ad-hoc action run: %w", err)
		}

		if !streamedLogs && run.RunnerJob != nil && run.RunnerJob.LogStreamID != "" {
			if err := s.streamRawAdHocLogs(ctx, run.RunnerJob.LogStreamID, writer); err != nil {
				return nil, err
			}
			streamedLogs = true
		}

		if isTerminalAdHocStatus(run.Status) {
			return run, nil
		}
		if err := waitForAdHocPoll(ctx); err != nil {
			return nil, err
		}
	}
}

func (s *Service) streamRawAdHocLogs(ctx context.Context, logStreamID string, writer io.Writer) error {
	cursor := ""
	closedEmptyPolls := 0
	for {
		response, err := s.api.LogStreamTailLogs(ctx, logStreamID, cursor, "2s")
		if err != nil {
			return fmt.Errorf("read ad-hoc action logs: %w", err)
		}
		for _, record := range response.Logs {
			if record.LogAttributes["nuon.command_output"] != "true" {
				continue
			}
			if _, err := io.WriteString(writer, record.Body); err != nil {
				return fmt.Errorf("write ad-hoc action logs: %w", err)
			}
			if !strings.HasSuffix(record.Body, "\n") {
				if _, err := io.WriteString(writer, "\n"); err != nil {
					return fmt.Errorf("write ad-hoc action logs: %w", err)
				}
			}
		}
		if response.Next != "" {
			cursor = response.Next
		}
		if response.HasMore {
			continue
		}

		logStream, err := s.api.GetLogStream(ctx, logStreamID)
		if err != nil {
			return fmt.Errorf("get ad-hoc action log stream: %w", err)
		}
		if logStream.Open || len(response.Logs) > 0 {
			closedEmptyPolls = 0
			continue
		}

		closedEmptyPolls++
		if closedEmptyPolls >= adHocClosedLogDrainPolls {
			return nil
		}
	}
}

func isTerminalAdHocStatus(status string) bool {
	switch models.AppInstallActionWorkflowRunStatus(status) {
	case models.AppInstallActionWorkflowRunStatusFinished,
		models.AppInstallActionWorkflowRunStatusError,
		models.AppInstallActionWorkflowRunStatusTimedDashOut,
		models.AppInstallActionWorkflowRunStatusCancelled,
		models.AppInstallActionWorkflowRunStatusRetried:
		return true
	default:
		return false
	}
}

func waitForAdHocPoll(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second):
		return nil
	}
}

func printAdHocWaitError(err error, asJSON bool) error {
	if asJSON {
		return ui.PrintError(err)
	}
	fmt.Fprintln(os.Stderr, err)
	return err
}

func readAdHocInput(command, scriptPath string) (string, error) {
	if command != "" && scriptPath != "" {
		return "", fmt.Errorf("provide either --command or --script, not both")
	}
	if command == "" && scriptPath == "" {
		return "", fmt.Errorf("either --command or --script is required")
	}
	if scriptPath == "" {
		return "", nil
	}

	contents, err := os.ReadFile(scriptPath)
	if err != nil {
		return "", fmt.Errorf("read script %q: %w", scriptPath, err)
	}
	if len(contents) == 0 {
		return "", fmt.Errorf("script %q is empty", scriptPath)
	}
	return string(contents), nil
}

func parseAdHocEnv(envFile string, values []string) (map[string]string, error) {
	env := make(map[string]string)
	if envFile != "" {
		file, err := os.Open(envFile)
		if err != nil {
			return nil, fmt.Errorf("open env file %q: %w", envFile, err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for lineNumber := 1; scanner.Scan(); lineNumber++ {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
			key, value, err := parseEnvAssignment(line)
			if err != nil {
				return nil, fmt.Errorf("parse env file %q line %d: %w", envFile, lineNumber, err)
			}
			env[key] = trimMatchingQuotes(strings.TrimSpace(value))
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("read env file %q: %w", envFile, err)
		}
	}

	for _, value := range values {
		key, envValue, err := parseEnvAssignment(value)
		if err != nil {
			return nil, fmt.Errorf("parse --env %q: %w", value, err)
		}
		env[key] = envValue
	}

	return env, nil
}

func parseEnvAssignment(value string) (string, string, error) {
	key, envValue, ok := strings.Cut(value, "=")
	key = strings.TrimSpace(key)
	if !ok || !validEnvKey(key) {
		return "", "", fmt.Errorf("expected KEY=VALUE")
	}
	return key, envValue, nil
}

func validEnvKey(key string) bool {
	if key == "" || !(key[0] == '_' || key[0] >= 'A' && key[0] <= 'Z' || key[0] >= 'a' && key[0] <= 'z') {
		return false
	}
	for i := 1; i < len(key); i++ {
		char := key[i]
		if char != '_' && !(char >= 'A' && char <= 'Z') && !(char >= 'a' && char <= 'z') && !(char >= '0' && char <= '9') {
			return false
		}
	}
	return true
}

func trimMatchingQuotes(value string) string {
	if len(value) < 2 {
		return value
	}
	if value[0] == value[len(value)-1] && (value[0] == '\'' || value[0] == '"') {
		return value[1 : len(value)-1]
	}
	return value
}
