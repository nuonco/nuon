package componenthealth

import (
	"context"
	"time"

	"github.com/sourcegraph/conc"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"k8s.io/apimachinery/pkg/runtime/schema"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

const (
	reportInterval = 60 * time.Second
	reportTimeout  = 45 * time.Second
)

var ownerGVRs = map[string]schema.GroupVersionResource{
	"ReplicaSet":              {Group: "apps", Version: "v1", Resource: "replicasets"},
	"Deployment":              {Group: "apps", Version: "v1", Resource: "deployments"},
	"StatefulSet":             {Group: "apps", Version: "v1", Resource: "statefulsets"},
	"DaemonSet":               {Group: "apps", Version: "v1", Resource: "daemonsets"},
	"Job":                     {Group: "batch", Version: "v1", Resource: "jobs"},
	"HorizontalPodAutoscaler": {Group: "autoscaling", Version: "v2", Resource: "horizontalpodautoscalers"},
}

var watchedGVRs = []schema.GroupVersionResource{
	{Group: "apps", Version: "v1", Resource: "deployments"},
	{Group: "apps", Version: "v1", Resource: "statefulsets"},
	{Group: "apps", Version: "v1", Resource: "daemonsets"},
	{Group: "", Version: "v1", Resource: "pods"},
	{Group: "", Version: "v1", Resource: "services"},
	{Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
	{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
	{Group: "batch", Version: "v1", Resource: "jobs"},
}

type Params struct {
	fx.In

	APIClient     nuonrunner.Client
	L             *zap.Logger `name:"system"`
	LC            fx.Lifecycle
	Process       string `name:"process"`
	Cluster       *ClusterProvider
	Terraform     *TerraformProvider
	ManifestKinds *ManifestKindsProvider
}

type Engine struct {
	l             *zap.Logger
	apiClient     nuonrunner.Client
	process       string
	cluster       *ClusterProvider
	terraform     *TerraformProvider
	manifestKinds *ManifestKindsProvider
	idx           *index

	ctx      context.Context
	cancelFn func()
	wg       *conc.WaitGroup
}

func New(params Params) (*Engine, error) {
	ctx, cancelFn := context.WithCancel(context.Background())
	e := &Engine{
		l:             params.L,
		apiClient:     params.APIClient,
		process:       params.Process,
		cluster:       params.Cluster,
		terraform:     params.Terraform,
		manifestKinds: params.ManifestKinds,
		idx:           newIndex(),
		ctx:           ctx,
		cancelFn:      cancelFn,
		wg:            conc.NewWaitGroup(),
	}

	params.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			e.wg.Go(func() { e.run(e.ctx) })
			return nil
		},
		OnStop: func(context.Context) error {
			e.cancelFn()
			e.wg.Wait()
			return nil
		},
	})

	return e, nil
}

func (e *Engine) run(ctx context.Context) {
	if e.process != "install" {
		return
	}

	e.cluster.Load(ctx)
	if e.manifestKinds != nil {
		e.manifestKinds.Load()
	}

	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()

	e.reportSafe(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if e.cluster.Get() == nil {
				e.cluster.Load(ctx)
				if e.manifestKinds != nil {
					e.manifestKinds.Load()
				}
			}
			e.reportSafe(ctx)
		}
	}
}

func (e *Engine) reportSafe(ctx context.Context) {
	reportCtx, cancel := context.WithTimeout(ctx, reportTimeout)
	defer cancel()
	if err := e.report(reportCtx); err != nil {
		e.l.Error("unable to report component health", zap.Error(err))
	}
}
