# temporal-gen-v2

Code generator for Temporal activity and workflow wrappers used by ctl-api. Annotate Go functions with `@temporal-gen`
(and related tags); `go generate` emits `*_gen.go` helpers (`Await*`, activity wrappers, etc.).

## Docs and examples

Full tag reference, YAML options, and worked examples:

- [examples/AGENTS.md](examples/AGENTS.md)
- Sample sources under [examples/](examples/) (`workflow.go`, `activity.go`, `tags.go`, `temporal-gen.yaml`)

## Regenerating

From the monorepo root (or the package that owns `go:generate`):

```bash
go generate ./services/ctl-api/...
go generate ./pkg/workflows/types/executors/...
```

ctl-api wires generation through `services/ctl-api/cmd/gen` and package `go:generate` directives. After changing
annotations or tag defaults, regenerate before relying on the wrappers.
