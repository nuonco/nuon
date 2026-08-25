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

## Tags (`tags.go`)

Demonstrates `@tag <name>`, which pulls default options from a tag config so a
broad class of activities can share one set of defaults instead of repeating
annotations on every function.

A tag is a single name — there is no value. **Tags apply to activities only** —
`@tag` on a workflow, query, signal or update is an error.

*   [Config](file://./temporal-gen.yaml)
*   [Source](file://./tags.go)
*   [Generated](file://./tags_gen.go)

### Defining tags

The vocabulary comes from one of two places:

**A `temporal-gen.yaml` file**, discovered by walking **up** from the directory
being generated, stopping at the module root (the directory holding `go.mod`).
Override it with `--config <path>`, or turn it off with `--no-config`.

```yaml
version: 1

# applied to every annotated activity, ahead of any tag
defaults:
  start-to-close-timeout: 1m

tags:
  db-read:
    start-to-close-timeout: 30s
    max-retries: 3
  bulk:
    start-to-close-timeout: 1h
    heartbeat-timeout: 1m
  critical:
    max-retries: 870
  best-effort:
    max-retries: 3
```

**Or in Go**, for callers that invoke the generator as a library (see
`services/ctl-api/cmd/gen`). When `Options.Tags` is set it is the *whole*
vocabulary — file discovery and `--config` are skipped, so there is only ever
one source of truth for what a tag means:

```go
temporalgen.Generate(ctx, temporalgen.Options{
	Dir: ".",
	Tags: &tags.Config{
		Defaults: &tags.Attrs{StartToCloseTimeout: "1m"},
		Tags: map[string]*tags.Attrs{
			"db-read": {StartToCloseTimeout: "30s", MaxRetries: generics.ToPtr(3)},
		},
	},
})
```

Either way, usage looks the same:

```go
// @temporal-gen-v2 activity
// @tag db-read
// @tag critical
// @start-to-close-timeout 10s   // explicit annotation wins over the tag
func GetOrg(ctx context.Context, id string) (*app.Org, error)
```

Every tag must be declared. An unknown tag, the same tag twice, a `@tag` with no
config to resolve it against, and `@tag` on a workflow are all hard errors —
**including without `--validate`**, since the non-strict path would otherwise
warn and silently skip the activity, dropping a wrapper you expected to exist.

### Defining tag attributes

Attribute names **are** the annotation names with the leading `@` stripped, so
there is only one vocabulary. Allowed: `task-queue`,
`schedule-to-close-timeout`, `schedule-to-start-timeout`,
`start-to-close-timeout`, `heartbeat-timeout`, `wait-for-cancellation`,
`disable-eager-execution`, `max-retries`, `retry-policy-max-attempts`.

Structural annotations (`@as-wrapper`, `@by-field`, `@local`, `@namespace`,
`@options-callback`, ...) are deliberately **not** settable from a tag: they
describe an individual function rather than a class of them. They are also
one-way — the annotation language has no "off" form for a boolean, so a
function could never opt back out of a tag that set `@as-wrapper`.

### Precedence

Lowest to highest:

1. `defaults:` block
2. tags, in the order they appear in the source (later wins)
3. explicit `@annotations` on the activity
4. call-site `opts` passed to `Await...`

Tags usually set disjoint attributes and simply compose; source order only
matters when two tags set the same attribute. The position of a `@tag` line
relative to other annotations does not matter — tag defaults are always applied
before an activity's own annotations.

Applied tags are recorded as a `// tags: bulk, critical` comment on the
generated wrapper so a config change shows up legibly in the diff.

## Generated Code Structure

For every source file (e.g., `foo.go`), the generator produces `foo_gen.go` containing:

1.  **Type-Safe Wrappers**: `Await...` functions that handle serialization and options.
2.  **Client Structs**: (If queries/updates/signals are present) A `...Client` struct wrapping the Temporal Client.
3.  **Options Patterns**: Functional options for call-time customization (e.g., `With...WorkflowID`).
