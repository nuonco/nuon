package telemetryexport

import (
	"context"
	"slices"
	"testing"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/audit"
	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	"github.com/nuonco/nuon/pkg/runner/settings"
)

type recordingLifecycle struct {
	fx.Lifecycle
	order *[]string
}

func (l recordingLifecycle) Append(hook fx.Hook) {
	if stop := hook.OnStop; stop != nil {
		hook.OnStop = func(ctx context.Context) error {
			*l.order = append(*l.order, "audit-client")
			return stop(ctx)
		}
	}
	l.Lifecycle.Append(hook)
}

func TestAuditClientStopsBeforeBothCollectors(t *testing.T) {
	var order []string
	app := fx.New(
		fx.Supply(&audit.Writer{}),
		fx.Supply(&settings.Settings{Cfg: &runnerconfig.Config{IsNuonctl: true}}),
		fx.Supply(fx.Annotate(zap.NewNop(), fx.ResultTags(`name:"system"`))),
		fx.Provide(
			func(lc fx.Lifecycle) *Supervisor {
				lc.Append(fx.Hook{OnStop: func(context.Context) error {
					order = append(order, "audit-collector")
					return nil
				}})
				return &Supervisor{}
			},
			func(lc fx.Lifecycle) *VendorSupervisor {
				lc.Append(fx.Hook{OnStop: func(context.Context) error {
					order = append(order, "vendor-collector")
					return nil
				}})
				return &VendorSupervisor{}
			},
			asAuditRouteLifecycle,
			func(params audit.ClientParams) *audit.Client {
				params.Lifecycle = recordingLifecycle{Lifecycle: params.Lifecycle, order: &order}
				return audit.NewClient(params)
			},
		),
		fx.Invoke(func(*audit.Client) {}),
		fx.Invoke(func(*Supervisor, *VendorSupervisor) {}),
		fx.NopLogger,
	)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := app.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if err := app.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if len(order) != 3 || order[0] != "audit-client" || !slices.Contains(order, "audit-collector") || !slices.Contains(order, "vendor-collector") {
		t.Fatalf("audit client did not stop before both collectors: %q", order)
	}
}
