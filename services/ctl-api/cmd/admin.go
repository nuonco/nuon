package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/adminops"
)

const (
	adminOperationTimeout = 5 * time.Minute
	adminResultMarker     = "ADMIN_RESULT"
)

type adminDeps struct {
	fx.In

	Ops *adminops.Ops
	Cfg *internal.Config
	L   *zap.Logger
}

type adminResult struct {
	Command string `json:"command"`
	Success bool   `json:"success"`
	Result  any    `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

type adminOperation func(ctx context.Context, deps *adminDeps) (any, error)

func (c *cli) registerAdmin() {
	adminCmd := &cobra.Command{
		Use:   "admin",
		Short: "run admin operations as one-off commands",
	}

	var force bool
	postPromotionCmd := &cobra.Command{
		Use:   "post-promotion",
		Short: "enqueue the promotion signal for the version in config, unless it is already the latest promotion",
		Args:  cobra.NoArgs,
		Run: func(*cobra.Command, []string) {
			c.runAdmin("post-promotion", func(ctx context.Context, deps *adminDeps) (any, error) {
				if force {
					return adminOutput(deps.Ops.Promote(ctx, deps.Cfg.Version))
				}
				return adminOutput(deps.Ops.PromoteOnce(ctx, deps.Cfg.Version))
			})
		},
	}
	postPromotionCmd.Flags().BoolVar(&force, "force", false, "enqueue a new promotion even if this version is already the latest")
	adminCmd.AddCommand(postPromotionCmd)

	runnersCmd := &cobra.Command{
		Use:   "runners",
		Short: "run admin operations on runners",
	}

	var updateOpts adminops.UpdateRunnersOptions
	runnersUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "set install runners to the runner image tag in config",
		Args:  cobra.NoArgs,
		Run: func(*cobra.Command, []string) {
			c.runAdmin("runners update", func(ctx context.Context, deps *adminDeps) (any, error) {
				return adminOutput(deps.Ops.UpdateInstallRunners(ctx, deps.Cfg.RunnerContainerImageTag, updateOpts))
			})
		},
	}
	runnersUpdateCmd.Flags().BoolVar(&updateOpts.DryRun, "dry-run", false, "report the runners that would be updated without changing them")
	runnersUpdateCmd.Flags().BoolVar(&updateOpts.RestartSignal, "restart-signal", true, "enqueue a restart signal for each updated runner")
	runnersCmd.AddCommand(runnersUpdateCmd)
	adminCmd.AddCommand(runnersCmd)

	rootCmd.AddCommand(adminCmd)
}

func adminOutput[T any](out *T, err error) (any, error) {
	if out == nil {
		return nil, err
	}
	return out, err
}

func (c *cli) runAdmin(command string, op adminOperation) {
	start := time.Now()

	out, err := c.execAdmin(op)
	res := adminResult{Command: command, Success: err == nil, Result: out}
	if err != nil {
		zap.L().Error("admin command failed", zap.String("command", command), zap.Error(err))
		res.Error = err.Error()
	}

	code := writeAdminResult(res)

	ensureMinRuntime(start)

	os.Exit(code)
}

func (c *cli) execAdmin(op adminOperation) (any, error) {
	var deps adminDeps
	app := fx.New(append(c.providers(), fx.Populate(&deps))...)
	if err := app.Err(); err != nil {
		return nil, fmt.Errorf("unable to build app: %w", err)
	}

	startCtx, cancelStart := context.WithTimeout(context.Background(), app.StartTimeout())
	defer cancelStart()
	if err := app.Start(startCtx); err != nil {
		return nil, fmt.Errorf("unable to start app: %w", err)
	}

	defer func() {
		stopCtx, cancelStop := context.WithTimeout(context.Background(), app.StopTimeout())
		defer cancelStop()
		if err := app.Stop(stopCtx); err != nil {
			deps.L.Error("unable to stop app", zap.Error(err))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), adminOperationTimeout)
	defer cancel()

	ctx, err := deps.Ops.WithActor(ctx)
	if err != nil {
		return nil, err
	}

	return op(ctx, &deps)
}

func writeAdminResult(res adminResult) int {
	code := 0
	if !res.Success {
		code = 1
	}

	body, err := json.Marshal(res)
	if err != nil {
		code = 1
		body, _ = json.Marshal(adminResult{
			Command: res.Command,
			Error:   fmt.Sprintf("unable to encode result: %s", err),
		})
	}
	fmt.Fprintf(os.Stdout, "%s %s\n", adminResultMarker, body)

	return code
}
