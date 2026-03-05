# Signal Migration Plan: Queue-Compatible Signals

This document outlines the plan for migrating install signals to be queue-compatible.

## Overview

We are migrating from the current flat-namespace `eventloop.Signal` pattern to the new queue-compatible `signal.Signal` pattern. Each signal becomes its own subpackage under `signals/v2/`.

**Key Decisions:**
- Each signal is its own subpackage
- No backwards compatibility needed - leave existing signals in place
- Execute calls activities directly using helper methods
- Tests use the testworker pattern with testify suites

---

## Directory Structure

```
internal/app/installs/signals/
├── signals.go                           # Existing (leave in place)
├── migration_plan.md                    # This file
└── v2/
    ├── provisionrunner/
    │   ├── init.go                      # catalog.Register()
    │   ├── signal.go                    # Signal struct + Type/Validate/Execute
    │   └── signal_test.go               # Tests using testworker pattern
    ├── awaitrunnerhealthy/
    │   ├── init.go
    │   ├── signal.go
    │   └── signal_test.go
    ├── deploycomponentsyncandplan/
    │   ├── init.go
    │   ├── signal.go
    │   └── signal_test.go
    └── ... (one subpackage per signal)
```

---

## Signals to Migrate (32 total)

### Core Operations
| Signal Type | Package Name | Priority |
|-------------|--------------|----------|
| `forgotten` | `forgotten` | Low |
| `restart` | `restart` | Medium |
| `restart-children` | `restartchildren` | Medium |
| `created` | `created` | High |
| `updated` | `updated` | High |

### Action Workflows
| Signal Type | Package Name | Priority |
|-------------|--------------|----------|
| `sync-action-workflow-triggers` | `syncactionworkflowtriggers` | Medium |
| `action-workflow-run` | `actionworkflowrun` | Medium |
| `execute-action-workflow` | `executeactionworkflow` | High |

### Flow Operations
| Signal Type | Package Name | Priority |
|-------------|--------------|----------|
| `execute-flow` | `executeflow` | High |
| `rerun-flow` | `rerunflow` | Medium |

### Dependencies/State
| Signal Type | Package Name | Priority |
|-------------|--------------|----------|
| `poll-dependencies` | `polldependencies` | Medium |
| `generate-state` | `generatestate` | Medium |

### Install Stack
| Signal Type | Package Name | Priority |
|-------------|--------------|----------|
| `generate-install-stack-version` | `generateinstallstackversion` | High |
| `await-install-stack-version-run` | `awaitinstallstackversionrun` | High |
| `update-install-stack-outputs` | `updateinstallstackoutputs` | Medium |

### Runner
| Signal Type | Package Name | Priority |
|-------------|--------------|----------|
| `provision-runner` | `provisionrunner` | High |
| `await-runner-healthy` | `awaitrunnerhealthy` | High |
| `reprovision-runner` | `reprovisionrunner` | Low (deprecated) |

### Sandbox
| Signal Type | Package Name | Priority |
|-------------|--------------|----------|
| `provision-sandbox` | `provisionsandbox` | Medium |
| `provision-sandbox-plan` | `provisionsandboxplan` | High |
| `provision-sandbox-apply-plan` | `provisionsandboxapplyplan` | High |
| `deprovision-sandbox-plan` | `deprovisionsandboxplan` | High |
| `deprovision-sandbox-apply-plan` | `deprovisionsandboxapplyplan` | High |
| `reprovision-sandbox-plan` | `reprovisionsandboxplan` | Medium |
| `reprovision-sandbox-apply-plan` | `reprovisionsandboxapplyplan` | Medium |

### DNS
| Signal Type | Package Name | Priority |
|-------------|--------------|----------|
| `provision-dns` | `provisiondns` | Medium |
| `deprovision-dns` | `deprovisiondns` | Medium |

### Component Deploy
| Signal Type | Package Name | Priority |
|-------------|--------------|----------|
| `execute-deploy-component` | `executedeploycomponent` | High |
| `component-sync-image` | `componentsyncimage` | High |
| `component-deploy-sync-and-plan` | `componentdeploysyncandplan` | High |
| `component-deploy-apply-plan` | `componentdeployapplyplan` | High |
| `component-deploy-plan-only` | `componentdeployplanonly` | Medium |

### Component Teardown
| Signal Type | Package Name | Priority |
|-------------|--------------|----------|
| `execute-teardown-component` | `executeteardowncomponent` | High |
| `component-teardown-sync-and-plan` | `componentteardownsyncandplan` | High |
| `component-teardown-apply-plan` | `componentteardownapplyplan` | High |

### Secrets/Misc
| Signal Type | Package Name | Priority |
|-------------|--------------|----------|
| `sync-secrets` | `syncsecrets` | Medium |
| `workflow-approve-all` | `workflowapproveall` | Low |

---

## Reference Implementation

### Example Signal (from `pkg/queue/example/signal.go`)

```go
package example

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	ExampleSignalType signal.SignalType = "example-signal"
)

func init() {
	catalog.Register(ExampleSignalType, func() signal.Signal {
		return &ExampleSignal{}
	})
}

type ExampleSignal struct {
	Arg1 string `json:"arg_1"`
	Arg2 string `json:"arg_2"`

	isValidated bool
	isExecuted  bool
}

var _ signal.Signal = (*ExampleSignal)(nil)

func (e *ExampleSignal) Validate(ctx workflow.Context) error {
	e.isValidated = true
	return nil
}

func (e *ExampleSignal) Execute(ctx workflow.Context) error {
	e.isExecuted = true
	return nil
}

func (e *ExampleSignal) Type() signal.SignalType {
	return ExampleSignalType
}
```

### Signal Interface (from `pkg/queue/signal/signal.go`)

```go
package signal

import "go.temporal.io/sdk/workflow"

type SignalType string

type Signal interface {
	Type() SignalType

	// workflow handler methods
	Validate(ctx workflow.Context) error
	Execute(ctx workflow.Context) error
}
```

### Catalog Registration (from `pkg/queue/catalog/catalog.go`)

```go
package catalog

import "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"

var SignalCatalog map[signal.SignalType]func() signal.Signal = make(map[signal.SignalType]func() signal.Signal, 0)

func Register(typ signal.SignalType, fn func() signal.Signal) {
	SignalCatalog[typ] = fn
}
```

---

## File Templates

### `init.go`

```go
package {packagename}

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func init() {
	catalog.Register(SignalType, func() signal.Signal {
		return &Signal{}
	})
}
```

### `signal.go`

```go
package {packagename}

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

const SignalType signal.SignalType = "{signal-type}"

type Signal struct {
	// Fields specific to this signal - extracted from current Signal struct
	InstallID string `json:"install_id"`
	// ... other relevant fields
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	// Validation logic
	// - Check required fields
	// - Validate relationships exist in DB via activities
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Business logic - call activities directly via helper methods
	// Example:
	// _, err := activities.AwaitProvisionRunner(ctx, s.InstallID)
	// return err
	return nil
}
```

### `signal_test.go`

```go
package {packagename}

import (
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/workflows/worker"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	handleractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/testworker"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/testworker/seed"
	// ... other required imports from testworker/worker_test.go
)

const defaultNamespace = "default"

type TestService struct {
	fx.In

	DB     *gorm.DB `name:"psql"`
	V      *validator.Validate
	L      *zap.Logger
	Seed   *seed.Seeder
	Client *client.Client
}

type SignalTestSuite struct {
	suite.Suite

	app     *fxtest.App
	service TestService
}

func TestSignalSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(SignalTestSuite))
}

func (s *SignalTestSuite) SetupSuite() {
	s.app = fxtest.New(
		s.T(),
		// Copy FX providers from testworker/worker_test.go
		fx.Provide(internal.NewConfig),
		// ... all the providers ...
		fx.Populate(&s.service),
	)

	s.app.RequireStart()
}

func (s *SignalTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *SignalTestSuite) TestSignalExecutesSuccessfully() {
	ctx := s.service.Seed.EnsureAccount(s.T().Context(), s.T())
	ctx = s.service.Seed.EnsureOrg(ctx, s.T())

	// Create a queue and wait for it to be ready
	queue, err := s.service.Client.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     "test-owner-id",
		OwnerType:   "test-owner-type",
		Namespace:   defaultNamespace,
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.Nil(s.T(), err)
	require.NotNil(s.T(), queue)
	require.Nil(s.T(), s.service.Client.QueueReady(ctx, queue.ID))

	// Enqueue the signal
	enqueueResp, err := s.service.Client.EnqueueSignal(ctx, queue.ID, &Signal{
		InstallID: "test-install-id",
		// ... set other required fields
	})
	require.Nil(s.T(), err)
	require.NotNil(s.T(), enqueueResp)

	// Wait for the signal to complete
	finishedResp, err := s.service.Client.AwaitSignal(ctx, enqueueResp.ID)
	require.Nil(s.T(), err)
	require.NotNil(s.T(), finishedResp)
}

func (s *SignalTestSuite) TestSignalValidationFails() {
	ctx := s.service.Seed.EnsureAccount(s.T().Context(), s.T())
	ctx = s.service.Seed.EnsureOrg(ctx, s.T())

	queue, err := s.service.Client.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     "test-owner-id",
		OwnerType:   "test-owner-type",
		Namespace:   defaultNamespace,
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.Nil(s.T(), err)
	require.Nil(s.T(), s.service.Client.QueueReady(ctx, queue.ID))

	// Enqueue signal with missing required fields
	_, err = s.service.Client.EnqueueSignal(ctx, queue.ID, &Signal{
		// Missing required fields
	})
	// Assert validation error
	require.NotNil(s.T(), err)
}
```

---

## Repeatable Prompt Template

Use this prompt template when migrating each signal:

```
Migrate the `{SIGNAL_TYPE}` signal to queue-compatible format.

**Create subpackage:** `internal/app/installs/signals/v2/{packagename}/`

**Files to create:**
1. `init.go` - Register signal with catalog
2. `signal.go` - Implement Signal struct with Type/Validate/Execute
3. `signal_test.go` - Integration tests using testworker pattern

**Reference files:**
- Current signal definition: `internal/app/installs/signals/signals.go` (look for Operation{Name})
- Activity helpers: `internal/app/installs/worker/activities/`
- Example pattern: `internal/pkg/queue/example/signal.go`
- Test pattern: `internal/pkg/queue/testworker/enqueue_test.go`
- Test setup: `internal/pkg/queue/testworker/worker_test.go` (FX providers)

**Signal-specific fields from current Signal struct:**
{List fields relevant to this signal from signals.go Signal struct}

**Sub-signal structs (if applicable):**
{List any sub-signal structs like DeployComponentSubSignal, SandboxSubSignal, etc.}

**Activity methods to call in Execute():**
{List activity helper methods from worker/activities/ that implement this signal's logic}

**Validation requirements:**
{List any validation rules from the current Validate method or struct tags}

**Instructions:**
1. Create the subpackage directory structure
2. Copy the init.go template and update the package name
3. Create signal.go with only the fields needed for this specific signal
4. Implement Validate() to check required fields and preconditions
5. Implement Execute() to call the appropriate activities
6. Create signal_test.go copying the FX setup from testworker/worker_test.go
7. Add tests for successful execution and validation failures
8. Run tests with `INTEGRATION=true go test ./...`
```

---

## Execution Order

Recommended migration order (high-priority first):

1. **Start with a simple signal** to validate the pattern:
   - `await-runner-healthy` (simple, few dependencies)

2. **Core workflow signals:**
   - `provision-runner`
   - `created`
   - `updated`

3. **Component operations:**
   - `component-deploy-sync-and-plan`
   - `component-deploy-apply-plan`
   - `component-teardown-sync-and-plan`
   - `component-teardown-apply-plan`
   - `component-sync-image`

4. **Sandbox operations:**
   - `provision-sandbox-plan`
   - `provision-sandbox-apply-plan`
   - `deprovision-sandbox-plan`
   - `deprovision-sandbox-apply-plan`

5. **Flow operations:**
   - `execute-flow`
   - `execute-action-workflow`

6. **Remaining signals** in order of priority

---

## Testing Strategy

Each signal test should verify:

1. **Happy path**: Signal enqueues and executes successfully
2. **Validation failures**: Missing required fields are caught
3. **Execute failures**: Activity errors are propagated correctly
4. **Edge cases**: Signal-specific edge cases

Run tests with:
```bash
INTEGRATION=true go test ./internal/app/installs/signals/v2/...
```

---

## Notes

- The existing `signals.go` remains untouched - this is a parallel implementation
- Signals will need to be imported somewhere to trigger `init()` registration
- Consider creating a `v2/all.go` that imports all signal packages for registration
- Activity helper methods already exist in `worker/activities/` - reuse them
