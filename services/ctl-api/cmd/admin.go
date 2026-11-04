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
	adminCmd.AddCommand(&cobra.Command{
		Use:   "post-promotion",
		Short: "enqueue the promotion signal for the version in config",
		Args:  cobra.NoArgs,
		Run: func(*cobra.Command, []string) {
			c.runAdmin("post-promotion", runPostPromotion)
		},
	})
	rootCmd.AddCommand(adminCmd)
}

func runPostPromotion(ctx context.Context, deps *adminDeps) (any, error) {
	return deps.Ops.Promote(ctx, deps.Cfg.Version)
}

func (c *cli) runAdmin(command string, op adminOperation) {
	start := time.Now()

	res := adminResult{Command: command, Success: true}
	out, err := c.execAdmin(op)
	if err != nil {
		zap.L().Error("admin command failed", zap.String("command", command), zap.Error(err))
		res.Success = false
		res.Error = err.Error()
	} else {
		res.Result = out
	}

	ensureMinRuntime(start)

	os.Exit(writeAdminResult(res))
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
