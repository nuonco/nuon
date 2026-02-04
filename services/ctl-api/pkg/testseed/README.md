# testseed Package

The `testseed` package provides utilities for seeding test data in ctl-api integration tests. It follows the FX dependency injection pattern used throughout the codebase.

## Overview

This package provides:
- **Seeder**: Helper methods for creating test fixtures (accounts, orgs, apps, installs)
- **Config Fakers**: Faker providers for generating valid config structures

## Usage

### Setting Up in Tests

```go
import (
    "github.com/nuonco/nuon/services/ctl-api/pkg/testseed"
    testseedconfig "github.com/nuonco/nuon/services/ctl-api/pkg/testseed/config"
    "github.com/nuonco/nuon/services/ctl-api/tests"
)

type TestService struct {
    fx.In
    
    DB   *gorm.DB `name:"psql"`
    L    *zap.Logger
    Seed *testseed.Seeder
}

func (s *MyTestSuite) SetupSuite() {
    options := append(
        tests.CtlApiFXOptions(),
        fx.Provide(testseed.New),
        fx.Populate(&s.service),
    )
    
    s.app = fxtest.New(s.T(), options...)
    s.app.RequireStart()
}
```

### Creating Test Fixtures

The seeder provides context-chaining methods that set up the necessary context for ctl-api operations:

```go
func (s *MyTestSuite) TestSomething() {
    ctx := context.Background()
    
    // Create account and set account context
    ctx = s.service.Seed.EnsureAccount(ctx, s.T())
    
    // Create org and set org context
    ctx = s.service.Seed.EnsureOrg(ctx, s.T())
    
    // Create app (requires account and org context)
    testApp := s.service.Seed.EnsureApp(ctx, s.T())
    
    // Create install for the app
    install := s.service.Seed.EnsureInstall(ctx, s.T(), testApp.ID)
    
    // Now you can use these fixtures in your test
    // ctx has both account and org set for ctl-api operations
}
```

### Generating Fake Configs

The `config` subpackage provides faker utilities for generating valid config structures:

```go
import testseedconfig "github.com/nuonco/nuon/services/ctl-api/pkg/testseed/config"

// Get a minimal valid AppConfig
cfg := testseedconfig.GetMinimalAppConfig()

// Customize as needed
cfg.Components = append(cfg.Components, myComponent)
cfg.Actions = append(cfg.Actions, myAction)

// Or get individual configs
sandbox := testseedconfig.GetMinimalSandboxConfig()
runner := testseedconfig.GetMinimalRunnerConfig()
```

## Available Seeder Methods

### EnsureAccount(ctx, t) context.Context
Creates a test account with a fake email and subject ID. Returns context with account ID set.

### EnsureOrg(ctx, t) context.Context
Creates a sandbox organization. Returns context with org ID set.

### EnsureApp(ctx, t) *app.App
Creates a test app. Requires org and account in context. Returns the created app.

### EnsureInstall(ctx, t, appID) *app.Install
Creates a test install for the given app. Requires org and account in context. Returns the created install.

## Available Config Fakers

### GetMinimalAppConfig() *config.AppConfig
Returns a minimal valid AppConfig with:
- Version: "1"
- Sandbox: minimal sandbox config
- Runner: minimal runner config
- Empty Components and Actions slices

### GetMinimalSandboxConfig() *config.AppSandboxConfig
Returns a minimal valid sandbox config using a public repo (no VCS connection required).

### GetMinimalSandboxConfigWithConnectedRepo() *config.AppSandboxConfig
Returns a sandbox config using a connected repo (for testing VCS functionality).

### GetMinimalRunnerConfig() *config.AppRunnerConfig
Returns a minimal valid runner config using "kubernetes" runner type.

## Design Patterns

### Context Chaining
The seeder methods follow a context-chaining pattern where each method returns an updated context:

```go
ctx = seed.EnsureAccount(ctx, t)
ctx = seed.EnsureOrg(ctx, t)
app := seed.EnsureApp(ctx, t)
```

This ensures the context has all required information (account ID, org ID) for subsequent operations.

### Minimal vs Complete Configs
The faker providers generate **minimal valid configs** by default. This means:
- Only required fields are populated
- Optional fields are left as zero values or empty
- Configs will pass validation

This design allows tests to:
1. Start with a valid baseline
2. Customize only what they need
3. Keep test setup code focused and readable

### FX Integration
The seeder integrates with FX dependency injection:
- Uses `fx.In` for dependencies
- Provided via `fx.Provide(testseed.New)`
- Accessed through `fx.Populate`

## Testing the Seeder

The config fakers have their own tests in `config/fakes_test.go`. Run them with:

```bash
go test ./services/ctl-api/pkg/testseed/config/...
```

## Migration from Old Seeder

If you're migrating from the old `internal/pkg/queue/testworker/seed` package:

1. Update imports:
   ```go
   // Old
   import "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/testworker/seed"
   
   // New
   import "github.com/nuonco/nuon/services/ctl-api/pkg/testseed"
   ```

2. Update FX provides:
   ```go
   // Old
   fx.Provide(seed.New)
   
   // New
   fx.Provide(testseed.New)
   ```

3. Method signatures are the same, so no code changes needed!

## Future Extensions

Potential additions:
- More seeder methods: `EnsureComponent`, `EnsureVCSConnection`, etc.
- More config fakers: inputs, permissions, policies, secrets, etc.
- Functional options pattern for customizing generated configs
- Builders for complex nested structures
