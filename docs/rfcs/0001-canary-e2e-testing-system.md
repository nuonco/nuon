# RFC 0001: Canary E2E Testing System

**Status:** Draft
**Author:** Robert Bruce
**Created:** 2026-01-29
**Last Updated:** 2026-01-29

## Executive Summary

This RFC proposes a comprehensive end-to-end (E2E) testing system for the Nuon platform using a dedicated Temporal instance to run automated CLI-based tests on a scheduled basis. The system will execute real-world user workflows (org creation, app sync, install deployment) to verify platform functionality and catch regressions early.

## Motivation

### Current State
- Manual testing of CLI workflows is time-consuming and error-prone
- Regressions in critical user flows (org creation, app sync, installs) are discovered by customers
- No automated verification of end-to-end user journeys
- Production incidents could be prevented with continuous E2E testing

### Goals
1. **Automated E2E Testing**: Run comprehensive CLI-based tests that simulate real user workflows
2. **Scheduled Execution**: Multiple test schedules (hourly smoke tests, daily full suite, weekend comprehensive)
3. **Production Safety**: Complete isolation from production workflows via dedicated Temporal instance
4. **Local Development**: Easy to run tests locally against local infrastructure
5. **Observability**: Clear test results, metrics, and alerting on failures

### Non-Goals
- Unit testing (covered by existing test suites)
- Load testing (separate concern)
- Browser-based UI testing (covered by separate e2e service)

## Proposed Solution

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                Production Temporal Cluster                   │
│  (Production workflows: installs, deploys, builds)          │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                Canary Temporal Cluster                       │
│  (E2E test workflows: isolated, scheduled)                  │
│                                                              │
│  Scheduled Workflows:                                        │
│  - quick-smoke (every 6h)                                    │
│  - full-suite (daily 4pm)                                    │
│  - prod-verify (daily 2am)                                   │
│  - weekend-comprehensive (Saturday 6am)                      │
│                                                              │
│  Workers:                                                    │
│  - ctl-api canary worker (namespace: canary)                │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   ctl-api Service                            │
│                                                              │
│  /internal/app/canary/                                       │
│    ├── worker/              (Temporal workflows)            │
│    │   ├── activities/       (Test execution)               │
│    │   │   ├── cli/          (CLI command executor)         │
│    │   │   ├── setup/        (Environment provisioning)     │
│    │   │   ├── validation/   (API/DB verification)          │
│    │   │   └── cleanup/      (Resource cleanup)             │
│    │   ├── provision.go      (Setup canary environment)     │
│    │   ├── e2e_test_suite.go (Test orchestrator)            │
│    │   └── scheduled_test_suite.go (Cron handler)           │
│    └── service/              (HTTP handlers)                │
│        ├── start_schedule.go                                 │
│        ├── stop_schedule.go                                  │
│        └── list_schedules.go                                 │
└─────────────────────────────────────────────────────────────┘
```

## Detailed Design

### 1. Infrastructure Components

#### 1.1 Dedicated Canary Temporal Cluster

**Why Separate?**
- **Isolation**: Runaway canary tests cannot impact production workflows
- **Resource Independence**: Scale canary temporal separately based on test load
- **Safe Experimentation**: Test new Temporal versions without production risk
- **Different Retention**: 7-day retention for tests vs 30+ days for production

**Infrastructure:**
```
infra/temporal-canary/
├── main.tf           # ECS/Fargate deployment
├── rds.tf            # Dedicated PostgreSQL (db.t3.small)
├── elasticache.tf    # Redis for visibility
├── dns.tf            # temporal-canary.nuon.co
└── variables.tf
```

**Services:**
- Temporal Frontend (port 7233)
- Temporal History
- Temporal Matching
- Temporal Worker
- Temporal Web UI (port 8080)

**Cost Estimate:** ~$50-100/month

#### 1.2 Local Development Setup

**docker-compose.canary.yml:**
```yaml
services:
  postgresql-canary:
    image: postgres:14
    ports: ["5433:5432"]  # Different from main postgres

  temporal-canary:
    image: temporalio/auto-setup:1.22.0
    ports: ["7234:7233"]  # Different from main temporal

  temporal-canary-web:
    image: temporalio/web:2.21.3
    ports: ["8234:8080"]  # Different from main temporal web
```

**Setup Command:**
```bash
nctl scripts exec reset-dependencies-canary
```

### 2. Canary Worker Architecture

#### 2.1 Worker Registration

Located in: `services/ctl-api/internal/app/canary/worker/`

**Key Components:**
- `worker.go` - Worker initialization with dedicated canary Temporal client
- `workflows.go` - Workflow registration
- `activities.go` - Activity registration

**Integration Point:**
```go
// services/ctl-api/cmd/worker.go

if (namespace == "all" || namespace == "canary") && !shouldSkipNamespace("canary") {
    providers = append(providers,
        fx.Provide(temporalpkg.NewCanaryClient),  // Dedicated client
        fx.Provide(canaryactivities.New),
        fx.Provide(canaryworker.NewWorkflows),
        fx.Provide(worker.AsWorker(canaryworker.New)),
    )
}
```

#### 2.2 Configuration

**Environment Variables:**
```bash
# Production Temporal (main workflows)
TEMPORAL_HOST=temporal.nuon.co:7233
TEMPORAL_NAMESPACE=general

# Canary Temporal (test workflows)
CANARY_TEMPORAL_HOST=temporal-canary.nuon.co:7233
CANARY_TEMPORAL_NAMESPACE=canary

# Target API for tests
CANARY_API_URL=https://api.nuon.co

# Local Development
TEMPORAL_HOST=localhost:7233
CANARY_TEMPORAL_HOST=localhost:7234
CANARY_API_URL=http://localhost:8081
```

### 3. Test Workflows

#### 3.1 Main Workflows

**ScheduledTestSuite** (scheduled_test_suite.go)
```
Purpose: Orchestrate complete test run (provision → test → cleanup)
Trigger: Cron schedule
Flow:
  1. Provision canary environment (account, token, org)
  2. Execute E2ETestSuite as child workflow
  3. Cleanup canary environment (always runs)
  4. Store results in database
  5. Send notifications on failure
```

**E2ETestSuite** (e2e_test_suite.go)
```
Purpose: Execute test scenarios in parallel or sequence
Flow:
  1. Iterate through requested test scenarios
  2. Execute each scenario (org_lifecycle, app_sync, install_deploy)
  3. Collect results
  4. Return aggregated pass/fail status
```

#### 3.2 Test Scenarios

**Org Lifecycle Test:**
```go
Activities:
  1. RunCLICommand("orgs", "create", "test-org-{timestamp}")
  2. VerifyOrgInDatabase(orgID)
  3. RunCLICommand("orgs", "update", orgID, "--name", "updated")
  4. RunCLICommand("orgs", "delete", orgID)
  5. VerifyOrgDeleted(orgID)
```

**App Sync Test:**
```go
Activities:
  1. CreateTestGitRepo(templateType)
  2. RunCLICommand("apps", "sync", repoPath)
  3. VerifyAppCreated(appID)
  4. VerifyComponentsCreated(appID)
  5. TriggerComponentBuild(componentID)
  6. WaitForBuildComplete(buildID)
```

**Install Deploy Test:**
```go
Activities:
  1. RunCLICommand("installs", "create", appID, "--region", "us-west-2")
  2. WaitForInstallProvisioned(installID)
  3. RunCLICommand("installs", "deploy", installID, componentID)
  4. MonitorDeploymentStatus(installID)
  5. VerifyInstallHealthy(installID)
```

### 4. Scheduled Execution

#### 4.1 Schedule Definitions

**Location:** `pkg/types/workflows/canary/schedules.go`

```go
var (
    // Quick smoke tests - every 6 hours
    QuickSmokeSchedule = TestSchedule{
        Name:        "quick-smoke",
        CronExpr:    "0 */6 * * *",
        Scenarios:   []string{"org_lifecycle", "app_sync"},
        Environment: "stage",
        SandboxMode: true,
        Timeout:     30 * time.Minute,
    }

    // Full test suite - daily at 4pm UTC
    FullSuiteSchedule = TestSchedule{
        Name:        "full-suite",
        CronExpr:    "0 16 * * *",
        Scenarios:   []string{"org_lifecycle", "app_sync", "install_deploy", "component_build"},
        Environment: "stage",
        SandboxMode: false,
        Timeout:     2 * time.Hour,
    }

    // Production verification - daily at 2am UTC
    ProductionVerifySchedule = TestSchedule{
        Name:        "prod-verify",
        CronExpr:    "0 2 * * *",
        Scenarios:   []string{"org_lifecycle"},
        Environment: "prod",
        SandboxMode: false,
        Timeout:     15 * time.Minute,
    }

    // Weekend comprehensive - Saturdays at 6am UTC
    WeekendComprehensiveSchedule = TestSchedule{
        Name:        "weekend-comprehensive",
        CronExpr:    "0 6 * * 6",
        Scenarios:   []string{
            "org_lifecycle",
            "app_sync",
            "install_deploy",
            "component_build",
            "release_flow",
            "multi_cloud",
        },
        Environment: "stage",
        SandboxMode: false,
        Timeout:     4 * time.Hour,
    }
)
```

#### 4.2 Schedule Management API

**Endpoints:**
- `POST /v1/general/canary/schedules/start` - Start a schedule
- `POST /v1/general/canary/schedules/stop` - Stop a schedule
- `GET /v1/general/canary/schedules` - List all schedules with status

**nuonctl Commands:**
```bash
nctl canary schedule start quick-smoke
nctl canary schedule stop quick-smoke
nctl canary schedule list
nctl canary schedule start-all
nctl canary schedule stop-all
```

### 5. CLI Executor

#### 5.1 Activity Implementation

**Location:** `services/ctl-api/internal/app/canary/worker/activities/cli/executor.go`

```go
type Executor struct {
    apiURL   string
    apiToken string
}

func (e *Executor) RunCLICommand(ctx context.Context, args []string) (*CommandResult, error) {
    cmd := exec.CommandContext(ctx, "nuon", args...)

    cmd.Env = []string{
        fmt.Sprintf("NUON_API_URL=%s", e.apiURL),
        fmt.Sprintf("NUON_API_TOKEN=%s", e.apiToken),
        "PATH=" + os.Getenv("PATH"),
    }

    output, err := cmd.CombinedOutput()

    return &CommandResult{
        Output:   string(output),
        ExitCode: getExitCode(err),
        Duration: elapsed,
        Success:  err == nil,
    }, nil
}
```

#### 5.2 Environment Provisioning

**Setup Activities:**
1. `CreateCanaryAccount` - Create AccountTypeCanary account
2. `CreateAPIToken` - Generate API token for CLI auth
3. `CreateGitHubInstall` - Setup GitHub integration (if needed)
4. `CreateTestOrg` - Create organization for tests

**Cleanup Activities:**
1. `DeleteTestOrg` - Hard delete test org
2. `DeleteCanaryAccount` - Remove test account
3. `CleanupResources` - Remove any leftover resources

### 6. Data Storage

#### 6.1 Database Schema

```sql
-- Store scheduled test run results
CREATE TABLE canary_scheduled_runs (
    id              VARCHAR(26) PRIMARY KEY,
    schedule_name   VARCHAR(100) NOT NULL,
    canary_id       VARCHAR(26) NOT NULL,
    started_at      TIMESTAMP NOT NULL,
    completed_at    TIMESTAMP,
    duration_ms     INT,
    environment     VARCHAR(20),
    sandbox_mode    BOOLEAN,
    total_tests     INT,
    passed_tests    INT,
    failed_tests    INT,
    results         JSONB,
    created_at      TIMESTAMP DEFAULT NOW(),

    INDEX idx_schedule_started (schedule_name, started_at DESC),
    INDEX idx_environment (environment, started_at DESC),
    INDEX idx_failed (failed_tests, started_at DESC)
);

-- Track individual test scenario results
CREATE TABLE canary_test_results (
    id              VARCHAR(26) PRIMARY KEY,
    run_id          VARCHAR(26) REFERENCES canary_scheduled_runs(id),
    scenario        VARCHAR(100) NOT NULL,
    passed          BOOLEAN NOT NULL,
    error_message   TEXT,
    duration_ms     INT,
    details         JSONB,
    created_at      TIMESTAMP DEFAULT NOW(),

    INDEX idx_run_id (run_id),
    INDEX idx_scenario_passed (scenario, passed, created_at DESC)
);

-- Track schedule execution history
CREATE TABLE canary_schedule_history (
    id              VARCHAR(26) PRIMARY KEY,
    schedule_name   VARCHAR(100) NOT NULL,
    cron_expr       VARCHAR(50),
    started_at      TIMESTAMP NOT NULL,
    stopped_at      TIMESTAMP,
    started_by      VARCHAR(100),
    stopped_by      VARCHAR(100),

    INDEX idx_schedule (schedule_name, started_at DESC)
);
```

#### 6.2 Result Storage Activity

```go
func (a *Activities) StoreScheduledTestResults(ctx context.Context, result *ScheduledTestSuiteResponse) error {
    run := &CanaryScheduledRun{
        ID:           result.CanaryID,
        ScheduleName: result.ScheduleName,
        StartedAt:    result.StartTime,
        CompletedAt:  result.EndTime,
        DurationMs:   int(result.Duration.Milliseconds()),
        Environment:  result.Environment,
        SandboxMode:  result.SandboxMode,
        TotalTests:   result.TotalTests,
        PassedTests:  result.PassedTests,
        FailedTests:  result.FailedTests,
        Results:      result.Results,
    }

    return a.db.Create(run).Error
}
```

### 7. Monitoring & Alerting

#### 7.1 Metrics

**DataDog Metrics:**
```go
// Test execution metrics
canary.scheduled.total_tests (gauge)
canary.scheduled.passed_tests (gauge)
canary.scheduled.failed_tests (gauge)
canary.scheduled.duration (timing)
canary.scheduled.failures (count)

// Tags:
// - schedule: quick-smoke, full-suite, etc.
// - environment: local, stage, prod
// - scenario: org_lifecycle, app_sync, etc.
```

#### 7.2 Alerting

**Slack Notifications:**
- Send to `#canary-alerts` on test failures
- Include failed test details and Temporal UI link
- Daily summary of all test runs

**DataDog Monitors:**
- Alert if `canary.scheduled.failed_tests > 0` for 2 consecutive runs
- Alert if scheduled test doesn't run within expected window
- Alert if canary Temporal cluster is down

**PagerDuty:**
- Only for production verification failures (critical path)
- Weekend comprehensive failures are non-critical (Slack only)

### 8. Local Development Workflow

#### 8.1 Setup

```bash
# 1. Start main infrastructure
nctl scripts exec reset-dependencies

# 2. Start canary infrastructure
nctl scripts exec reset-dependencies-canary

# 3. Start ctl-api API server
nctl services dev --dev ctl-api

# 4. Start main workers (skip canary)
cd services/ctl-api
go run main.go worker --namespace all --skip canary

# 5. Start canary worker (separate terminal)
cd services/ctl-api
go run main.go worker --namespace canary
```

#### 8.2 Running Tests Locally

```bash
# Trigger one-time test run
curl -X POST http://localhost:8081/v1/general/provision-canary \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"sandbox_mode": true}'

# Or via nuonctl
nctl canary test --env local --scenarios org_lifecycle,app_sync

# Start a schedule locally
nctl canary schedule start quick-smoke

# View results
open http://localhost:8234  # Canary Temporal Web UI
```

### 9. Production Deployment

#### 9.1 Infrastructure

```bash
# Deploy canary Temporal cluster
cd infra/temporal-canary
terraform init
terraform plan -var-file=../vars/prod.tfvars
terraform apply -var-file=../vars/prod.tfvars

# Outputs:
# - temporal_canary_host = temporal-canary.nuon.co
# - temporal_canary_web = temporal-canary-web.nuon.co
```

#### 9.2 Worker Deployment

**Option A: Dedicated ECS Service**
```hcl
resource "aws_ecs_service" "canary_worker" {
  name            = "canary-worker"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.canary_worker.arn
  desired_count   = 1

  # Environment variables point to canary temporal
}
```

**Option B: Add to Existing ctl-api Workers**
```bash
# Run with namespace flag
ctl-api worker --namespace canary
```

#### 9.3 Enable Schedules

```bash
# infra/scripts/enable-canary-schedules.sh

curl -X POST https://api.nuon.co/v1/general/canary/schedules/start \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"schedule_name": "quick-smoke"}'

curl -X POST https://api.nuon.co/v1/general/canary/schedules/start \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"schedule_name": "full-suite"}'
```

## Implementation Phases

### Phase 1: Foundation (Week 1)
- [ ] Create `infra/temporal-canary/` Terraform module
- [ ] Deploy canary Temporal to staging
- [ ] Create `docker-compose.canary.yml` for local dev
- [ ] Create `services/ctl-api/internal/app/canary/` directory structure
- [ ] Implement canary worker with dedicated Temporal client
- [ ] Add canary namespace to `cmd/worker.go`
- [ ] Test worker startup and connection

**Deliverable:** Canary worker can connect to dedicated Temporal instance

### Phase 2: Basic Test Execution (Week 2)
- [ ] Implement CLI executor activity
- [ ] Implement setup activities (create account, token)
- [ ] Implement cleanup activities (delete resources)
- [ ] Create simple org lifecycle test
- [ ] Test end-to-end: provision → test → cleanup

**Deliverable:** Can execute basic org CRUD test via Temporal workflow

### Phase 3: Test Scenarios (Week 3)
- [ ] Implement org lifecycle test (full CRUD)
- [ ] Implement app sync test
- [ ] Implement install deployment test
- [ ] Add validation activities (verify via API/DB)
- [ ] Create E2ETestSuite orchestrator workflow

**Deliverable:** Complete test suite covering major user workflows

### Phase 4: Scheduled Execution (Week 4)
- [ ] Define schedule configurations
- [ ] Implement ScheduledTestSuite workflow
- [ ] Create schedule management API endpoints
- [ ] Implement result storage in database
- [ ] Add nuonctl schedule commands

**Deliverable:** Tests run automatically on schedule

### Phase 5: Observability (Week 5)
- [ ] Add DataDog metrics
- [ ] Implement Slack notifications
- [ ] Create DataDog dashboards
- [ ] Set up alerts for failures
- [ ] Add result viewing commands

**Deliverable:** Full visibility into test execution and failures

### Phase 6: Production Rollout (Week 6)
- [ ] Deploy canary Temporal to production
- [ ] Deploy canary worker to production
- [ ] Enable quick-smoke schedule
- [ ] Monitor for 1 week
- [ ] Enable full-suite schedule
- [ ] Enable weekend-comprehensive schedule

**Deliverable:** Canary tests running in production

## Success Metrics

### Test Coverage
- ✅ Org CRUD operations
- ✅ App sync workflow
- ✅ Install deployment
- ✅ Component builds
- ⏳ Release flow (Phase 7)
- ⏳ Multi-cloud testing (Phase 7)

### Execution Metrics
- **Quick Smoke**: <5 minutes, 4x daily
- **Full Suite**: <30 minutes, 1x daily
- **Production Verify**: <10 minutes, 1x daily
- **Weekend Comprehensive**: <2 hours, 1x weekly

### Reliability
- <5% false positive rate
- >95% schedule adherence (tests run on time)
- Zero production impact from canary tests

### Detection
- Catch regressions within 6 hours (quick-smoke frequency)
- Identify breaking changes before customer impact
- Validate major releases with comprehensive suite

## Alternatives Considered

### Alternative 1: Shared Temporal Namespace
**Approach:** Use main Temporal with separate `canary` namespace

**Pros:**
- Simpler infrastructure (one Temporal cluster)
- Lower operational cost
- Single Temporal UI

**Cons:**
- Shared fate (Temporal issues affect both)
- Resource contention with production
- Cannot test Temporal upgrades independently

**Decision:** Rejected. Production isolation is critical.

### Alternative 2: Separate Canary Service
**Approach:** Create `services/workers-canary/` as standalone service

**Pros:**
- Complete isolation from ctl-api
- Independent deployment

**Cons:**
- More operational overhead
- Code duplication
- Separate infrastructure management
- Doesn't follow current pattern (workers in ctl-api)

**Decision:** Rejected. Current pattern is to consolidate workers in ctl-api.

### Alternative 3: GitHub Actions for E2E Tests
**Approach:** Run CLI tests in GitHub Actions on every PR

**Pros:**
- Simple setup
- Integrated with CI/CD

**Cons:**
- Cannot run on schedule independent of commits
- GitHub Actions environment differs from production
- Harder to manage long-running tests
- Less control over test environment

**Decision:** Rejected. Temporal provides better orchestration for scheduled, long-running tests.

## Open Questions

### 1. Test Data Cleanup
**Question:** How aggressively should we clean up test data?

**Options:**
- A. Delete immediately after each test
- B. Keep for 24 hours for debugging
- C. Hard delete vs soft delete

**Recommendation:** Keep for 24 hours with scheduled cleanup job. Use hard delete for test data.

### 2. Production Testing Scope
**Question:** How much testing should run against production?

**Options:**
- A. None (too risky)
- B. Read-only tests only
- C. Minimal CRUD with sandbox mode
- D. Full test suite in sandbox mode

**Recommendation:** Start with minimal org lifecycle test in sandbox mode (creates/deletes test org). Expand carefully.

### 3. Test Parallelization
**Question:** Should test scenarios run in parallel or sequence?

**Options:**
- A. All parallel (fastest, but higher resource usage)
- B. All sequential (safest, but slower)
- C. Configurable per schedule

**Recommendation:** Sequential for now, add parallel option in Phase 7.

### 4. Failure Handling
**Question:** Should one test failure fail the entire suite?

**Options:**
- A. Fail fast (stop on first failure)
- B. Continue on failure (run all tests)
- C. Configurable per schedule

**Recommendation:** Continue on failure to gather complete results. Mark overall run as failed if any test fails.

### 5. CLI Binary Source
**Question:** Where does the canary worker get the `nuon` CLI binary?

**Options:**
- A. Build from source in worker container
- B. Download latest release from GitHub
- C. Use specific pinned version
- D. Build and include in worker image

**Recommendation:** Include pinned CLI version in worker Docker image. Update via deployment, not at runtime.

## Security Considerations

### API Token Management
- Use dedicated service account with minimal permissions
- Rotate tokens regularly (90-day rotation)
- Store tokens in AWS Secrets Manager
- Never log token values

### Test Account Isolation
- Use `AccountTypeCanary` to mark test accounts
- Restrict permissions to sandbox operations
- Separate AWS account for canary orgs (orgs-stage)
- Automated cleanup of abandoned test resources

### Production Testing Safeguards
- Sandbox mode required for production tests
- Resource limits on test orgs (max installs, max components)
- Rate limiting on test execution
- Manual approval for new production test scenarios

## References

- [Temporal Cron Workflows](https://docs.temporal.io/workflows#cron-schedule)
- [Nuon RBAC System](./AGENTS.md#account--organization-permission-system)
- [ctl-api Worker Architecture](../services/ctl-api/AGENTS.md#workers-pattern)
- [Existing Canary Implementation](../pkg/types/workflows/canary/)

## Changelog

- **2026-01-29**: Initial RFC draft
- **[TBD]**: Phase 1 implementation complete
- **[TBD]**: Production rollout

---

## Appendix A: Full Directory Structure

```
services/ctl-api/internal/app/canary/
├── worker/
│   ├── worker.go                       # Worker initialization
│   ├── workflows.go                    # Workflow registration
│   ├── provision.go                    # Provision workflow
│   ├── deprovision.go                  # Deprovision workflow
│   ├── e2e_test_suite.go              # Test orchestrator
│   ├── scheduled_test_suite.go         # Cron handler
│   └── activities/
│       ├── activities.go               # Activities struct
│       ├── setup/
│       │   ├── create_account.go
│       │   ├── create_api_token.go
│       │   └── create_github_install.go
│       ├── cli/
│       │   ├── executor.go
│       │   ├── org_tests.go
│       │   ├── app_tests.go
│       │   └── install_tests.go
│       ├── validation/
│       │   ├── api_check.go
│       │   └── db_check.go
│       ├── cleanup/
│       │   ├── delete_org.go
│       │   └── delete_account.go
│       ├── metrics.go
│       └── notifications.go
├── service/
│   ├── start_schedule.go
│   ├── stop_schedule.go
│   └── list_schedules.go
└── helpers/
    └── result_storage.go

pkg/types/workflows/canary/
├── schedules.go
├── provision.go
├── deprovision.go
├── e2e_test_suite.go
└── scheduled_test_suite.go

infra/temporal-canary/
├── main.tf
├── variables.tf
├── outputs.tf
├── rds.tf
├── elasticache.tf
├── ecs.tf
├── dns.tf
└── README.md

bins/nuonctl/cmd/
└── canary_schedule.go
```

## Appendix B: Example Test Output

```
$ nctl canary schedule start full-suite
✅ Started canary schedule: full-suite

$ nctl canary results --schedule full-suite --limit 1

Canary Test Run: full-suite
Started:  2026-01-29 16:00:00 UTC
Completed: 2026-01-29 16:24:32 UTC
Duration: 24m32s
Environment: stage

Results:
  ✅ org_lifecycle (2m15s)
  ✅ app_sync (8m42s)
  ✅ install_deploy (12m18s)
  ❌ component_build (1m17s)
     Error: Build timed out after 15 minutes

Summary: 3/4 tests passed (75%)

View Details: https://temporal-canary-web.nuon.co/namespaces/canary/workflows/canary_full-suite_1738166400
```
