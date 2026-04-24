# Nuon v2 Signal System — A K8s Operator Analogy

## Overview

Nuon's v2 signal system is a custom orchestration layer built on top of [Temporal](https://temporal.io). It manages all asynchronous work in the platform — deploys, provisioning, teardowns, health checks, syncs — using a pattern that closely mirrors **Kubernetes Operators**, where a long-lived controller watches for declared intent and reconciles it to completion.

The system is **imperative** (run this task once) rather than declarative (maintain desired state), but the mechanics are the same: write intent to a durable store, a controller picks it up, drives it through a lifecycle, and recovers on crash.

---

## Core Concepts

### The Analogy at a Glance

| K8s Operator Concept | Nuon v2 Signal System | Description |
|---|---|---|
| **Controller / Operator** | **Queue** | A long-lived process per entity that watches for work and dispatches it |
| **Pod (spawned by a Job)** | **Handler** | An ephemeral execution context spawned per unit of work |
| **Reconcile logic** | **Signal** | The domain-specific business logic (Terraform plan, Helm install, etc.) |
| **Custom Resource** | **QueueSignal row** | A declaration of intent persisted to a durable store (Postgres) |
| **Status Conditions** | **Directives** | Control flow metadata written to Postgres that parents read to decide what to do next |
| **Informer List on startup** | **requeueSignals()** | On startup, query the store for incomplete work and re-dispatch it |
| **Sub-resources** | **Multiple queues per entity** | A single domain object can have multiple queues, each handling a different concern |

---

### Queue — The Operator

A **Queue** is a long-lived Temporal workflow bound to a single domain entity (one install, one runner, one VCS connection). It is the equivalent of a K8s controller's run loop.

**What it does:**
- Accepts work via Temporal Updates (analogous to a K8s watch receiving events)
- Controls concurrency with `MaxInFlight` (like a controller's work queue parallelism) and `MaxDepth` (backlog limit)
- Is pausable, stoppable, and restartable — operational levers for the entity it manages
- Uses Temporal's `ContinueAsNew` to keep its event history bounded (like a controller that periodically re-lists)
- Runs `requeueSignals()` on startup to recover incomplete work (like a controller's initial List)

**Lazy lifecycle:**
Queues are lazy — they self-terminate after 10 minutes of idle time and restart on-demand via Temporal's `UpdateWithStart`. This is analogous to a K8s controller that scales to zero when idle and is triggered back up by an event.

> **Pause vs. Lazy:** Pause means "there IS work but don't process it yet" (operational hold). Lazy means "no work, free resources" (resource efficiency). Both are needed.

**Code:** [`queue/queue.go`](../services/ctl-api/internal/pkg/queue/queue.go), [`queue/worker.go`](../services/ctl-api/internal/pkg/queue/worker.go)

---

### Handler — The Pod

A **Handler** is an ephemeral Temporal child workflow spawned per signal. It is the equivalent of a Pod spawned by a K8s Job — a generic execution container that runs a standard lifecycle regardless of what's inside.

**Lifecycle:**
1. **Ready** — Handler initializes, loads the signal from Postgres, registers update handlers
2. **Validate** — Runs `sig.Validate(ctx)` to check preconditions
3. **Execute** — Runs `sig.Execute(ctx)` to perform the actual work
4. **Cache** — After execution, stays alive briefly (default 1 min) so subsequent signals can reuse it via `UpdateWithStart` instead of spawning a new workflow

The queue communicates with the handler through **Temporal Updates wrapped in Activities**. Workflow code cannot make external RPC calls directly, so activities act as the bridge — similar to how a K8s controller uses the API server to communicate with Pods rather than calling them directly.

**Code:** [`queue/handler/handler.go`](../services/ctl-api/internal/pkg/queue/handler/handler.go), [`queue/handler/run.go`](../services/ctl-api/internal/pkg/queue/handler/run.go), [`queue/handler/execute.go`](../services/ctl-api/internal/pkg/queue/handler/execute.go)

---

### Signal — The Reconcile Logic

A **Signal** is a Go struct that implements the `Signal` interface:

```go
type Signal interface {
    Type() SignalType
    Validate(ctx workflow.Context) error
    Execute(ctx workflow.Context) error
}
```

This is the domain logic — the equivalent of the `Reconcile()` function inside a K8s controller. Signals are decoupled from each other and live in separate directories. They are registered at init-time via a catalog (like registering reconcilers with a controller manager).

Examples of signals:
- `componentdeploysyncandplan.Signal` — Syncs Terraform state and generates a plan
- `componentdeployapplyplan.Signal` — Applies a Terraform plan
- `awaitrunnerhealthy.Signal` — Waits for a runner to be healthy
- `generatestate.Signal` — Generates install state
- `componentsyncimage.Signal` — Syncs a container image

**Code:** [`queue/signal/signal.go`](../services/ctl-api/internal/pkg/queue/signal/signal.go)

---

### QueueSignal Row — The Custom Resource

When the system needs work done, it writes a **QueueSignal row** to Postgres with status `queued`. This is the declaration of intent — analogous to creating a Custom Resource in K8s.

The queue (operator) picks it up, drives it through the handler lifecycle, and updates the status to `in_progress`, then `success` or `error`. If the system crashes mid-execution, the signal row remains in Postgres and is recovered on restart via `requeueSignals()`.

---

### Directives — Status Conditions

**Directives** are the control flow mechanism between execution levels. They are string values written to Postgres by child signals and read by parent signals to decide what to do next. They are analogous to **status conditions on K8s sub-resources** — the parent controller reads the child's status to decide the next action.

| Directive | Meaning |
|---|---|
| `continue` | Step/group succeeded, proceed to next |
| `stop` | Terminal failure, stop the workflow |
| `retry` | Step failed, clone and retry just this step |
| `retry-group` | Step failed, clone and retry the entire group (plan + apply) |
| `skip-group` | Skip remaining steps in this group |
| `await-approval` | Pause execution and wait for human approval |

---

## Architecture: Multiple Queues per Entity

Just as a K8s object can have multiple sub-resources each handled by a different controller, a single domain entity can have **multiple queues**, each handling a different level of orchestration.

### Install Queues (3 queues per install)

```
┌──────────────────────────────────────────────────────────────────────┐
│                         Install Entity                               │
│                                                                      │
│  ┌─────────────────────┐  ┌─────────────────────┐  ┌──────────────┐ │
│  │  install-workflows   │  │ install-workflow-    │  │ install-     │ │
│  │  (MaxInFlight: 10)   │  │ steps                │  │ signals      │ │
│  │                      │  │ (MaxInFlight: 10)    │  │ (MIF: 20)   │ │
│  │  WHAT to do          │  │ IN WHAT ORDER        │  │ DO IT        │ │
│  │                      │  │                      │  │              │ │
│  │  execute-workflow    │  │  execute-workflow-   │  │ sync-and-   │ │
│  │  (ManualDeploy,     │  │  step-group          │  │ plan,       │ │
│  │   DeployAll, etc.)   │  │  execute-workflow-   │  │ apply-plan, │ │
│  │                      │  │  step                │  │ sync-image  │ │
│  └──────────┬───────────┘  └──────────┬───────────┘  └──────┬───────┘ │
│             │ enqueues to              │ enqueues to         │         │
│             └──────────────►           └─────────────►       │         │
└──────────────────────────────────────────────────────────────────────┘
```

### Other Entity Queues

| Entity | Queues | Purpose |
|---|---|---|
| **Runner** | 9 queues | One per job group: health-checks, sync, build, deploy, sandbox, runner, operations, management, actions |
| **VCS Connection** | 1 queue | Health checks, webhook subscriptions |
| **App Branch** | 1 queue | App-level operations |

---

## Concrete Example: Manual Deploy of a Terraform Component

Let's walk through deploying a Terraform component called **"api-server"** to an install, tracing the full path from API call to infrastructure change.

### Step 1: API Creates Intent

The API handler creates a `Workflow` record (type: `ManualDeploy`) in Postgres and enqueues an **`execute-workflow`** signal to the **`install-workflows`** queue.

> **K8s analogy:** A user runs `kubectl apply -f deploy.yaml`. The API server persists the Deployment resource and the Deployment controller's informer picks it up.

### Step 2: Step Generation

The `execute-workflow` signal calls `ManualDeploySteps()` — a **step generator function** (not a signal) that acts as a blueprint. It queries the database for the component graph and generates an ordered list of `WorkflowStep` records:

| Group | Step | Signal Type |
|---|---|---|
| 0 | Generate install state | `generatestate.Signal` |
| 1 | Await runner healthy | `awaitrunnerhealthy.Signal` |
| 2 | Pre-deploy action hooks | *(action workflow signals)* |
| 3 | Sync and plan api-server | `componentdeploysyncandplan.Signal` |
| 3 | Apply api-server | `componentdeployapplyplan.Signal` |
| 4 | Post-deploy action hooks | *(action workflow signals)* |

> **K8s analogy:** The Deployment controller computes the desired ReplicaSet spec from the Deployment's template.

### Step 3: Group-by-Group Execution

The `execute-workflow` signal iterates over step groups. For each group, it enqueues an **`execute-workflow-step-group`** signal to the **`install-workflow-steps`** queue.

The group signal then enqueues individual **`execute-workflow-step`** signals to the same queue — one per step in the group. Steps within a group can run in parallel or sequentially based on the group's `Parallel` flag.

> **K8s analogy:** The ReplicaSet controller creates individual Pods for each replica.

### Step 4: Domain Signal Execution

Each `execute-workflow-step` signal extracts the inner domain signal from the step record (e.g., `componentdeploysyncandplan.Signal`) and enqueues it to the **`install-signals`** queue. This is where the actual Terraform/Helm work happens.

The step signal then **blocks**, waiting for the domain signal to complete by polling its QueueSignal status via `AwaitSignal`.

> **K8s analogy:** The Pod runs the container image. The ReplicaSet controller watches the Pod's status conditions.

### Step 5: Directives Flow Upward

When the domain signal completes (success or failure), the step signal writes a **directive** to Postgres. The group signal reads it and decides what to do:

```
Domain Signal (install-signals)
    │ completes → status written to QueueSignal row
    ▼
Step Signal (install-workflow-steps)
    │ reads status → writes directive to WorkflowStep
    ▼
Group Signal (install-workflow-steps)
    │ reads directive → decides: continue / retry / stop
    ▼
Flow Signal (install-workflows)
    │ reads group result → moves to next group or terminates
    ▼
Workflow Status Updated
```

> **K8s analogy:** Pod status conditions → ReplicaSet status → Deployment status. Each level reads its children's status to compute its own.

### Step 6: Retry-Group (Plan + Apply Atomicity)

If `terraform apply` fails, the `componentdeployapplyplan.Signal` implements `SignalWithRetryGroup`:

```go
type SignalWithRetryGroup interface {
    RetryGroup() bool
}
```

This tells the group signal to use the **`retry-group`** directive instead of `retry`. The entire group (plan + apply) is cloned and re-run because you can't apply a stale plan. The group's `GroupRetryIdx` is incremented to track retry attempts.

> **K8s analogy:** Similar to how a StatefulSet will recreate the entire Pod (not just a container) on failure, because the containers have ordering dependencies.

---

## Crash Recovery

### requeueSignals() — The Informer List

On every queue startup (initial or after `ContinueAsNew`), the queue runs `requeueSignals()`:

1. Queries Postgres for QueueSignal rows with status `queued` or `in_progress`
2. Re-injects each one into the queue's internal dispatch channel
3. The dispatcher picks them up and spawns handlers as normal

> **K8s analogy:** When a controller starts, it performs a full List of all resources and adds them to its work queue. This ensures nothing is lost even if the controller was down.

**Note:** There is no periodic sweep. If a signal is dropped (e.g., channel full) and the queue never restarts, the signal remains orphaned until the queue's idle timeout triggers a restart.

---

## Cancellation

Cancel propagates **top-down through direct Temporal Updates**, bypassing the queue and directive systems entirely. This is a direct RPC chain:

```
API
 └──► execute-workflow handler      (cancel-workflow Update)
       └──► execute-step-group handler  (cancel Update)
             └──► execute-step handler      (cancel Update)
                   └──► domain signal handler     (cancel Update → workflow.Context cancellation)
```

> **K8s analogy:** `kubectl delete` sends a deletion signal directly to each Pod (via the API server), rather than waiting for the controller to notice and propagate it.

---

## Idempotency Model

The system does **not** require K8s-style idempotency (run reconcile repeatedly until convergence). Instead it requires:

| Requirement | Meaning |
|---|---|
| **Temporal determinism** | Workflow code must be replay-safe (no side effects, no non-deterministic calls) |
| **Activity-level idempotency** | Activities (the only place with side effects) must be safe to retry on transient failures |

Signals run once to completion — they are tasks, not reconciliation loops. But the overall system achieves the same reliability guarantees through Temporal's durable execution and the requeue mechanism.

---

## Summary Table

| Concept | K8s Operator | Nuon v2 Signals |
|---|---|---|
| Controller runtime | Controller Manager process | Temporal Worker |
| Controller instance | One per CRD | One Queue workflow per entity |
| Work item | Custom Resource event | QueueSignal row in Postgres |
| Execution unit | Pod | Handler (child workflow) |
| Business logic | Reconcile() | Signal.Execute() |
| Control flow | Status conditions + requeue | Directives in Postgres |
| Crash recovery | Informer re-list | requeueSignals() on startup |
| Concurrency control | Worker queue parallelism | MaxInFlight / MaxDepth |
| Cancellation | API server → Pod deletion | Direct Temporal Update chain |
| Resource efficiency | HPA scale-to-zero | Lazy queues (10 min idle timeout) |
| Operational control | Cordon / drain a node | Pause / stop a queue |
