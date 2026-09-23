package workflow

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/launcher"
	pkgplantypes "github.com/nuonco/nuon/bins/runner/internal/pkg/plantypes"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/runner/workspace"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestImageActionRunsWithoutAShellAsNonRootAndCollectsOutputs(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker is required")
	}
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("Docker daemon is required")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	dir := t.TempDir()
	source := `package main
import ("encoding/json"; "os"; "time")
func main() {
 if os.Getenv("MODE") == "fail" { os.Exit(7) }
 if os.Getenv("MODE") == "wait" { time.Sleep(time.Hour) }
 wd, _ := os.Getwd()
 out, err := os.OpenFile(os.Getenv("NUON_ACTIONS_OUTPUT_FILEPATH"), os.O_APPEND|os.O_WRONLY, 0)
 if err != nil { panic(err) }
 defer out.Close()
 if err := json.NewEncoder(out).Encode(map[string]any{"uid": os.Getuid(), "cwd": wd, "value": os.Getenv("VALUE"), "args": os.Args[1:]}); err != nil { panic(err) }
 os.Stdout.WriteString("fixture stdout\n")
 os.Stderr.WriteString("fixture stderr\n")
}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte(source), 0o600))
	build := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(dir, "fixture"), filepath.Join(dir, "main.go"))
	build.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=linux")
	output, err := build.CombinedOutput()
	require.NoError(t, err, string(output))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM scratch\nCOPY fixture /fixture\nUSER 65532:65532\nENTRYPOINT [\"/does-not-exist\"]\n"), 0o600))
	image := fmt.Sprintf("nuon-shell-free-test:%d", time.Now().UnixNano())
	output, err = exec.CommandContext(ctx, "docker", "build", "-t", image, dir).CombinedOutput()
	require.NoError(t, err, string(output))
	t.Cleanup(func() { _ = exec.Command("docker", "image", "rm", image).Run() })
	wk, err := workspace.New(validator.New(), workspace.WithLogger(zap.NewNop()), workspace.WithWorkspaceID("shell-free-test"), workspace.WithTmpRoot(t.TempDir()), workspace.WithGitSource(&plantypes.GitSource{URL: "https://github.com/jonmorehouse/empty", Ref: "main", Path: "."}))
	require.NoError(t, err)
	require.NoError(t, wk.Init(ctx))
	require.NoError(t, os.Chmod(wk.Root(), 0o777))
	cfg := &models.AppActionWorkflowStepConfig{Name: "check", Command: `/fixture 'two words' ""`}
	h := &handler{launcher: launcher.NewDockerLauncher(nil), state: &handlerState{
		workspace: wk, auth: &pkgplantypes.PlanAuth{},
		run:         &models.AppInstallActionWorkflowRun{ID: "test-run"},
		plan:        &plantypes.ActionWorkflowRunPlan{SourceImage: image, ImageDigestRef: image},
		workflowCfg: &models.AppActionWorkflowConfig{Steps: []*models.AppActionWorkflowStepConfig{cfg}},
	}}
	require.NoError(t, h.createExecEnv(ctx, zap.NewNop(), nil, cfg))
	core, logs := observer.New(zap.InfoLevel)
	require.NoError(t, h.execCommandInContainer(ctx, zap.New(core), cfg, nil, map[string]string{"VALUE": "line one\nline two"}))
	require.Equal(t, 1, logs.FilterMessage("fixture stdout").Len())
	require.Equal(t, 1, logs.FilterMessage("fixture stderr").Len())
	outputs, err := h.parseOutputs(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 65532, outputs["uid"])
	require.Equal(t, containerWorkspaceMount, outputs["cwd"])
	require.Equal(t, "line one\nline two", outputs["value"])
	require.Equal(t, []any{"two words", ""}, outputs["args"])
	require.Error(t, h.execCommandInContainer(ctx, zap.NewNop(), cfg, nil, map[string]string{"MODE": "fail"}))
	timeoutCtx, stop := context.WithTimeout(ctx, time.Second)
	defer stop()
	require.Error(t, h.execCommandInContainer(timeoutCtx, zap.NewNop(), cfg, nil, map[string]string{"MODE": "wait"}))
	remaining, err := exec.CommandContext(ctx, "docker", "ps", "-aq", "--filter", "label=nuon.run_id=test-run").Output()
	require.NoError(t, err)
	require.Empty(t, bytes.TrimSpace(remaining))
}
