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

## Generated Code Structure

For every source file (e.g., `foo.go`), the generator produces `foo_gen.go` containing:

1.  **Type-Safe Wrappers**: `Await...` functions that handle serialization and options.
2.  **Client Structs**: (If queries/updates/signals are present) A `...Client` struct wrapping the Temporal Client.
3.  **Options Patterns**: Functional options for call-time customization (e.g., `With...WorkflowID`).
