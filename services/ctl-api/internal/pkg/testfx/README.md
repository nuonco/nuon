# testfx - Shared Test Configuration for ctl-api Integration Tests

This package provides reusable FX configuration for ctl-api integration tests, eliminating boilerplate code duplication.

## Usage

### Basic Example

```go
package service

import (
    "testing"
    "go.uber.org/fx"
    "go.uber.org/fx/fxtest"
    "github.com/stretchr/testify/suite"

    "github.com/nuonco/nuon/services/ctl-api/internal/pkg/testfx"
)

type MyTestSuite struct {
    suite.Suite
    app     *fxtest.App
    service MyTestService
}

func TestMySuite(t *testing.T) {
    if os.Getenv("INTEGRATION") != "true" {
        t.Skip("INTEGRATION is not set, skipping")
    }
    suite.Run(t, new(MyTestSuite))
}

func (s *MyTestSuite) SetupSuite() {
    gin.SetMode(gin.TestMode)

    // Use shared options + your service-specific setup
    options := append(
        testfx.CommonTestOptions(), // or CommonTestOptionsWithValidator()
        fx.Provide(MyService),
        fx.Populate(&s.service),
    )

    s.app = fxtest.New(s.T(), options...)
    s.app.RequireStart()
}

func (s *MyTestSuite) TearDownSuite() {
    s.app.RequireStop()
}
```

## Available Functions

### `CommonTestOptions() []fx.Option`

Returns FX options with the **custom validator** (includes `entity_name` validation).

**Use this for most tests** - especially tests that validate entity names, app names, component names, etc.

Includes:
- Configuration (`internal.NewConfig`)
- Logging (`log.New`, `dblog.New`)
- External services (loops, GitHub, metrics, propagator)
- Temporal dependencies (full stack)
- Eventloop client
- Databases (PostgreSQL, ClickHouse)
- **Custom validator** (`validatorpkg.New`)
- Auth clients (authz, analytics, account)
- All domain helpers (accounts, vcs, actions, components, apps, runners, installs)
- Endpoint audit
- Database invokers

### `CommonTestOptionsWithValidator() []fx.Option`

Returns FX options with the **standard validator**.

**Deprecated** - Use `CommonTestOptions()` instead unless you specifically need the standard validator.

Same as above but uses `validator.New` instead of the custom validator.

## What's Included

The shared configuration provides all the common dependencies needed for ctl-api integration tests:

- ✅ Configuration management
- ✅ Logging infrastructure
- ✅ Database connections (PostgreSQL + ClickHouse)
- ✅ Temporal workflow engine setup
- ✅ External service clients (GitHub, Loops, Metrics)
- ✅ Auth and authorization clients
- ✅ All domain helpers (cross-domain functionality)
- ✅ Validation (custom or standard)

## What You Still Need to Provide

Each test suite must still provide:

1. **Service under test** - Your specific service via `fx.Provide(MyService)`
2. **Population target** - Your test service struct via `fx.Populate(&s.service)`
3. **Router setup** - Gin router with middleware (if testing HTTP endpoints)
4. **Test data** - Account, org, and other test entities
5. **Cleanup** - TearDown logic

## Migration Example

### Before (100+ lines of boilerplate)

```go
func (s *MyTestSuite) SetupSuite() {
    s.app = fxtest.New(
        s.T(),
        fx.Provide(internal.NewConfig),
        fx.Provide(log.New),
        fx.Provide(dblog.New),
        fx.Provide(loops.New),
        fx.Provide(ghpkg.New),
        fx.Provide(metrics.New),
        fx.Provide(propagator.New),
        fx.Provide(gzip.AsGzip(gzip.New)),
        fx.Provide(largepayload.AsLargePayload(largepayload.New)),
        fx.Provide(signaldb.NewPayloadConverter),
        fx.Provide(dataconverter.New),
        fx.Provide(temporal.New),
        fx.Provide(eventloop.New),
        fx.Provide(psql.AsPSQL(psql.New)),
        fx.Provide(ch.AsCH(ch.New)),
        fx.Provide(validatorpkg.New),
        fx.Provide(authz.New),
        fx.Provide(analytics.New),
        fx.Provide(account.New),
        fx.Provide(accountshelpers.New),
        fx.Provide(vcshelpers.New),
        fx.Provide(actionshelpers.New),
        fx.Provide(componentshelpers.New),
        fx.Provide(appshelpers.New),
        fx.Provide(runnershelpers.New),
        fx.Provide(installshelpers.New),
        fx.Provide(api.NewEndpointAudit),
        fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
        fx.Provide(MyService),
        fx.Populate(&s.service),
    )
    s.app.RequireStart()
}
```

### After (5 lines + your service)

```go
func (s *MyTestSuite) SetupSuite() {
    options := append(
        testfx.CommonTestOptions(),
        fx.Provide(MyService),
        fx.Populate(&s.service),
    )
    s.app = fxtest.New(s.T(), options...)
    s.app.RequireStart()
}
```

## Benefits

- **DRY**: Eliminate 40+ lines of boilerplate per test suite
- **Consistency**: All tests use the same configuration
- **Maintainability**: Update common config in one place
- **Type-safe**: FX dependency injection ensures correctness
- **Easy to extend**: Just append additional options as needed

## Related Packages

- `testdb` - Provides database setup and truncation utilities
- `internal/integration/base_test.go` - Base suite for end-to-end integration tests
