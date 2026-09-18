# Queues

## Core Requirements

This package is the backbone of our distributed workflow tooling in the product. Different objects in our system are
backed by these, meaning that when an org, app, install or other is created we automatically spin up a queue to send
signals too.

Signals are workflows that do a user action, such as generating a plan, waiting for approvals, or running a deploy. It's
important that signals can be decoupled from each other and live in different directories. Signal packages register
themselves via `init` functions so the queue client can resolve handlers by signal type.

Some requirements of the queue system:

1. we can send signals and see them persisted into the DB.
1. a signal can be retried, cancelled, etc and if temporal fails can be recovered.
1. we can run signals on a cron, so that a queue can have an "auto emitter" that sends a signal on a timeline.
1. we can use signals for control-flow. Since they are just workflows with queries, we can always ping a workflow and 
   the query will respond by restarting the loop and then returning the data from the db if it's not running.

The queue system has a nice client that we can use anywhere. It uses activities for things, so we want to be able to
embed it and expose as much functionality as we can in different places.

## Queue Ownership Pattern

Queues use a **polymorphic relationship** via `OwnerID` and `OwnerType` fields on the `Queue` model. Owner models
should declare a GORM polymorphic association to access their queues — **do NOT store a `QueueID` foreign key** on the
owner model. The Queue already knows its owner.

**Correct pattern** (follow `Runner` as the reference implementation):

```go
// On the owner model:
Queues []Queue `json:"queues,omitzero" gorm:"polymorphic:Owner;polymorphicValue:vcs_connections"`

// When creating a queue, set OwnerID/OwnerType:
queueClient.Create(ctx, &queueclient.CreateQueueRequest{
    OwnerID:   owner.ID,
    OwnerType: "vcs_connections",
    // ...
})

// To access queues, use Preload:
db.Preload("Queues").First(&owner, "id = ?", ownerID)
```

**Anti-pattern — do NOT do this:**

```go
// ❌ Storing a QueueID on the owner and manually updating it
QueueID string `json:"queue_id" gorm:"default:null"`

db.Model(&Owner{}).Where("id = ?", id).Update("queue_id", q.ID)
```

The polymorphic relationship is the single source of truth. The Queue's `OwnerID`/`OwnerType` fields are indexed and
used for lookups. Adding a reverse FK creates redundancy and drift risk.

## Routing and layering

No queue may contain both a signal and a signal it awaits, directly or transitively. Treat each owner's queues as a
layered DAG: a parent runs in one layer and every callback-bearing child runs in a lower layer. `MaxInFlight` throttles
work within one layer; it must never be relied on to make a parent and child sharing that layer safe.

Queue names, defaults, and capacities live in `queuenames`. Name every enqueue with `QueueName` or `QueueID`.
`GetDefaultQueueByOwner` resolves a declared default. `GetOnlyQueueByOwner` is only for owner types declared as having
one dynamic queue. Do not look up a queue by owner and take the first row.

| Owner | Layers |
| --- | --- |
| `app_branches` | default → `app-branch-workflows` → generate/step-groups → steps → signals → sandbox/component/install queues |
| `apps` | default or `app-workflows` → generate/step-groups → steps → signals |
| `installs` | `install-workflows` → generate/step-groups → steps → `install-signals` → `install-approvals` |
| `components` | default → `component-workflow-steps` |
| `runners` | `runner-signals`, job-group queues, and process queues are separate execution lanes |
| `orgs` | `org-signals` |

A callback-bearing enqueue into the caller's queue is rejected as non-retryable. Fire-and-forget work may use the
caller's queue.

## Recovering a wedged queue

Cancel the affected run, force-restart the queue, then re-trigger the run. Confirm the replacement signals use the
declared layers before recovering more runs.

## Testing

We are building tests into the core system here, so that we can easily make sure the core queue tooling works properly
under different failure scenarios. I am trying to make sure that we also run the worker itself, vs just using the
temporal stubbing tooling.
