# Queues

## Core Requirements

This package is the backbone of our distributed workflow tooling in the product. Different objects in our system are
backed by these, meaning that when an org, app, install or other is created we automatically spin up a queue to send
signals too.

Signals are workflows that do a user action, such as generating a plan, waiting for approvals, or running a deploy. It's
important that signals can be decoupled from each other and live in different directories. We have a registration system
where we use `init` functions to pick

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

## Testing

We are building tests into the core system here, so that we can easily make sure the core queue tooling works properly
under different failure scenarios. I am trying to make sure that we also run the worker itself, vs just using the
temporal stubbing tooling.
