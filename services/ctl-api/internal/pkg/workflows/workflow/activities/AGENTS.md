# Activities Package - Logging Convention

All activities should use structured logging via `temporalzap` package.

## Basic Pattern

```go
import (
	"go.uber.org/zap"
	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
)

func (a *Activities) MyActivity(ctx context.Context, req *MyRequest) (*MyResult, error) {
	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(zap.String("request_id", req.ID))
	
	l.Info("starting activity")
	// ... logic ...
	if err != nil {
		l.Error("activity failed", zap.Error(err))
		return nil, err
	}
	l.Info("activity completed")
	return result, nil
}
```

## Logging Levels

- **Info**: Activity start, completion, major milestones
- **Debug**: Intermediate steps, type assertions
- **Error**: Failures with `zap.Error(err)`

## Field Types

- `zap.String(key, value)`
- `zap.Int(key, value)`
- `zap.Bool(key, value)`
- `zap.Strings(key, []string)`
- `zap.Error(err)` - Always for errors

## Best Practices

1. Get logger from context with `temporalzap.GetActivityLogger(ctx)`
2. Build context progressively with `.With()`
3. Include relevant IDs for tracing
4. Log errors immediately with context
5. Avoid logging sensitive data (plan contents, secrets)
