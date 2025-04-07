# Profiling Package

This package provides profiling capabilities for Go applications using the standard `net/http/pprof` package.

## Usage

### Environment Variables

The profiler can be configured using the following environment variables:

- `ENABLE_PROFILING`: Set to "true", "1", or "yes" to enable profiling (case insensitive)
- `PROFILING_PORT`: Set to a valid port number to override the default port (6060)

Example:
```bash
# Enable profiling
ENABLE_PROFILING=true ./your-application

# Enable profiling with custom port
ENABLE_PROFILING=true PROFILING_PORT=7070 ./your-application
```

### Programmatic Usage

```go
import "github.com/powertoolsdev/mono/pkg/profiles"

// Using environment variables only
app := fx.New(
    profiles.EnvModule(),
    // other providers...
)

// Or with explicit configuration
app := fx.New(
    profiles.Module(profiles.ProfilerOptions{
        Enabled: true,
        Port: 8080,
    }),
    // other providers...
)
```

## Accessing Profiles

Once the profiler is running, profiles can be accessed at:

```
http://localhost:<port>/debug/pprof/
```

Common endpoints:
- `/debug/pprof/heap` - Heap profile
- `/debug/pprof/profile` - CPU profile
- `/debug/pprof/goroutine` - Goroutines stack traces
- `/debug/pprof/block` - Blocking profile

For detailed usage instructions, see the [Go pprof documentation](https://pkg.go.dev/net/http/pprof).
