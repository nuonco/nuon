# Temporal Gen V2 Examples

This directory contains examples demonstrating the capabilities of `temporal-gen-v2`.
Each example file focuses on a specific feature set and is paired with a generated `_gen.go` file.

## Workflows (`workflow.go`)

Demonstrates the `@temporal-gen-v2 workflow` annotation.

*   **Simple Workflow**: Basic await wrapper generation.
    *   [Source](file://./workflow.go)
    *   [Generated](file://./workflow_gen.go)
*   **Complex Workflow**: Shows timeouts (`@execution-timeout`, `@task-timeout`), task queues (`@task-queue`), and wait policies.
*   **ID Templates**: Shows usage of `@id-template` with the new `{{.Req}}` and `{{.Info}}` accessors.
*   **Dynamic IDs**: Shows usage of `@id-generator` to call a Go function for ID generation.
*   **Exec Variant**: Every workflow also gets an `Exec` variant (e.g., `ExecSimpleWorkflow`) that returns a `workflow.ChildWorkflowFuture` for non-blocking execution.

## Activities (`activity.go`)

Demonstrates the `@temporal-gen-v2 activity` annotation.

*   **Simple Activity**: Basic await wrapper generation.
    *   [Source](file://./activity.go)
    *   [Generated](file://./activity_gen.go)
*   **Timeouts & Retries**: Shows `@schedule-to-close-timeout`, `@start-to-close-timeout`, and `@max-retries`.
*   **Wrapper Structs**: Shows `@as-wrapper` which bundles multiple arguments into a single struct.
*   **By Field**: Shows `@by-field` helper generation (useful for getter-style activities).
*   **Call-Time Customization**: All generated `Await` functions accept variadic `workflow.ActivityOptions` to override defaults at call time.

## Queries & Updates (`queries_updates.go`)

Demonstrates Client-side generation for `@query` and `@update`.

*   **Queries**: Generates type-safe methods on the client struct (e.g., `QueryHandler`).
    *   [Source](file://./queries_updates.go)
    *   [Generated](file://./queries_updates_gen.go)
*   **Updates**: Generates client methods for sending Updates.
    *   Supports `@id` to specify the Update ID.
    *   Supports `UpdateWithStart` via options.
*   **Signals**: (Not shown in file yet, but supported) Generates client methods for sending Signals, including `SignalWithStart`.

## Labels (`labels.go`)

Demonstrates `@label <key> <value>`, which pulls default options from a
`temporal-gen.yaml` config file so a broad class of activities can share one
set of defaults instead of repeating annotations on every function.

Labels are key/value pairs, mirroring the existing `@memo key value` form.
**Labels apply to activities only** — `@label` on a workflow, query, signal or
update is an error.

*   [Config](file://./temporal-gen.yaml)
*   [Source](file://./labels.go)
*   [Generated](file://./labels_gen.go)

### Defining labels

The config is discovered by walking **up** from the directory being generated,
stopping at the module root (the directory holding `go.mod`). Override it with
`--config <path>`, or turn it off with `--no-config`.

It declares which keys exist, which values each key permits, and the defaults
each value implies:

```yaml
version: 1

# applied to every annotated activity, ahead of any label
defaults:
  start-to-close-timeout: 1m

labels:
  access:
    description: how the activity reaches its data
    values:
      db-only:
        description: read-only database access
        start-to-close-timeout: 30s
        max-retries: 3
      bulk:
        start-to-close-timeout: 1h
        heartbeat-timeout: 1m

  tier:
    values:
      critical:
        max-retries: 870
      best-effort:
        max-retries: 3
```

```go
// @temporal-gen-v2 activity
// @label access db-only
// @label tier critical
// @start-to-close-timeout 10s   // explicit annotation wins over the label
func GetOrg(ctx context.Context, id string) (*app.Org, error)
```

Both the key and the value must be declared in the config. Unknown keys,
unknown values, and setting the same key twice are all hard errors, so a typo
fails loudly rather than silently doing nothing.

### Defining label attributes

Attribute names **are** the annotation names with the leading `@` stripped, so
there is only one vocabulary. Allowed: `task-queue`,
`schedule-to-close-timeout`, `schedule-to-start-timeout`,
`start-to-close-timeout`, `heartbeat-timeout`, `wait-for-cancellation`,
`disable-eager-execution`, `max-retries`, `retry-policy-max-attempts`.

Structural annotations (`@as-wrapper`, `@by-field`, `@local`, `@namespace`,
`@options-callback`, ...) are deliberately **not** settable from a label: they
describe an individual function rather than a class of them. They are also
one-way — the annotation language has no "off" form for a boolean, so a
function could never opt back out of a label that set `@as-wrapper`.

### Precedence

Lowest to highest:

1. `defaults:` block
2. labels, in the order they appear in the source (later wins)
3. explicit `@annotations` on the activity
4. call-site `opts` passed to `Await...`

Distinct label keys usually set disjoint attributes and simply compose; source
order only matters when two keys set the same attribute. The position of a
`@label` line relative to other annotations does not matter — label defaults
are always applied before an activity's own annotations.

Applied labels are recorded as a `// labels: access=bulk, tier=critical`
comment on the generated wrapper so a config change shows up legibly in the
diff.

## Generated Code Structure

For every source file (e.g., `foo.go`), the generator produces `foo_gen.go` containing:

1.  **Type-Safe Wrappers**: `Await...` functions that handle serialization and options.
2.  **Client Structs**: (If queries/updates/signals are present) A `...Client` struct wrapping the Temporal Client.
3.  **Options Patterns**: Functional options for call-time customization (e.g., `With...WorkflowID`).
