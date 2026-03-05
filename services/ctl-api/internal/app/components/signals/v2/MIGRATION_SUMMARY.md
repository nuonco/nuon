# Component Signals V2 Migration Summary

This document summarizes the migration of all component workflow signals from the old event loop pattern to the new v2 queue-based system.

## Migration Overview

All 9 component signals have been successfully migrated from `/internal/app/components/worker/` to `/internal/app/components/signals/v2/`.

## Signal Mapping

| Old Signal Type | New Signal Type | Directory | Complexity | Status |
|----------------|-----------------|-----------|------------|--------|
| `created` | `component-created` | `/v2/created/` | Simple | ✅ Complete |
| `restart` | `component-restart` | `/v2/restart/` | Simple | ✅ Complete |
| `provision` | `component-provision` | `/v2/provision/` | Simple (noop) | ✅ Complete |
| `update_component_type` | `component-update-component-type` | `/v2/updatecomponenttype/` | Simple | ✅ Complete |
| `config_created` | `component-config-created` | `/v2/configcreated/` | Simple | ✅ Complete |
| `queue_build` | `component-queue-build` | `/v2/queuebuild/` | Medium | ✅ Complete |
| `poll_dependencies` | `component-poll-dependencies` | `/v2/polldependencies/` | Medium | ✅ Complete |
| `build` | `component-build` | `/v2/build/` | Complex | ✅ Complete |
| `delete` | `component-delete` | `/v2/delete/` | Complex | ✅ Complete |

## Implementation Details

### Phase 1: Simple Signals (5 signals)

**created** (`/v2/created/`)
- Updates component status to `active`
- Uses `updateStatus` helper method
- Source: `/worker/provision.go:Created()`

**restart** (`/v2/restart/`)
- Updates component status to `active` (same as created)
- Uses `updateStatus` helper method
- Source: `/worker/restarted.go:Restarted()`

**provision** (`/v2/provision/`)
- No-op signal (workflow does nothing)
- Source: `/worker/provision.go:Provision()`

**updatecomponenttype** (`/v2/updatecomponenttype/`)
- Updates component type via activity
- Requires `ComponentType` field
- Source: `/worker/update_component_type.go:UpdateComponentType()`

**configcreated** (`/v2/configcreated/`)
- Identical logic to queuebuild
- Gets component then queues a build
- Source: Same as queuebuild pattern

### Phase 2: Medium Complexity Signals (2 signals)

**queuebuild** (`/v2/queuebuild/`)
- Gets component by ID
- Queues a component build with org context
- Source: `/worker/queue_build.go:QueueBuild()`

**polldependencies** (`/v2/polldependencies/`)
- Polls component's app until status is `active`
- 10 second poll interval
- Fails if app status becomes `error`
- Source: `/worker/poll_dependencies.go:PollDependencies()`

### Phase 3: Complex Signals (2 signals)

**build** (`/v2/build/`)
- Most complex signal with multiple helper methods
- Creates log stream for build execution
- Validates component is active
- Executes build via `execBuild` helper
- Sends notifications on failure
- Updates build status through multiple states
- Helper methods copied:
  - `execBuild()` - Executes the actual build workflow
  - `updateBuildStatus()` - Updates build status (calls both v1 and v2 activities)
  - `sendNotification()` - Sends email and Slack notifications
- Source: `/worker/build.go:Build()` + `/worker/exec_build.go:execBuild()`

**delete** (`/v2/delete/`)
- Polls until component is unused (60 minute timeout)
- Checks if component is in active app config
- Checks if component has dependent installs
- Updates status through multiple states
- Helper methods copied:
  - `pollComponentBeingUnused()` - Polls for component being safe to delete
  - `updateStatus()` - Updates component status
- Source: `/worker/delete.go:Delete()`

## Field Mapping

Old signals used `sreq.ID` for component ID. New signals use explicit fields:

- `ComponentID` - Component identifier (required in all signals)
- `BuildID` - Build identifier (required only in `build` signal)
- `SandboxMode` - Sandbox mode flag (optional in `build` signal)
- `ComponentType` - Component type (required in `updatecomponenttype` signal)

## Helper Methods Pattern

Helper methods from `/worker/` were copied as private methods on signal structs:

- `updateStatus()` - Used by: created, restart, delete
- `updateBuildStatus()` - Used by: build
- `sendNotification()` - Used by: build
- `pollComponentBeingUnused()` - Used by: delete (already in same file)
- `execBuild()` - Used by: build (already in same file)

## Testing Structure

Each signal includes integration tests following the `testworker` pattern:

```go
type SignalTestSuite struct {
    suite.Suite
    app     *fxtest.App
    service testworker.TestService
}
```

Tests verify:
- Signal validation
- Happy path execution
- Error handling
- Status updates
- Edge cases

## Next Steps

1. **Update component service files** to use v2 signals instead of old event loop signals
2. **Run code generation** with `./run-nuonctl.sh scripts reset-generated-code`
3. **Run integration tests** with `INTEGRATION=true go test ./internal/app/components/signals/v2/...`
4. **Remove old worker methods** after verifying v2 signals work in production

## Architecture Benefits

The v2 signal system provides:

- **Type Safety** - Explicit field definitions with validation
- **Testability** - Each signal is independently testable
- **Maintainability** - Clear separation of concerns
- **Observability** - Better logging and error handling
- **Extensibility** - Easy to add new signals following established pattern

## Files Created

Total files created: 27 (3 per signal × 9 signals)

```
/internal/app/components/signals/v2/
├── build/
│   ├── init.go
│   ├── signal.go
│   └── signal_test.go
├── configcreated/
│   ├── init.go
│   ├── signal.go
│   └── signal_test.go
├── created/
│   ├── init.go
│   ├── signal.go
│   └── signal_test.go
├── delete/
│   ├── init.go
│   ├── signal.go
│   └── signal_test.go
├── polldependencies/
│   ├── init.go
│   ├── signal.go
│   └── signal_test.go
├── provision/
│   ├── init.go
│   ├── signal.go
│   └── signal_test.go
├── queuebuild/
│   ├── init.go
│   ├── signal.go
│   └── signal_test.go
├── restart/
│   ├── init.go
│   ├── signal.go
│   └── signal_test.go
└── updatecomponenttype/
    ├── init.go
    ├── signal.go
    └── signal_test.go
```

## Verification Checklist

- [x] All 9 signals implemented
- [x] All helper methods copied as private methods
- [x] All signals follow v2 pattern with init.go, signal.go, signal_test.go
- [x] All signals registered in catalog via init()
- [x] All signals implement signal.Signal interface
- [x] All signals have proper validation
- [x] All signals have integration tests
- [x] Code formatted with go fmt
