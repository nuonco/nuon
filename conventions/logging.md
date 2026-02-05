# Logging Conventions

This document defines logging standards across the Nuon monorepo. **Never use `fmt.Println` for logging.**

## Quick Reference

| Component | Logger | Import |
|-----------|--------|--------|
| `services/ctl-api` (HTTP) | `*zap.Logger` via FX | `go.uber.org/zap` |
| `services/ctl-api` (Workflows) | `log.WorkflowLogger(ctx)` | `github.com/nuonco/nuon/services/ctl-api/internal/pkg/log` |
| `pkg/` with logger access | `*zap.Logger` passed in | `go.uber.org/zap` |
| `pkg/` init functions | Standard `log` package | `log` |
| `bins/cli` | CLI output helpers | `github.com/nuonco/nuon/bins/cli/internal/ui` |
| `bins/runner` | `*zap.Logger` | `go.uber.org/zap` |

---

## HTTP Services (ctl-api)

Services receive a logger via FX dependency injection:

```go
type Params struct {
    fx.In
    L *zap.Logger  // Logger injected by FX
}

type service struct {
    l *zap.Logger  // Store as lowercase 'l' by convention
}

func New(params Params) *service {
    return &service{l: params.L}
}

func (s *service) MyHandler(ctx *gin.Context) {
    s.l.Info("processing request",
        zap.String("org_id", orgID),
    )
}
```

## Temporal Workflows

Workflows must use `log.WorkflowLogger` to ensure deterministic logging and log stream integration:

```go
import "github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"

func (w *Workflows) MyWorkflow(ctx workflow.Context, req *Request) error {
    l, err := log.WorkflowLogger(ctx)
    if err != nil {
        return errors.Wrap(err, "unable to get workflow logger")
    }
    
    l.Info("starting workflow", zap.String("id", req.ID))
    return nil
}
```

## pkg/ Packages

**Preferred**: Accept a logger as a dependency:

```go
type Client struct {
    l *zap.Logger
}

func New(l *zap.Logger) *Client {
    return &Client{l: l}
}
```

**Init functions**: Use standard `log` package:

```go
import "log"

func init() {
    log.Println("initializing package")
}
```

**No logger access**: Return errors to caller instead of printing:

```go
// ✅ Good - return error
func (s *sync) cleanup(ctx context.Context, id string) error {
    _, err := s.client.Delete(ctx, id)
    return err
}

// ❌ Bad - swallowing error
func (s *sync) cleanup(ctx context.Context, id string) {
    _, err := s.client.Delete(ctx, id)
    if err != nil {
        fmt.Println("failed:", err)  // Never do this
    }
}
```

---

## Log Levels

| Level | When to Use |
|-------|-------------|
| `Debug` | Detailed troubleshooting info, not shown in production |
| `Info` | Normal operations, state changes, significant events |
| `Warn` | Recoverable issues, non-critical failures |
| `Error` | Failures that affect functionality |

## Structured Fields

Always use `zap.*` field functions:

```go
// ✅ Good
s.l.Error("failed to create install",
    zap.Error(err),
    zap.String("org_id", orgID),
    zap.Int("retry_count", retries),
)

// ❌ Bad
s.l.Error(fmt.Sprintf("failed for org %s: %v", orgID, err))
```

**Common field types:**

- `zap.Error(err)` — Always use for errors
- `zap.String("key", val)` — String values
- `zap.Int("key", val)` — Integer values
- `zap.Bool("key", val)` — Boolean values
- `zap.Any("key", val)` — Complex types (use sparingly)
- `zap.Duration("key", dur)` — Time durations

## Field Naming

Use snake_case:

```go
// ✅ Good
zap.String("org_id", orgID)
zap.String("install_id", installID)

// ❌ Bad
zap.String("orgId", orgID)
```

## What NOT to Log

- Secrets, tokens, passwords
- Full request/response bodies (use Debug if needed)
- High-frequency operations in tight loops
- `fmt.Println` — bypasses structured logging entirely
