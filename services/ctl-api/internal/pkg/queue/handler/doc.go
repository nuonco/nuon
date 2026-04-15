// Package handler implements the per-signal Temporal workflow that validates,
// executes, and cancels queue signals.
//
// # Architecture: Server → Handler → Signal (with embedded Hooks)
//
// The queue system is structured like an HTTP server stack.
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
//	│  handler/ (this package)                                 │
//	│                                                         │
//	│  The handler workflow runs the following lifecycle:       │
//	│                                                         │
//	│  1. initializeState (state.go)                           │
//	│     • Loads signal from DB + catalog deserialization      │
//	│     • Calls ApplyParams for config/validator injection    │
//	│     • Populates infrastructure fields on embedded Hooks:  │
//	│       QueueSignalID, QueueID, SignalType, OrgID          │
//	│     • Calls ApplyInit (if signal implements               │
//	│       SignalWithInit) so it can set domain fields:         │
//	│       InstallID, ComponentID, Operation, LogStreamID     │
//	│                                                         │
//	│  2. validateHandler (validate.go)                        │
//	│     • hooks.PreExecuteHooks(ctx, "validate")             │
//	│     • sig.Validate(ctx)                                  │
//	│     • hooks.PostExecuteHooks(ctx, event, outcome)        │
//	│                                                         │
//	│  3. executeHandler (execute.go)                          │
//	│     • hooks.PreExecuteHooks(ctx, "execute")              │
//	│     • sig.Execute(ctx)                                   │
//	│     • hooks.PostExecuteHooks(ctx, event, outcome)        │
//	│                                                         │
//	│  4. cancelHandler (cancel.go) — via update at any time   │
//	│     • Cancels executing context if mid-execute            │
//	│     • Calls sig.Cancel(ctx) if SignalWithCancel           │
//	│     • hooks.PostExecuteHooks(ctx, event, outcome)        │
//	│                                                         │
//	│  PreExecuteHooks are fail-open: if the activity errors,  │
//	│  execution is allowed. PostExecuteHooks use a             │
//	│  disconnected context so cancellation doesn't prevent    │
//	│  delivery.                                               │
//	└────────────────────┬────────────────────────────────────┘
//	                     │
//	                     ▼
//	┌─────────────────────────────────────────────────────────┐
//	│  signal.Signal implementations  (the "handlers")        │
//	│                                                         │
//	│  Every signal embeds signal.Hooks to satisfy the         │
//	│  GetHooks() *Hooks method on the Signal interface.       │
//	│                                                         │
//	│  Required interface:                                     │
//	│    Type() SignalType                                     │
//	│    GetHooks() *Hooks       (via embedding signal.Hooks)  │
//	│    Validate(workflow.Context) error                      │
//	│    Execute(workflow.Context) error                       │
//	│                                                         │
//	│  Optional interfaces:                                    │
//	│    SignalWithInit   — Init(ctx) to set domain fields     │
//	│    SignalWithParams — WithParams(*Params) for config     │
//	│    SignalWithCancel — Cancel(ctx) for cleanup on cancel  │
//	│    SleepAfter       — controls handler workflow linger   │
//	│                                                         │
//	│  In Init(), signals set domain-specific Hooks fields:    │
//	│                                                         │
//	│    func (s *MySignal) Init(ctx workflow.Context) error { │
//	│        s.Hooks.InstallID = &s.InstallID                  │
//	│        s.Hooks.ComponentID = &s.ComponentID              │
//	│        s.Hooks.Operation = "component-deploy"            │
//	│        return nil                                        │
//	│    }                                                     │
//	└─────────────────────────────────────────────────────────┘
//
// # Lifecycle hooks
//
// Hooks observe signal phases without modifying signal logic. They run as
// Temporal activities (signal/lifecycle_activities.go), so they have access to
// DB, HTTP clients, and other dependencies.
//
// Current implementations (signal/hooks/):
//   - WebhookSignalLifecycleHook — delivers webhook events for user-facing
//     operations (e.g. component-deploy, sandbox-provision). Fires on execute
//     and cancel phases; skips validate. Publishes "before" on PreExecute and
//     "after" on PostExecute. Supports both configured URLs and per-org DB
//     webhook subscriptions.

// # Adding lifecycle hooks
//
// Hook authors never interact with the handler or Hooks struct directly.
//
//  1. Implement signal.SignalLifecycleHook (Name, Supports, PreExecute,
//     PostExecute) in a new file under signal/hooks/.
//  2. Register it via fx.Provide(signal.AsSignalLifecycleHook(...)) in
//     fxmodules/workers_shared.go.
//
// The hook's Supports(event) method controls which phases and operations it
// runs for. PreExecute can return Allow: false to block execution.
// PostExecute errors are logged but do not fail the signal.
package handler
