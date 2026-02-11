# testseed

Test fixture utilities for ctl-api integration tests.

## Setup

```go
type TestService struct {
    fx.In

    DB     *gorm.DB `name:"psql"`
    Seeder *testseed.Seeder
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

## Creating Fixtures

Each entity type has two or three functions:

- `Build*` -- creates an in-memory struct with fake defaults (no DB)
- `Create*` -- builds and persists to the database (uses account/org from context if set)
- `Ensure*` -- creates, persists, and sets the entity's ID in the context (Account and Org only)

### Setting up an account and org in SetupTest

`EnsureAccount` and `EnsureOrg` return an updated context with the IDs set.
Subsequent `Create*` calls on the seeder will pick up these IDs automatically.

```go
func (s *MyTestSuite) SetupTest() {
    s.ctx = context.Background()
    s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
    s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
}
```

### Creating entities

`Create*` methods build a struct with fake defaults and persist it in one step.
If the context has an account or org ID (set via `Ensure*`), those are used
instead of generating new ones.

```go
// Create an account (standalone, not added to context)
acct := s.service.Seeder.CreateAccount(s.ctx, s.T())

// Create an org (uses account from context for CreatedByID)
org := s.service.Seeder.CreateOrg(s.ctx, s.T())

// Create an app (uses org and account from context)
testApp := s.service.Seeder.CreateApp(s.ctx, s.T())

// Create an install (uses org and account from context)
install := s.service.Seeder.CreateInstall(s.ctx, s.T())
```

### Building without persisting, then saving manually

`Build*` functions create in-memory structs with generated IDs and fake
defaults. You can override fields and save to the DB yourself.

```go
// Build an app in memory
myApp := testseed.BuildApp()

// Override with the test account and org
myApp.CreatedBy = *s.testAcc
myApp.CreatedByID = s.testAcc.ID
myApp.Org = s.testOrg
myApp.OrgID = s.testOrg.ID

// Persist to the database
res := s.service.DB.WithContext(s.ctx).Create(myApp)
require.NoError(s.T(), res.Error)
```
