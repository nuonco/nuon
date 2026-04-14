// Package handler implements the per-signal Temporal workflow that validates,
// executes, and cancels queue signals.
//
// # Architecture: Server → Middleware → Handler
//
// The queue system is structured like an HTTP server stack. Understanding this
// analogy is essential for knowing where to put new logic.
//
//	┌─────────────────────────────────────────────────────────┐
//	│  queue/handle_signal.go  (the "server / router")        │
//	│                                                         │
//	│  Dispatches signals to handler workflows via Temporal    │
//	│  update calls: ready → validate → execute.              │
//	│                                                         │
//	│  Has: queue signal ID, queue ID, status from DB.        │
//	│  Does NOT have: the deserialized signal object, so no   │
//	│  access to install ID, component ID, operation name,    │
//	│  log stream ID, or any signal-specific context.         │
//	└────────────────────┬────────────────────────────────────┘
//	                     │ Temporal workflow update
//	                     ▼
//	┌─────────────────────────────────────────────────────────┐
//	│  handler/middleware.go  (the "middleware layer")         │
//	│                                                         │
//	│  withLifecycle(ctx, phase, innerFunc):                   │
//	│    1. Builds a SignalPhaseEvent from the handler's       │
//	│       in-memory state — including install ID, component  │
//	│       ID, and operation from the deserialized signal     │
//	│       (via SignalWithLifecycleContext interface).         │
//	│    2. Runs before-phase lifecycle hooks. Two outcomes:     │
//	│                                                         │
//	│       a) Hook activity fails (timeout, crash, etc):     │
//	│          → Fail-open: treat as allowed, continue.       │
//	│          This ensures infrastructure issues in hooks     │
//	│          never block signal execution.                   │
//	│                                                         │
//	│       b) Hook activity succeeds and returns a decision: │
//	│          → Allow=true: continue to inner function.      │
//	│          → Allow=false: block execution, return error.  │
//	│          A hook might return Allow=false to enforce a    │
//	│          policy, e.g. "this terraform workspace is locked, reject    │
//	│          deploys" or "policy violated".              │
//	│                                                         │
//	│    3. Calls the inner function (sig.Validate/Execute).   │
//	│    4. Runs after-phase lifecycle hooks (best-effort,     │
//	│       disconnected context) with outcome including       │
//	│       duration, error, and log stream metadata.          │
//	│                                                         │
//	│  afterLifecycle(ctx, phase, outcome):                    │
//	│    After-phase only variant for cancel, which IS the     │
//	│    operation rather than wrapping one.                   │
//	│                                                         │
//	│  This layer exists inside the handler workflow because   │
//	│  it needs access to the signal object that only exists   │
//	│  in-memory after deserialization. The router layer       │
//	│  (handle_signal.go) cannot provide this context — it     │
//	│  sits on the wrong side of the Temporal update boundary. │
//	└────────────────────┬────────────────────────────────────┘
//	                     │
//	                     ▼
//	┌─────────────────────────────────────────────────────────┐
//	│  signal.Signal implementations  (the "handlers")        │
//	│                                                         │
//	│  sig.Validate(ctx) — validates the signal is runnable.  │
//	│  sig.Execute(ctx)  — performs the actual work.           │
//	│                                                         │
//	│  Each signal type lives in its own package and is        │
//	│  registered via init(). The handler workflow             │
//	│  deserializes it from the DB and calls through the       │
//	│  middleware layer.                                       │
//	└─────────────────────────────────────────────────────────┘
//
// # Why the middleware lives in the handler, not the router
//
// The router (handle_signal.go) orchestrates signals by calling Temporal
// workflow updates on the handler workflow. At that layer, the signal is just
// a database row — a queue signal ID with a type and serialized payload. The
// deserialized signal object (with InstallID, ComponentID, Operation, and
// LogStreamID) only exists inside the handler workflow after initialization.
//
// "In-memory" here does NOT mean fragile — Temporal's event sourcing makes
// this crash-safe. The signal is loaded via activities (initializeState),
// whose results are recorded in the workflow event history. If the worker
// crashes, Temporal replays the activity results on a new worker, fully
// reconstructing the handler's state. The point is that this state is local
// to the handler workflow's goroutine and not accessible from the router,
// which runs in a separate activity on the queue workflow.
//
// Lifecycle hooks need that rich context to make meaningful decisions:
//   - The webhook hook uses Operation to filter user-facing events.
//   - The log stream cleanup hook reads LogStreamID from the outcome metadata.
//   - Before-phase hooks can block execution based on install/component context.
//
// Moving hooks to the router would require serializing all this context back
// through activity parameters, losing type-safe interface access and coupling
// the router to signal internals it shouldn't know about.
//
// # Adding lifecycle hooks
//
// Hook authors never interact with this middleware layer. To add a new hook:
//  1. Implement signal.SignalLifecycleHook in signal/hooks/.
//  2. Register it via fx.Provide(signal.AsSignalLifecycleHook(...)) in
//     fxmodules/workers_shared.go.
//
// The hook's Supports() method controls which phases and operations it runs
// for. The middleware calls hooks via Temporal activities, so hooks run in a
// normal activity context with access to DB, HTTP clients, etc.
//

package handler
