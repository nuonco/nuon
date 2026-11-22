// Package installcomponenthealth runs pkg/runner/componenthealth's periodic
// report for a runnerless install (install_config.use_runner=false) on the
// control plane, since no Runner process exists to run componenthealth.Engine
// itself.
//
// Activities and Workflows are provided in fxmodules/workers_shared.go, and
// the control-plane-installs worker (services/ctl-api/internal/app/installs/
// worker/worker.go) registers InstallComponentHealthWorkflow and every
// activity from Activities.AllActivities() on that same worker.
//
// Health probe types expected to fail for a runnerless install's private
// endpoints when run from the control plane, which sits outside the
// install's VPC unlike a Runner: any TCP or HTTP(S) probe target that
// resolves to a private IP or a Kubernetes-internal DNS name (ClusterIP
// services, in-cluster ingress backends, VPC-only load balancers). These
// time out or fail to resolve from ctl-api and should report degraded or
// unknown, not panic the report cycle -- collectProbes already isolates one
// failing probe from the rest with a per-target goroutine and a
// panic-recovered WaitGroup.
//
// Cluster API access only works because install credentials route through
// the EKS/AKS/GKE control plane's public or peered API endpoint, not the
// data plane. A cluster configured with a fully private API endpoint and no
// peering or VPN from ctl-api fails cluster collection entirely, reported as
// ClusterAccessError rather than a crash.
//
// Terraform-sourced identity rows (collectTerraform) and Kubernetes object
// collection through the cluster's API server (collectCluster) are not
// network probes and are unaffected by this.
package installcomponenthealth

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// TaskQueue is where this workflow and its ReportComponentHealth activity run:
// the same control-plane queue every runnerless install job already runs on
// (controlplanejob.TaskQueueInstalls), so the control-plane-installs worker
// needs no second queue or poller.
const TaskQueue = "control-plane-installs"

// StopSignalName cancels the per-install reporting loop. Sent when
// ReconcileUseRunner moves the runner back to awaiting-heartbeat (a real
// Runner has shown up), or when the install is deprovisioned.
const StopSignalName = "install-component-health-stop"

// reportInterval mirrors componenthealth.reportInterval, so a runnerless
// install's reports age at the same rate a Runner's would.
const reportInterval = 60 * time.Second

// ticksPerRun bounds one workflow execution's history to ~200 report cycles
// (~3.3 hours at reportInterval) before continuing as new, so a long-lived
// install's history never grows unbounded.
const ticksPerRun = 200

// WorkflowID returns the deterministic, per-install workflow ID. Starting the
// workflow twice for the same install (e.g. a retried ReconcileUseRunner
// activity) is therefore idempotent: Temporal rejects a second start of an
// already-running execution with the same ID.
func WorkflowID(installID string) string {
	return "install-component-health-" + installID
}

type Workflows struct{}

func NewWorkflows() *Workflows { return &Workflows{} }

// Request is both the start payload and the continue-as-new carryover.
type Request struct {
	InstallID string `json:"install_id" validate:"required"`
	RunnerID  string `json:"runner_id" validate:"required"`
}

// InstallComponentHealthWorkflow is the runnerless-install stand-in for
// pkg/runner/componenthealth.Engine's background ticker: every reportInterval
// it runs one ReportComponentHealth activity on TaskQueue, which builds a
// throwaway Engine scoped to this install and pushes exactly one report
// through the same CreateComponentHealth path a Runner process would use.
//
// It runs until StopSignalName arrives, and continues-as-new every
// ticksPerRun cycles to keep workflow history bounded indefinitely. A failed
// report cycle is logged and retried on the next tick rather than failing the
// workflow, mirroring Engine.reportSafe swallowing a report() error.
func (w *Workflows) InstallComponentHealthWorkflow(ctx workflow.Context, req Request) error {
	l := workflow.GetLogger(ctx)
	stopCh := workflow.GetSignalChannel(ctx, StopSignalName)

	actCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskQueue:           TaskQueue,
		StartToCloseTimeout: 45 * time.Second,
		HeartbeatTimeout:    30 * time.Second,
	})

	for i := 0; i < ticksPerRun; i++ {
		if stopCh.ReceiveAsync(nil) {
			l.Info("stopping install component health workflow", "install_id", req.InstallID)
			return nil
		}

		if err := AwaitReportComponentHealth(actCtx, ReportComponentHealthRequest{Req: &ReportComponentHealthInput{
			InstallID: req.InstallID,
			RunnerID:  req.RunnerID,
		}}); err != nil {
			l.Warn("install component health report failed", "install_id", req.InstallID, "error", err.Error())
		}

		if i == ticksPerRun-1 {
			break
		}

		stopped := false
		selector := workflow.NewSelector(ctx)
		selector.AddReceive(stopCh, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)
			stopped = true
		})
		selector.AddFuture(workflow.NewTimer(ctx, reportInterval), func(f workflow.Future) {})
		selector.Select(ctx)
		if stopped {
			l.Info("stopping install component health workflow", "install_id", req.InstallID)
			return nil
		}
	}

	return workflow.NewContinueAsNewError(ctx, w.InstallComponentHealthWorkflow, Request{
		InstallID: req.InstallID,
		RunnerID:  req.RunnerID,
	})
}
