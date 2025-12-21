# Cross-Namespace Signal Sending Plan

## Overview

This plan addresses how to send signals from one queue (e.g., installs) to another queue in a different namespace (e.g., runners) using the new queue-based system instead of the old `evClient` event loop pattern.

## Problem Statement

The original event loop system used `evClient.Send(ctx, runnerID, signal)` to send signals across namespaces:

```go
// OLD: Event loop pattern
w.evClient.Send(ctx, install.RunnerID, &runnersignals.Signal{
    Type: runnersignals.OperationProvisionServiceAccount,
})
```

In the new queue-based system, we need to:
1. Find the target queue by its owner (e.g., a runner)
2. Enqueue a signal to that queue
3. Optionally await the signal's completion

## Architecture

### Queue Ownership Model

Queues have an owner relationship via `OwnerID` and `OwnerType` fields:

```go
type Queue struct {
    ID        string
    OrgID     string
    OwnerID   string  // e.g., runner ID
    OwnerType string  // e.g., "runners"
    Workflow  signaldb.WorkflowRef
    // ...
}
```

### Queue Client Methods

The queue client provides two key methods for cross-namespace communication:

1. **EnqueueSignal** - Sends a signal to a queue
2. **AwaitSignal** - Waits for a signal to complete

## Implementation Pattern

### Step 1: Queue Lookup by Owner

We need a new activity to find a queue by its owner:

```go
// Location: internal/pkg/queue/client/get.go (add new method)

func (c *Client) GetQueueByOwner(ctx context.Context, ownerID, ownerType string) (*app.Queue, error) {
    var q app.Queue
    if res := c.db.WithContext(ctx).
        First(&q, "owner_id = ? AND owner_type = ?", ownerID, ownerType); res.Error != nil {
        return nil, errors.Wrap(res.Error, "unable to get queue by owner")
    }
    return &q, nil
}
```

### Step 2: Activity Wrapper for Queue Operations

Create activities for queue operations in the shared activities package:

```go
// Location: internal/app/shared/worker/activities/queue.go (new file)

type QueueActivities struct {
    queueClient *queueclient.Client
}

type EnqueueSignalToOwnerRequest struct {
    OwnerID    string        `validate:"required"`
    OwnerType  string        `validate:"required"`
    Signal     signal.Signal `validate:"required"`
}

type EnqueueSignalToOwnerResponse struct {
    QueueSignalID string
}

// @temporal-gen activity
func (a *QueueActivities) EnqueueSignalToOwner(
    ctx context.Context,
    req *EnqueueSignalToOwnerRequest,
) (*EnqueueSignalToOwnerResponse, error) {
    // 1. Find the queue by owner
    queue, err := a.queueClient.GetQueueByOwner(ctx, req.OwnerID, req.OwnerType)
    if err != nil {
        return nil, errors.Wrap(err, "unable to find queue for owner")
    }

    // 2. Enqueue the signal to that queue
    resp, err := a.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
        QueueID: queue.ID,
        Signal:  req.Signal,
    })
    if err != nil {
        return nil, errors.Wrap(err, "unable to enqueue signal")
    }

    return &EnqueueSignalToOwnerResponse{
        QueueSignalID: resp.QueueSignalID,
    }, nil
}

// @temporal-gen activity
func (a *QueueActivities) AwaitQueueSignal(
    ctx context.Context,
    queueSignalID string,
) (*handler.FinishedResponse, error) {
    return a.queueClient.AwaitSignal(ctx, queueSignalID)
}
```

### Step 3: Usage in Signals

#### Pattern A: Fire-and-Forget (No Wait)

For signals that don't need to wait for completion (e.g., `provisionrunner`):

```go
// Location: internal/app/installs/signals/v2/provisionrunner/signal.go

func (s *Signal) Execute(ctx workflow.Context) error {
    install, err := activities.AwaitGet(ctx, activities.GetRequest{InstallID: s.InstallID})
    if err != nil {
        return errors.Wrap(err, "unable to get install")
    }

    // Send signal to runner's queue
    _, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
        OwnerID:   install.RunnerID,
        OwnerType: plugins.TableName(nil, &app.Runner{}), // "runners"
        Signal: &runnersignals.ProvisionServiceAccountSignal{
            RunnerID: install.RunnerID,
        },
    })
    if err != nil {
        return errors.Wrap(err, "unable to send provision signal to runner")
    }

    return nil
}
```

#### Pattern B: Await Completion (Wait for Result)

For signals that need to wait for the cross-namespace operation to complete:

```go
func (s *Signal) Execute(ctx workflow.Context) error {
    install, err := activities.AwaitGet(ctx, activities.GetRequest{InstallID: s.InstallID})
    if err != nil {
        return errors.Wrap(err, "unable to get install")
    }

    // Send signal to runner's queue
    resp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
        OwnerID:   install.RunnerID,
        OwnerType: plugins.TableName(nil, &app.Runner{}),
        Signal: &runnersignals.ProvisionServiceAccountSignal{
            RunnerID: install.RunnerID,
        },
    })
    if err != nil {
        return errors.Wrap(err, "unable to send provision signal to runner")
    }

    // Wait for the signal to complete
    finishedResp, err := sharedactivities.AwaitAwaitQueueSignal(ctx, resp.QueueSignalID)
    if err != nil {
        return errors.Wrap(err, "runner provisioning failed")
    }

    if finishedResp.Error != nil {
        return errors.Errorf("runner provisioning failed: %s", *finishedResp.Error)
    }

    return nil
}
```

## Required Changes

### 1. Queue Client Enhancement

**File**: `internal/pkg/queue/client/get.go`

Add method:
```go
func (c *Client) GetQueueByOwner(ctx context.Context, ownerID, ownerType string) (*app.Queue, error)
```

### 2. Shared Queue Activities

**File**: `internal/app/shared/worker/activities/queue.go` (new file)

Create:
- `QueueActivities` struct with queue client dependency
- `EnqueueSignalToOwner()` activity
- `AwaitQueueSignal()` activity
- Register activities in FX dependency injection

### 3. Runner Signal Definitions

**File**: `internal/app/runners/signals/v2/` (new directory structure)

Create signal structs that implement `signal.Signal` interface:
- `ProvisionServiceAccountSignal`
- `ReprovisionServiceAccountSignal`

Each with:
```go
type ProvisionServiceAccountSignal struct {
    RunnerID string `json:"runner_id"`
}

func (s *ProvisionServiceAccountSignal) Type() signal.SignalType {
    return "provision-service-account"
}

func (s *ProvisionServiceAccountSignal) Validate() error {
    if s.RunnerID == "" {
        return errors.New("runner_id is required")
    }
    return nil
}
```

### 4. Update Install Signals

**Files**:
- `internal/app/installs/signals/v2/provisionrunner/signal.go`
- `internal/app/installs/signals/v2/reprovisionrunner/signal.go`

Replace TODO comments with actual queue-based signal sending using Pattern A or B above.

## Benefits

1. **Type Safety**: Signal structs are strongly typed, not generic event loop signals
2. **Observability**: Queue signals are persisted in the database with status tracking
3. **Reliability**: Built on Temporal's workflow update mechanism (not fire-and-forget signals)
4. **Flexibility**: Can choose fire-and-forget or await-completion patterns
5. **Consistent Pattern**: Same queue system across all namespaces

## Migration Strategy

1. **Phase 1**: Add `GetQueueByOwner()` to queue client
2. **Phase 2**: Create shared queue activities
3. **Phase 3**: Define runner signal v2 types
4. **Phase 4**: Update install signals to use new pattern
5. **Phase 5**: Test cross-namespace communication end-to-end

## Open Questions

1. **Error Handling**: Should fire-and-forget pattern log errors or fail hard?
   - Recommendation: Log and continue (non-blocking)

2. **Timeout Configuration**: What timeout for await pattern?
   - Recommendation: Use default activity timeout or make configurable per signal type

3. **Owner Type Naming**: Use table names or constants?
   - Recommendation: Use `plugins.TableName(nil, &app.Runner{})` for consistency

4. **Multiple Queues per Owner**: Can one owner have multiple queues?
   - Current assumption: 1:1 relationship (one queue per owner)
   - If needed: Add queue name/type to lookup criteria

## Examples

### Current Signals Affected

1. **provisionrunner** - Sends `OperationProvisionServiceAccount` to runner
2. **reprovisionrunner** - Sends `OperationReprovisionServiceAccount` to runner

Both should use **Pattern A** (fire-and-forget) as they don't wait for completion in the original implementation.

## Testing Approach

1. **Unit Tests**: Mock queue client in signal tests
2. **Integration Tests**: Test queue lookup, enqueue, and await in integration test suite
3. **End-to-End Tests**: Create install → verify runner receives signal → verify completion
