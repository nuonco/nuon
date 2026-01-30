# RFC 0001: Canary E2E Testing System

**Status:** Draft
**Author:** Robert Bruce
**Created:** 2026-01-29
**Last Updated:** 2026-01-29 (Integrated Playwright Dashboard Testing)

## Executive Summary

This RFC proposes a comprehensive end-to-end (E2E) testing system for the Nuon platform using a dedicated Temporal instance to run automated tests on a scheduled basis. The system will execute real-world user workflows through both CLI commands and browser automation (Playwright) to verify platform functionality and catch regressions early. All test execution and results are tracked in Temporal workflows with centralized storage, metrics, and alerting.

## Motivation

### Current State
- Manual testing of CLI workflows is time-consuming and error-prone
- Regressions in critical user flows (org creation, app sync, installs) are discovered by customers
- No automated verification of end-to-end user journeys
- Production incidents could be prevented with continuous E2E testing

### Goals
1. **Automated E2E Testing**: Run comprehensive tests simulating real user workflows via CLI and browser
2. **Dashboard UX Testing**: Verify critical user flows through actual browser interactions (Playwright)
3. **Scheduled Execution**: Multiple test schedules (hourly smoke tests, daily full suite, weekend comprehensive)
4. **Production Safety**: Complete isolation from production workflows via dedicated Temporal instance
5. **Local Development**: Easy to run tests locally against local infrastructure
6. **Observability**: Clear test results, metrics, and alerting on failures
7. **Centralized Tracking**: All test types (CLI + Dashboard) recorded in unified system

### Non-Goals
- Unit testing (covered by existing test suites in dashboard-ui and ctl-api)
- Load testing / Performance testing (separate concern)
- Visual regression testing (future consideration)
- Complete browser compatibility matrix (focus on Chrome initially)

## Proposed Solution

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                Production Temporal Cluster                   │
│  (Production workflows: installs, deploys, builds)          │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                Canary Temporal Cluster                       │
│  (E2E test workflows: CLI + Playwright, isolated, scheduled)│
│                                                              │
│  Scheduled Workflows:                                        │
│  - quick-smoke (every 6h) - CLI + critical Playwright       │
│  - full-suite (daily 4pm) - All tests                       │
│  - prod-verify (daily 2am) - CLI only                       │
│  - weekend-comprehensive (Saturday 6am) - Full coverage     │
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
│    │   │   ├── dashboard/    (Playwright test runner)       │
│    │   │   ├── setup/        (Environment provisioning)     │
│    │   │   ├── validation/   (API/DB verification)          │
│    │   │   └── cleanup/      (Resource cleanup)             │
│    │   ├── provision.go      (Setup canary environment)     │
│    │   ├── e2e_test_suite.go (CLI + Playwright orchestrator)│
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
│                     # Used for both persistence and visibility
├── dns.tf            # temporal-canary.nuon.co
└── variables.tf
```

**Database Configuration:**
- **PostgreSQL RDS**: Single database handles both persistence and visibility
- **No Redis/ElastiCache needed**: Modern Temporal (1.20+) uses PostgreSQL for visibility
- **Simpler architecture**: One less service to maintain and monitor

**Services:**
- Temporal Frontend (port 7233)
- Temporal History
- Temporal Matching
- Temporal Worker
- Temporal Web UI (port 8080)

**Cost Estimate:** ~$40-80/month (PostgreSQL RDS + ECS tasks)

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

**CLI Test Scenarios:**

*Org Lifecycle Test:*
```go
Activities:
  1. RunCLICommand("orgs", "create", "test-org-{timestamp}")
  2. VerifyOrgInDatabase(orgID)
  3. RunCLICommand("orgs", "update", orgID, "--name", "updated")
  4. RunCLICommand("orgs", "delete", orgID)
  5. VerifyOrgDeleted(orgID)
```

*App Sync Test:*
```go
Activities:
  1. CreateTestGitRepo(templateType)
  2. RunCLICommand("apps", "sync", repoPath)
  3. VerifyAppCreated(appID)
  4. VerifyComponentsCreated(appID)
  5. TriggerComponentBuild(componentID)
  6. WaitForBuildComplete(buildID)
```

*Install Deploy Test:*
```go
Activities:
  1. RunCLICommand("installs", "create", appID, "--region", "us-west-2")
  2. WaitForInstallProvisioned(installID)
  3. RunCLICommand("installs", "deploy", installID, componentID)
  4. MonitorDeploymentStatus(installID)
  5. VerifyInstallHealthy(installID)
```

**Dashboard Test Scenarios (Playwright):**

*Dashboard Org Creation:*
```go
Activities:
  1. RunPlaywrightTest("e2e/org-creation.spec.ts")
     - Navigate to dashboard
     - Click "Create Organization"
     - Fill form and submit
     - Verify success toast
     - Verify redirect to org page
  2. VerifyOrgInDatabase(orgID)
```

*Dashboard App Detail Page:*
```go
Activities:
  1. RunCLICommand("apps", "sync", repoPath)  // Setup via CLI
  2. RunPlaywrightTest("e2e/app-detail.spec.ts")
     - Navigate to app detail page
     - Verify components displayed
     - Click "Build" button
     - Verify build started notification
```

*Dashboard Install Deploy:*
```go
Activities:
  1. RunPlaywrightTest("e2e/install-deploy.spec.ts")
     - Navigate to install detail page
     - Click "Deploy" button
     - Monitor deployment status in UI
     - Verify deployment completes
```

*Dashboard Error Handling:*
```go
Activities:
  1. RunPlaywrightTest("e2e/error-handling.spec.ts")
     - Trigger API error scenarios
     - Verify user-friendly error messages
     - Test form validation
     - Verify network error handling
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
        Scenarios:   []string{
            "org_lifecycle",              // CLI
            "app_sync",                   // CLI
            "dashboard_org_creation",     // Playwright
        },
        Environment: "stage",
        SandboxMode: true,
        Timeout:     30 * time.Minute,
    }

    // Full test suite - daily at 4pm UTC
    FullSuiteSchedule = TestSchedule{
        Name:        "full-suite",
        CronExpr:    "0 16 * * *",
        Scenarios:   []string{
            "org_lifecycle",              // CLI
            "app_sync",                   // CLI
            "install_deploy",             // CLI
            "component_build",            // CLI
            "dashboard_org_creation",     // Playwright
            "dashboard_app_detail",       // Playwright
            "dashboard_install_deploy",   // Playwright
        },
        Environment: "stage",
        SandboxMode: false,
        Timeout:     2 * time.Hour,
    }

    // Production verification - daily at 2am UTC
    ProductionVerifySchedule = TestSchedule{
        Name:        "prod-verify",
        CronExpr:    "0 2 * * *",
        Scenarios:   []string{
            "org_lifecycle",              // CLI only in prod
        },
        Environment: "prod",
        SandboxMode: false,
        Timeout:     15 * time.Minute,
    }

    // Weekend comprehensive - Saturdays at 6am UTC
    WeekendComprehensiveSchedule = TestSchedule{
        Name:        "weekend-comprehensive",
        CronExpr:    "0 6 * * 6",
        Scenarios:   []string{
            "org_lifecycle",              // CLI
            "app_sync",                   // CLI
            "install_deploy",             // CLI
            "component_build",            // CLI
            "release_flow",               // CLI
            "multi_cloud",                // CLI
            "dashboard_org_creation",     // Playwright
            "dashboard_app_detail",       // Playwright
            "dashboard_install_deploy",   // Playwright
            "dashboard_error_handling",   // Playwright
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

### 6. Dashboard Testing with Playwright

#### 6.1 Overview

Dashboard E2E tests use Playwright to automate real browser interactions and verify the user experience. These tests execute within the same canary Temporal workflows as CLI tests, providing unified tracking and reporting.

**Why Playwright:**
- Tests actual user interactions (clicks, forms, navigation)
- Catches UI/UX bugs that API tests miss
- Verifies visual elements and error messages
- Industry standard (used by GitHub, Microsoft)
- Can capture screenshots/videos on failure

#### 6.2 Playwright Setup

**Location:** `services/dashboard-ui/e2e/`

**Configuration:**
```typescript
// services/dashboard-ui/playwright.config.ts

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,  // Sequential for canary tests
  retries: 1,
  workers: 1,
  timeout: 60000,

  reporter: [
    ['json', { outputFile: 'playwright-report/results.json' }],
    ['html', { outputFolder: 'playwright-report/html' }],
  ],

  use: {
    baseURL: process.env.DASHBOARD_URL || 'http://localhost:4000',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },

  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
})
```

**Authentication Fixture:**
```typescript
// services/dashboard-ui/e2e/fixtures/auth.ts

import { test as base } from '@playwright/test'

export const test = base.extend<{ authenticatedPage: Page }>({
  authenticatedPage: async ({ page }, use) => {
    // Auth token set by canary activity
    const authToken = process.env.CANARY_AUTH_TOKEN
    const orgId = process.env.CANARY_ORG_ID

    await page.context().addCookies([
      {
        name: 'appSession',
        value: authToken,
        domain: 'localhost',
        path: '/',
      },
    ])

    await use(page)
  },
})
```

#### 6.3 Example Playwright Test

```typescript
// services/dashboard-ui/e2e/org-creation.spec.ts

import { test, expect } from './fixtures/auth'

test.describe('Organization Creation Flow', () => {
  test('should create organization via dashboard UI', async ({ authenticatedPage: page }) => {
    await page.goto('/')

    // Click create organization button
    await page.click('button:has-text("Create Organization")')

    // Fill in organization name
    const orgName = `test-org-${Date.now()}`
    await page.fill('input[name="name"]', orgName)

    // Submit form
    await page.click('button[type="submit"]:has-text("Create")')

    // Wait for success toast
    await expect(page.locator('text=Organization created')).toBeVisible({
      timeout: 10000
    })

    // Verify redirect to org page
    await expect(page).toHaveURL(/\/orgs\/org[a-z0-9]+/)

    // Verify org name appears in page
    await expect(page.locator(`text=${orgName}`)).toBeVisible()
  })

  test('should validate required fields', async ({ authenticatedPage: page }) => {
    await page.goto('/')
    await page.click('button:has-text("Create Organization")')

    // Try to submit without filling name
    await page.click('button[type="submit"]:has-text("Create")')

    // Should show validation error
    await expect(page.locator('text=Organization name is required')).toBeVisible()
  })
})
```

#### 6.4 Canary Worker Integration

**Playwright Activity:**
```go
// services/ctl-api/internal/app/canary/worker/activities/dashboard/run_playwright.go

package dashboard

import (
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "time"

    "go.temporal.io/sdk/activity"
)

type PlaywrightRunner struct {
    dashboardPath string
    s3Client      *s3.Client
    s3Bucket      string
}

type RunPlaywrightRequest struct {
    TestFile     string            `json:"test_file"`
    DashboardURL string            `json:"dashboard_url"`
    AuthToken    string            `json:"auth_token"`
    OrgID        string            `json:"org_id"`
    Environment  map[string]string `json:"environment"`
    Timeout      time.Duration     `json:"timeout"`
}

type PlaywrightTestResult struct {
    TestFile      string           `json:"test_file"`
    Success       bool             `json:"success"`
    TestsPassed   int              `json:"tests_passed"`
    TestsFailed   int              `json:"tests_failed"`
    TestsSkipped  int              `json:"tests_skipped"`
    Duration      time.Duration    `json:"duration"`
    Tests         []PlaywrightTest `json:"tests"`
    Screenshots   []string         `json:"screenshots"`  // S3 URLs
    Videos        []string         `json:"videos"`       // S3 URLs
    TraceFiles    []string         `json:"trace_files"` // S3 URLs
    ErrorMessage  string           `json:"error_message,omitempty"`
}

type PlaywrightTest struct {
    Title    string        `json:"title"`
    Status   string        `json:"status"`
    Duration time.Duration `json:"duration"`
    Error    *TestError    `json:"error,omitempty"`
}

// @temporal-gen activity
// @activity-queue "default"
func (r *PlaywrightRunner) RunPlaywrightTest(ctx context.Context, req *RunPlaywrightRequest) (*PlaywrightTestResult, error) {
    logger := activity.GetLogger(ctx)
    startTime := time.Now()

    // Prepare environment variables
    env := []string{
        fmt.Sprintf("DASHBOARD_URL=%s", req.DashboardURL),
        fmt.Sprintf("CANARY_AUTH_TOKEN=%s", req.AuthToken),
        fmt.Sprintf("CANARY_ORG_ID=%s", req.OrgID),
        "NODE_ENV=test",
        "CI=true",
    }

    // Execute Playwright tests
    cmd := exec.CommandContext(ctx, "npx", "playwright", "test", req.TestFile)
    cmd.Dir = r.dashboardPath
    cmd.Env = append(os.Environ(), env...)

    output, err := cmd.CombinedOutput()

    // Parse JSON report
    reportPath := filepath.Join(r.dashboardPath, "playwright-report", "results.json")
    report, _ := r.parsePlaywrightReport(reportPath)

    result := &PlaywrightTestResult{
        TestFile:     req.TestFile,
        TestsPassed:  report.Stats.Passed,
        TestsFailed:  report.Stats.Failed,
        TestsSkipped: report.Stats.Skipped,
        Success:      report.Stats.Failed == 0 && err == nil,
        Duration:     time.Since(startTime),
    }

    // Upload artifacts to S3 on failure
    if !result.Success {
        result.Screenshots, _ = r.uploadArtifacts(ctx, "playwright-report/**/*.png", "screenshots")
        result.Videos, _ = r.uploadArtifacts(ctx, "playwright-report/**/*.webm", "videos")
        result.TraceFiles, _ = r.uploadArtifacts(ctx, "playwright-report/**/*.zip", "traces")
    }

    return result, nil
}
```

#### 6.5 Workflow Integration

**E2ETestSuite with Playwright:**
```go
func (w *Workflows) E2ETestSuite(ctx workflow.Context, req *canary.E2ETestSuiteRequest) (*canary.E2ETestSuiteResponse, error) {
    // ... setup canary environment ...

    // Run CLI tests
    if contains(req.TestScenarios, "org_lifecycle") {
        // ... execute CLI org test ...
    }

    // Run Playwright tests
    playwrightTests := []struct {
        Scenario string
        TestFile string
    }{
        {"dashboard_org_creation", "e2e/org-creation.spec.ts"},
        {"dashboard_app_detail", "e2e/app-detail.spec.ts"},
        {"dashboard_install_deploy", "e2e/install-deploy.spec.ts"},
    }

    for _, test := range playwrightTests {
        if !contains(req.TestScenarios, test.Scenario) {
            continue
        }

        response.TotalTests++

        var playwrightResult *PlaywrightTestResult
        err := workflow.ExecuteActivity(ctx, "RunPlaywrightTest", &RunPlaywrightRequest{
            TestFile:     test.TestFile,
            DashboardURL: req.DashboardURL,
            AuthToken:    setupResult.APIToken,
            OrgID:        setupResult.OrgID,
            Timeout:      3 * time.Minute,
        }).Get(ctx, &playwrightResult)

        // Convert to canary TestResult
        testResult := &canary.TestResult{
            Scenario: test.Scenario,
            Passed:   playwrightResult.Success,
            Duration: playwrightResult.Duration,
            Details: map[string]interface{}{
                "test_file":     playwrightResult.TestFile,
                "tests_passed":  playwrightResult.TestsPassed,
                "tests_failed":  playwrightResult.TestsFailed,
                "screenshots":   playwrightResult.Screenshots,
                "videos":        playwrightResult.Videos,
                "traces":        playwrightResult.TraceFiles,
            },
        }

        response.Results[test.Scenario] = testResult
    }

    // ... cleanup and store results ...
}
```

#### 6.6 Artifact Storage

**S3 Upload on Failures:**
- Screenshots: `s3://canary-artifacts/screenshots/YYYY-MM-DD/*.png`
- Videos: `s3://canary-artifacts/videos/YYYY-MM-DD/*.webm`
- Traces: `s3://canary-artifacts/traces/YYYY-MM-DD/*.zip`

**Trace Viewer:**
Playwright traces can be viewed at `https://trace.playwright.dev/?trace=<S3_URL>`

#### 6.7 Test Scenarios

**Critical UI Flows:**
1. **Organization Creation** (`e2e/org-creation.spec.ts`)
   - Create org via dashboard form
   - Verify org appears in org switcher
   - Test form validation

2. **App Detail Page** (`e2e/app-detail.spec.ts`)
   - Navigate to app after CLI sync
   - View component list
   - Trigger component build from UI

3. **Install Deployment** (`e2e/install-deploy.spec.ts`)
   - View install detail page
   - Click deploy button
   - Monitor deployment status in UI

4. **Error Handling** (`e2e/error-handling.spec.ts`)
   - API error displays user-friendly message
   - Form validation shows inline errors
   - Network errors handled gracefully

#### 6.8 Notifications on Failure

**Slack Alert with Artifacts:**
```go
func (n *Notifier) SendPlaywrightFailureNotification(ctx context.Context, result *PlaywrightTestResult) error {
    message := fmt.Sprintf(`
🔴 *Playwright Test Failed*

*Test File:* %s
*Failed Tests:* %d/%d
*Duration:* %s

*Debug Artifacts:*
📸 <https://s3.amazonaws.com/%s|Screenshot>
🎬 <https://s3.amazonaws.com/%s|Video>
🔍 <https://trace.playwright.dev/?trace=%s|Trace>

<https://temporal-canary-web.nuon.co/workflows/%s|View in Temporal UI>
    `,
        result.TestFile,
        result.TestsFailed,
        result.TestsPassed+result.TestsFailed,
        result.Duration,
        result.Screenshots[0],
        result.Videos[0],
        result.TraceFiles[0],
        workflowID,
    )

    return n.slackClient.SendMessage(ctx, "#canary-alerts", message)
}
```

### 7. Handling Long-Running Tests

#### 7.1 Timeout Architecture

**Multi-Level Timeout Configuration:**

```go
// Workflow-level timeout (overall test suite)
ScheduledTestSuiteWorkflowTimeout = 4 * time.Hour  // Max for weekend-comprehensive

// Activity-level timeouts (individual test scenarios)
ActivityTimeouts = map[string]time.Duration{
    // CLI tests
    "org_lifecycle":     5 * time.Minute,
    "app_sync":         15 * time.Minute,
    "install_deploy":   30 * time.Minute,  // Long-running
    "component_build":  20 * time.Minute,  // Long-running

    // Playwright tests
    "dashboard_org_creation":    3 * time.Minute,
    "dashboard_app_detail":      5 * time.Minute,
    "dashboard_install_deploy": 10 * time.Minute,  // Long-running
}

// Activity heartbeat intervals (for progress reporting)
ActivityHeartbeatInterval = 30 * time.Second
```

**Workflow Configuration:**
```go
func (w *Workflows) E2ETestSuite(ctx workflow.Context, req *E2ETestSuiteRequest) (*E2ETestSuiteResponse, error) {
    // Set workflow timeout
    ctx = workflow.WithWorkflowTimeout(ctx, 4*time.Hour)

    // Configure activity options with timeouts + heartbeats
    activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
        StartToCloseTimeout: 30 * time.Minute,  // Default activity timeout
        HeartbeatTimeout:    2 * time.Minute,   // Expect heartbeat every 2min
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 2,  // Retry once on transient failures
        },
    })

    // Execute activities with scenario-specific timeouts
    for _, scenario := range req.TestScenarios {
        timeout := getScenarioTimeout(scenario)
        scenarioCtx := workflow.WithActivityOptions(activityCtx, workflow.ActivityOptions{
            StartToCloseTimeout: timeout,
        })

        // Execute test with proper timeout context
        err := workflow.ExecuteActivity(scenarioCtx, "RunTest", scenario).Get(scenarioCtx, &result)
    }
}
```

#### 7.2 Activity Heartbeats for Progress Tracking

Long-running activities report progress via heartbeats to prevent timeout:

```go
// CLI activity with heartbeat
func (e *Executor) RunCLICommand(ctx context.Context, req *CommandRequest) (*CommandResult, error) {
    // For long-running commands (install_deploy, component_build)
    if isLongRunning(req.Command) {
        go func() {
            ticker := time.NewTicker(30 * time.Second)
            defer ticker.Stop()

            for {
                select {
                case <-ctx.Done():
                    return
                case <-ticker.C:
                    // Report progress to Temporal
                    activity.RecordHeartbeat(ctx, map[string]interface{}{
                        "status": "in_progress",
                        "elapsed": time.Since(startTime),
                    })
                }
            }
        }()
    }

    // Execute command
    output, err := cmd.CombinedOutput()
    return &CommandResult{...}, nil
}

// Playwright activity with heartbeat
func (r *PlaywrightRunner) RunPlaywrightTest(ctx context.Context, req *RunPlaywrightRequest) (*PlaywrightTestResult, error) {
    // Start heartbeat goroutine
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                // Read current test progress from Playwright output
                progress := parsePlaywrightProgress()
                activity.RecordHeartbeat(ctx, map[string]interface{}{
                    "tests_completed": progress.Completed,
                    "tests_total": progress.Total,
                    "current_test": progress.CurrentTest,
                })
            }
        }
    }()

    // Execute Playwright test suite
    cmd := exec.CommandContext(ctx, "npx", "playwright", "test", req.TestFile)
    output, err := cmd.CombinedOutput()
    return parseResults(output), nil
}
```

#### 7.3 Parallel Execution for Performance

**Sequential vs Parallel Execution:**

```go
type E2ETestSuiteRequest struct {
    TestScenarios   []string
    ExecutionMode   string  // "sequential" or "parallel"
    MaxParallelism  int     // Max concurrent tests
}

func (w *Workflows) E2ETestSuite(ctx workflow.Context, req *E2ETestSuiteRequest) (*E2ETestSuiteResponse, error) {
    if req.ExecutionMode == "parallel" {
        return w.executeTestsInParallel(ctx, req)
    }
    return w.executeTestsSequentially(ctx, req)
}

// Parallel execution reduces total runtime
func (w *Workflows) executeTestsInParallel(ctx workflow.Context, req *E2ETestSuiteRequest) (*E2ETestSuiteResponse, error) {
    maxParallel := req.MaxParallelism
    if maxParallel == 0 {
        maxParallel = 3  // Default: run 3 tests concurrently
    }

    // Use selector to limit concurrency
    selector := workflow.NewSelector(ctx)
    inFlight := 0
    testQueue := make([]string, len(req.TestScenarios))
    copy(testQueue, req.TestScenarios)

    results := make(map[string]*TestResult)

    for len(testQueue) > 0 || inFlight > 0 {
        // Start new tests up to maxParallel
        for inFlight < maxParallel && len(testQueue) > 0 {
            scenario := testQueue[0]
            testQueue = testQueue[1:]

            future := workflow.ExecuteActivity(activityCtx, "RunTest", scenario)
            selector.AddFuture(future, func(f workflow.Future) {
                var result TestResult
                f.Get(ctx, &result)
                results[scenario] = &result
                inFlight--
            })
            inFlight++
        }

        // Wait for at least one to complete
        selector.Select(ctx)
    }

    return aggregateResults(results), nil
}
```

**Parallelization Strategy:**
- **Quick smoke tests**: Sequential (5-10 min total, simple)
- **Full suite**: Parallel with maxParallelism=3 (30 min → 15 min)
- **Weekend comprehensive**: Parallel with maxParallelism=5 (4 hours → 1.5 hours)

#### 7.4 Child Workflows for Very Long Tests

For extremely long tests (multi-hour deployments), use child workflows:

```go
// Parent workflow delegates to child workflow
func (w *Workflows) E2ETestSuite(ctx workflow.Context, req *E2ETestSuiteRequest) (*E2ETestSuiteResponse, error) {
    for _, scenario := range req.TestScenarios {
        if isVeryLongRunning(scenario) {
            // Use child workflow for isolation
            childWorkflowOptions := workflow.ChildWorkflowOptions{
                WorkflowID: fmt.Sprintf("canary-test-%s-%d", scenario, time.Now().Unix()),
                WorkflowExecutionTimeout: 2 * time.Hour,
            }
            childCtx := workflow.WithChildOptions(ctx, childWorkflowOptions)

            var childResult TestResult
            err := workflow.ExecuteChildWorkflow(childCtx, "LongRunningTestWorkflow", scenario).Get(childCtx, &childResult)
            results[scenario] = &childResult
        } else {
            // Use activity for normal tests
            err := workflow.ExecuteActivity(activityCtx, "RunTest", scenario).Get(activityCtx, &result)
            results[scenario] = &result
        }
    }
}
```

**Benefits of Child Workflows:**
- Independent timeout configuration
- Can be queried/cancelled separately
- Better visibility in Temporal UI
- Isolated retry policies

#### 7.5 Graceful Degradation & Timeouts

**Timeout Handling Strategy:**

```go
func (w *Workflows) E2ETestSuite(ctx workflow.Context, req *E2ETestSuiteRequest) (*E2ETestSuiteResponse, error) {
    response := &E2ETestSuiteResponse{
        Results: make(map[string]*TestResult),
    }

    for _, scenario := range req.TestScenarios {
        var result TestResult

        err := workflow.ExecuteActivity(activityCtx, "RunTest", scenario).Get(activityCtx, &result)

        if temporal.IsTimeoutError(err) {
            // Record timeout as test failure, but continue
            response.Results[scenario] = &TestResult{
                Scenario: scenario,
                Passed:   false,
                Error:    fmt.Sprintf("Test timed out after %s", getScenarioTimeout(scenario)),
                Duration: getScenarioTimeout(scenario),
            }
            response.FailedTests++
        } else if err != nil {
            // Other errors also continue
            response.Results[scenario] = &TestResult{
                Scenario: scenario,
                Passed:   false,
                Error:    err.Error(),
            }
            response.FailedTests++
        } else {
            // Success
            response.Results[scenario] = &result
            if result.Passed {
                response.PassedTests++
            } else {
                response.FailedTests++
            }
        }

        response.TotalTests++
    }

    // Always run cleanup, even if tests timeout
    workflow.ExecuteActivity(cleanupCtx, "Cleanup").Get(cleanupCtx, nil)

    return response, nil
}
```

**Key Points:**
- **Continue on timeout**: One slow test doesn't block others
- **Record timeout details**: Store which test timed out and after how long
- **Always cleanup**: Use defer-like pattern to ensure resource cleanup
- **Alert on timeouts**: Separate alerting for timeout vs failure

#### 7.6 Resource Management for Long Tests

**Playwright Resource Limits:**

```typescript
// playwright.config.ts
export default defineConfig({
  timeout: 60000,         // Per-test timeout (1 minute)
  globalTimeout: 600000,  // Global suite timeout (10 minutes)

  workers: 1,             // Sequential execution (avoid resource contention)

  use: {
    // Browser context timeout
    navigationTimeout: 30000,
    actionTimeout: 10000,

    // Cleanup after each test
    trace: 'retain-on-failure',
    video: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },

  // Retry failed tests once (transient browser issues)
  retries: 1,
})
```

**Docker Container Limits:**

```yaml
# docker-compose.canary.yml (for local development)
services:
  canary-worker:
    image: canary-worker:latest
    deploy:
      resources:
        limits:
          cpus: '2.0'       # Limit CPU for Playwright browser processes
          memory: 4G        # Playwright can be memory-intensive
        reservations:
          memory: 2G
    environment:
      - NODE_OPTIONS=--max-old-space-size=3072  # Node.js heap limit
```

#### 7.7 Monitoring Long-Running Tests

**DataDog Metrics for Duration Tracking:**

```go
// Emit metrics for long-running test detection
canary.test.duration (timing)
canary.test.timeout (count)
canary.test.heartbeat_count (count)

// Tags:
// - scenario: install_deploy, component_build, etc.
// - result: success, failure, timeout
// - duration_bucket: 0-5m, 5-15m, 15-30m, 30m+
```

**Temporal UI Visibility:**

Tests show real-time progress:
- Activity heartbeat details visible in Temporal UI
- Current test progress (e.g., "3/5 tests completed")
- Elapsed time for in-progress activities
- Timeline view shows parallel execution

#### 7.8 How Temporal Handles Extremely Long Tests

Temporal's architecture is uniquely suited for long-running tests (hours, days, or even longer). Here's why:

**Event Sourcing & Durability:**

Temporal uses event sourcing - every state change is persisted as an event in the database:

```
Workflow Execution: canary_full-suite_1738166400
Events:
  1. WorkflowExecutionStarted (2026-01-29 16:00:00)
  2. ActivityScheduled: provision_environment
  3. ActivityStarted: provision_environment
  4. ActivityCompleted: provision_environment (result: {account_id, token})
  5. ActivityScheduled: install_deploy
  6. ActivityStarted: install_deploy
  7. ActivityHeartbeat: install_deploy (progress: "deploying component 1/5")
  8. ActivityHeartbeat: install_deploy (progress: "deploying component 2/5")
  ... (workflow continues for hours)
```

**Key Benefits:**

1. **Worker Crashes Are Safe:**
   - If a worker process crashes during a 2-hour test, Temporal automatically reschedules the activity on another worker
   - The activity can resume from last heartbeat (if implemented with checkpointing)
   - No test progress is lost

2. **Workflow State Persists Forever:**
   - Workflow history stored in PostgreSQL (our canary Temporal database)
   - Can survive server restarts, deployments, maintenance windows
   - A test that starts on Friday can continue through the weekend

3. **Timeouts Are Configurable Per Activity:**
   ```go
   // Different activities have different time requirements
   StartToCloseTimeout: 2 * time.Hour  // Max time for activity
   ScheduleToCloseTimeout: 3 * time.Hour  // Including queue time
   HeartbeatTimeout: 2 * time.Minute  // Expect heartbeat every 2 min
   ```

**Worker Restart Example:**

```
Timeline:
16:00 - Start install_deploy test (expected duration: 30 min)
16:15 - Worker crashes (or deployment rollout)
16:15 - Temporal detects worker is gone (via heartbeat timeout)
16:15 - Temporal reschedules activity on different worker
16:16 - New worker picks up activity, resumes from checkpoint
16:30 - Activity completes successfully
```

The test continues seamlessly without user intervention.

**Continue-As-New for Extremely Long Workflows:**

For workflows that need to run indefinitely (e.g., daily scheduled tests running for months):

```go
func (w *Workflows) ScheduledTestSuite(ctx workflow.Context, req *ScheduledTestSuiteRequest) error {
    // Run test suite
    result := w.executeTests(ctx, req)

    // Store results
    workflow.ExecuteActivity(ctx, "StoreResults", result).Get(ctx, nil)

    // Wait for next schedule (using Temporal cron)
    // Temporal automatically handles "continue-as-new" for cron workflows
    // Event history is reset, preventing unbounded growth

    return nil  // Temporal will start new execution on next cron trigger
}
```

**Continue-As-New** prevents event history from growing unbounded over months/years of scheduled runs.

**Querying Running Workflows:**

While a test is running, you can query its state:

```bash
# Query current test progress
temporal workflow query \
  --workflow-id canary_full-suite_1738166400 \
  --query-type GetCurrentProgress

# Response:
{
  "total_tests": 7,
  "completed_tests": 4,
  "current_test": "install_deploy",
  "current_test_progress": "deploying component 3/5",
  "elapsed_time": "25m42s"
}
```

**Signals for Runtime Control:**

You can send signals to running workflows to modify behavior:

```go
// Define signal handler in workflow
func (w *Workflows) E2ETestSuite(ctx workflow.Context, req *E2ETestSuiteRequest) (*E2ETestSuiteResponse, error) {
    pauseChannel := workflow.GetSignalChannel(ctx, "pause")
    resumeChannel := workflow.GetSignalChannel(ctx, "resume")
    cancelChannel := workflow.GetSignalChannel(ctx, "cancel")

    for _, scenario := range req.TestScenarios {
        // Check for pause signal
        var shouldPause bool
        pauseChannel.ReceiveAsync(&shouldPause)
        if shouldPause {
            // Wait for resume signal
            resumeChannel.Receive(ctx, nil)
        }

        // Check for cancel signal
        var shouldCancel bool
        cancelChannel.ReceiveAsync(&shouldCancel)
        if shouldCancel {
            return nil, errors.New("workflow cancelled by user")
        }

        // Execute test
        workflow.ExecuteActivity(ctx, "RunTest", scenario).Get(ctx, &result)
    }
}
```

```bash
# Pause a running test suite (e.g., production issue detected)
temporal workflow signal \
  --workflow-id canary_full-suite_1738166400 \
  --signal-name pause

# Resume later
temporal workflow signal \
  --workflow-id canary_full-suite_1738166400 \
  --signal-name resume
```

**Activity Checkpointing for Resume:**

For extremely long activities (e.g., multi-hour Playwright test suites), implement checkpointing:

```go
func (r *PlaywrightRunner) RunPlaywrightTest(ctx context.Context, req *RunPlaywrightRequest) (*PlaywrightTestResult, error) {
    // Check for checkpoint from previous attempt
    checkpoint := loadCheckpoint(req.TestFile)

    // Skip already-completed tests
    testsToRun := req.Tests
    if checkpoint != nil {
        testsToRun = filterCompletedTests(req.Tests, checkpoint.CompletedTests)
        logger.Info("Resuming from checkpoint", "completed", len(checkpoint.CompletedTests))
    }

    results := checkpoint.Results

    for _, test := range testsToRun {
        // Run test
        result := executePlaywrightTest(test)
        results = append(results, result)

        // Save checkpoint after each test
        saveCheckpoint(req.TestFile, Checkpoint{
            CompletedTests: append(checkpoint.CompletedTests, test),
            Results: results,
        })

        // Send heartbeat with checkpoint info
        activity.RecordHeartbeat(ctx, map[string]interface{}{
            "completed": len(results),
            "total": len(req.Tests),
            "checkpoint": true,
        })
    }

    return aggregateResults(results), nil
}
```

**Comparison to Other Orchestration Tools:**

| Feature | Temporal | GitHub Actions | Jenkins | AWS Step Functions |
|---------|----------|----------------|---------|-------------------|
| Max execution time | **Days/months** | 6 hours | Unlimited* | 1 year (with heartbeats) |
| Survives restarts | **Yes** | No | No* | Yes |
| Pause/resume | **Yes** | No | Limited | No |
| Progress queries | **Yes** | Limited | Via API | Limited |
| Automatic retry | **Yes** | Yes | Limited | Yes |
| Heartbeat support | **Yes** | No | No | Yes |
| Event history | **Full** | Logs only | Logs only | Full |

*Jenkins jobs survive restarts only if agent stays up; pipeline state not durable

**Why This Matters for Canary Tests:**

1. **Weekend Comprehensive Suite** (4 hours): Can run uninterrupted even if workers restart
2. **Production Deploys** (30+ min): Safe to run even during deployments
3. **Multi-environment Tests**: Can wait for resources, retry on transient failures
4. **Debug Failed Tests**: Full event history shows exactly what happened and when
5. **Scheduled Execution**: Cron schedules work reliably for months without manual intervention

**Real-World Example:**

```
Scenario: Weekend comprehensive test suite starts Saturday 6am

06:00 - Workflow starts, provisions canary environment (5 min)
06:05 - Starts running tests in parallel
06:30 - Worker node scaled down (cost optimization)
06:31 - Temporal detects worker failure, reschedules activities
06:32 - Activities resume on remaining workers
08:00 - All tests complete, cleanup runs
08:05 - Results stored, notifications sent
08:05 - Workflow completes successfully

Total duration: 2h 5min
Survived: 1 worker failure
User intervention: None
Test data: Fully preserved
```

This is why Temporal is the gold standard for durable, long-running workflow orchestration.

### 8. Data Storage

#### 8.1 Database Schema

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

-- Track canary test environments for cleanup
CREATE TABLE canary_environments (
    id              VARCHAR(26) PRIMARY KEY,
    run_id          VARCHAR(26) REFERENCES canary_scheduled_runs(id),
    account_id      VARCHAR(26),
    org_id          VARCHAR(26),
    api_token       VARCHAR(100),       -- Encrypted token
    github_install_id VARCHAR(26),
    environment     VARCHAR(20),
    provisioned_at  TIMESTAMP NOT NULL,
    cleaned_up_at   TIMESTAMP,
    cleanup_status  VARCHAR(20),        -- 'pending', 'completed', 'failed'
    cleanup_error   TEXT,
    created_at      TIMESTAMP DEFAULT NOW(),

    INDEX idx_run_id (run_id),
    INDEX idx_cleanup_status (cleanup_status, provisioned_at DESC),
    INDEX idx_org_id (org_id)
);

-- Track Temporal workflow execution details for debugging
CREATE TABLE canary_workflow_executions (
    id              VARCHAR(26) PRIMARY KEY,
    run_id          VARCHAR(26) REFERENCES canary_scheduled_runs(id),
    workflow_id     VARCHAR(255) NOT NULL,
    workflow_type   VARCHAR(100),       -- 'ScheduledTestSuite', 'E2ETestSuite'
    run_id_temporal VARCHAR(255),       -- Temporal's run ID (UUID)
    namespace       VARCHAR(50),        -- 'canary'
    started_at      TIMESTAMP NOT NULL,
    completed_at    TIMESTAMP,
    status          VARCHAR(20),        -- 'running', 'completed', 'failed', 'timeout', 'cancelled'
    error_message   TEXT,
    created_at      TIMESTAMP DEFAULT NOW(),

    INDEX idx_workflow_id (workflow_id),
    INDEX idx_run_id (run_id),
    INDEX idx_status (status, started_at DESC),
    UNIQUE (workflow_id, run_id_temporal)
);

-- Track test artifacts (screenshots, videos, traces) for Playwright tests
CREATE TABLE canary_test_artifacts (
    id              VARCHAR(26) PRIMARY KEY,
    test_result_id  VARCHAR(26) REFERENCES canary_test_results(id),
    artifact_type   VARCHAR(20) NOT NULL,  -- 'screenshot', 'video', 'trace', 'log'
    s3_bucket       VARCHAR(100),
    s3_key          VARCHAR(500),
    s3_url          TEXT,                  -- Full S3 URL
    file_size_bytes BIGINT,
    content_type    VARCHAR(100),
    uploaded_at     TIMESTAMP NOT NULL,
    expires_at      TIMESTAMP,             -- Optional: for automatic cleanup
    created_at      TIMESTAMP DEFAULT NOW(),

    INDEX idx_test_result_id (test_result_id),
    INDEX idx_artifact_type (artifact_type, uploaded_at DESC),
    INDEX idx_expires_at (expires_at)
);

-- Track flaky tests for reliability analysis
CREATE TABLE canary_test_reliability (
    id              VARCHAR(26) PRIMARY KEY,
    scenario        VARCHAR(100) NOT NULL,
    test_type       VARCHAR(20),        -- 'cli', 'playwright'
    window_start    TIMESTAMP NOT NULL,
    window_end      TIMESTAMP NOT NULL,
    total_runs      INT NOT NULL,
    passed_runs     INT NOT NULL,
    failed_runs     INT NOT NULL,
    timeout_runs    INT NOT NULL,
    pass_rate       DECIMAL(5,2),       -- Calculated: passed_runs / total_runs * 100
    flakiness_score DECIMAL(5,2),       -- Custom score for flaky test detection
    created_at      TIMESTAMP DEFAULT NOW(),

    INDEX idx_scenario (scenario, window_start DESC),
    INDEX idx_pass_rate (pass_rate, scenario),
    INDEX idx_flakiness_score (flakiness_score DESC),
    UNIQUE (scenario, test_type, window_start)
);

-- Track active schedule configurations
CREATE TABLE canary_schedules (
    id              VARCHAR(26) PRIMARY KEY,
    schedule_name   VARCHAR(100) NOT NULL UNIQUE,
    cron_expr       VARCHAR(50) NOT NULL,
    test_scenarios  JSONB NOT NULL,     -- Array of scenario names
    environment     VARCHAR(20),
    sandbox_mode    BOOLEAN,
    execution_mode  VARCHAR(20),        -- 'sequential', 'parallel'
    max_parallelism INT,
    timeout_minutes INT,
    is_active       BOOLEAN DEFAULT true,
    last_run_at     TIMESTAMP,
    next_run_at     TIMESTAMP,
    temporal_schedule_id VARCHAR(255),  -- Temporal's schedule ID
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW(),
    created_by_id   VARCHAR(26),
    updated_by_id   VARCHAR(26),

    INDEX idx_schedule_name (schedule_name),
    INDEX idx_is_active (is_active, next_run_at),
    INDEX idx_last_run_at (last_run_at DESC)
);
```

#### 8.2 Result Storage Activity

```go
func (a *Activities) StoreScheduledTestResults(ctx context.Context, result *ScheduledTestSuiteResponse) error {
    // Use transaction to ensure all data is stored atomically
    return a.db.Transaction(func(tx *gorm.DB) error {
        // 1. Store main test run record
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
        if err := tx.Create(run).Error; err != nil {
            return err
        }

        // 2. Store individual test results
        for scenario, testResult := range result.TestResults {
            tr := &CanaryTestResult{
                ID:           ulid.Make().String(),
                RunID:        result.CanaryID,
                Scenario:     scenario,
                Passed:       testResult.Passed,
                ErrorMessage: testResult.Error,
                DurationMs:   int(testResult.Duration.Milliseconds()),
                Details:      testResult.Details,
            }
            if err := tx.Create(tr).Error; err != nil {
                return err
            }

            // 3. Store Playwright artifacts if present
            if artifacts, ok := testResult.Details["artifacts"].(map[string]interface{}); ok {
                if screenshots, ok := artifacts["screenshots"].([]string); ok {
                    for _, url := range screenshots {
                        artifact := &CanaryTestArtifact{
                            ID:            ulid.Make().String(),
                            TestResultID:  tr.ID,
                            ArtifactType:  "screenshot",
                            S3URL:         url,
                            S3Bucket:      "canary-artifacts",
                            S3Key:         extractS3Key(url),
                            UploadedAt:    time.Now(),
                            ExpiresAt:     time.Now().Add(30 * 24 * time.Hour), // 30 days
                        }
                        if err := tx.Create(artifact).Error; err != nil {
                            return err
                        }
                    }
                }

                // Similar for videos and traces...
            }
        }

        // 4. Store workflow execution details
        workflow := &CanaryWorkflowExecution{
            ID:             ulid.Make().String(),
            RunID:          result.CanaryID,
            WorkflowID:     result.WorkflowID,
            WorkflowType:   "ScheduledTestSuite",
            RunIDTemporal:  result.TemporalRunID,
            Namespace:      "canary",
            StartedAt:      result.StartTime,
            CompletedAt:    result.EndTime,
            Status:         result.Status,
            ErrorMessage:   result.Error,
        }
        if err := tx.Create(workflow).Error; err != nil {
            return err
        }

        // 5. Update schedule last_run_at timestamp
        if err := tx.Model(&CanarySchedule{}).
            Where("schedule_name = ?", result.ScheduleName).
            Update("last_run_at", result.StartTime).Error; err != nil {
            return err
        }

        return nil
    })
}
```

#### 8.3 Environment Tracking for Cleanup

```go
func (a *Activities) TrackCanaryEnvironment(ctx context.Context, req *TrackEnvironmentRequest) error {
    env := &CanaryEnvironment{
        ID:              ulid.Make().String(),
        RunID:           req.RunID,
        AccountID:       req.AccountID,
        OrgID:           req.OrgID,
        APIToken:        encrypt(req.APIToken),  // Encrypted
        GitHubInstallID: req.GitHubInstallID,
        Environment:     req.Environment,
        ProvisionedAt:   time.Now(),
        CleanupStatus:   "pending",
    }

    return a.db.Create(env).Error
}

func (a *Activities) CleanupCanaryEnvironment(ctx context.Context, envID string) error {
    var env CanaryEnvironment
    if err := a.db.First(&env, "id = ?", envID).Error; err != nil {
        return err
    }

    // Perform cleanup operations
    if err := a.deleteOrg(ctx, env.OrgID); err != nil {
        a.db.Model(&env).Updates(map[string]interface{}{
            "cleanup_status": "failed",
            "cleanup_error":  err.Error(),
        })
        return err
    }

    if err := a.deleteAccount(ctx, env.AccountID); err != nil {
        a.db.Model(&env).Updates(map[string]interface{}{
            "cleanup_status": "failed",
            "cleanup_error":  err.Error(),
        })
        return err
    }

    // Mark as cleaned up
    return a.db.Model(&env).Updates(map[string]interface{}{
        "cleanup_status": "completed",
        "cleaned_up_at":  time.Now(),
    }).Error
}
```

#### 8.4 Flaky Test Detection

```go
// Background job that runs daily to calculate test reliability
func CalculateTestReliability(db *gorm.DB) error {
    windowStart := time.Now().Add(-7 * 24 * time.Hour)  // Last 7 days
    windowEnd := time.Now()

    // Get all unique test scenarios
    var scenarios []struct {
        Scenario string
        TestType string
    }

    db.Raw(`
        SELECT DISTINCT scenario,
               CASE
                   WHEN scenario LIKE 'dashboard_%' THEN 'playwright'
                   ELSE 'cli'
               END as test_type
        FROM canary_test_results
        WHERE created_at >= ?
    `, windowStart).Scan(&scenarios)

    for _, s := range scenarios {
        var stats struct {
            TotalRuns   int
            PassedRuns  int
            FailedRuns  int
            TimeoutRuns int
        }

        db.Raw(`
            SELECT
                COUNT(*) as total_runs,
                SUM(CASE WHEN passed = true THEN 1 ELSE 0 END) as passed_runs,
                SUM(CASE WHEN passed = false AND error_message NOT LIKE '%timeout%' THEN 1 ELSE 0 END) as failed_runs,
                SUM(CASE WHEN error_message LIKE '%timeout%' THEN 1 ELSE 0 END) as timeout_runs
            FROM canary_test_results
            WHERE scenario = ? AND created_at >= ? AND created_at <= ?
        `, s.Scenario, windowStart, windowEnd).Scan(&stats)

        passRate := float64(stats.PassedRuns) / float64(stats.TotalRuns) * 100

        // Calculate flakiness score (higher = more flaky)
        // Tests that alternate between pass/fail are flaky
        flakinessScore := calculateFlakiness(db, s.Scenario, windowStart, windowEnd)

        reliability := &CanaryTestReliability{
            ID:             ulid.Make().String(),
            Scenario:       s.Scenario,
            TestType:       s.TestType,
            WindowStart:    windowStart,
            WindowEnd:      windowEnd,
            TotalRuns:      stats.TotalRuns,
            PassedRuns:     stats.PassedRuns,
            FailedRuns:     stats.FailedRuns,
            TimeoutRuns:    stats.TimeoutRuns,
            PassRate:       passRate,
            FlakinessScore: flakinessScore,
        }

        db.Create(reliability)
    }

    return nil
}
```

#### 8.5 Database Migration

**Migration File Location:** `services/ctl-api/migrations/YYYYMMDDHHMMSS_create_canary_tables.sql`

```sql
-- Migration: Create canary testing tables
-- Run: migrate -path ./migrations -database "postgresql://..." up

BEGIN;

-- Create all canary tables
-- (Full table definitions from section 8.1 above)

-- Create indexes for performance
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_canary_runs_schedule_time
    ON canary_scheduled_runs(schedule_name, started_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_canary_results_scenario_time
    ON canary_test_results(scenario, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_canary_artifacts_expires
    ON canary_test_artifacts(expires_at)
    WHERE expires_at IS NOT NULL;

-- Grant permissions to ctl-api user
GRANT SELECT, INSERT, UPDATE, DELETE ON canary_scheduled_runs TO ctlapi;
GRANT SELECT, INSERT, UPDATE, DELETE ON canary_test_results TO ctlapi;
GRANT SELECT, INSERT, UPDATE, DELETE ON canary_schedule_history TO ctlapi;
GRANT SELECT, INSERT, UPDATE, DELETE ON canary_environments TO ctlapi;
GRANT SELECT, INSERT, UPDATE, DELETE ON canary_workflow_executions TO ctlapi;
GRANT SELECT, INSERT, UPDATE, DELETE ON canary_test_artifacts TO ctlapi;
GRANT SELECT, INSERT, UPDATE, DELETE ON canary_test_reliability TO ctlapi;
GRANT SELECT, INSERT, UPDATE, DELETE ON canary_schedules TO ctlapi;

COMMIT;
```

**Rollback Migration:** `services/ctl-api/migrations/YYYYMMDDHHMMSS_create_canary_tables.down.sql`

```sql
BEGIN;

-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS canary_test_artifacts CASCADE;
DROP TABLE IF EXISTS canary_test_reliability CASCADE;
DROP TABLE IF EXISTS canary_workflow_executions CASCADE;
DROP TABLE IF EXISTS canary_test_results CASCADE;
DROP TABLE IF EXISTS canary_environments CASCADE;
DROP TABLE IF EXISTS canary_schedule_history CASCADE;
DROP TABLE IF EXISTS canary_schedules CASCADE;
DROP TABLE IF EXISTS canary_scheduled_runs CASCADE;

COMMIT;
```

#### 8.6 Data Retention Policies

**Purpose:** Prevent unbounded database growth while retaining useful historical data.

```sql
-- Retention policy configuration
RETENTION_POLICY = {
    "test_results": 90 days,              -- Keep detailed test results for 3 months
    "scheduled_runs": 1 year,             -- Keep run summaries for 1 year
    "artifacts": 30 days,                 -- Expensive S3 storage, delete sooner
    "environments": 7 days,               -- Should be cleaned up immediately
    "workflow_executions": 90 days,       -- Match Temporal retention
    "test_reliability": 1 year,           -- Historical trend data
}
```

**Automated Cleanup Jobs:**

```go
// Background job: Clean up old test data (runs daily at 3am)
func CleanupOldCanaryData(db *gorm.DB, s3Client *s3.Client) error {
    now := time.Now()

    // 1. Delete test results older than 90 days
    result := db.Where("created_at < ?", now.Add(-90*24*time.Hour)).
        Delete(&CanaryTestResult{})
    log.Info("Deleted old test results", "count", result.RowsAffected)

    // 2. Delete scheduled runs older than 1 year
    result = db.Where("created_at < ?", now.Add(-365*24*time.Hour)).
        Delete(&CanaryScheduledRun{})
    log.Info("Deleted old scheduled runs", "count", result.RowsAffected)

    // 3. Delete expired artifacts from S3 and database
    var expiredArtifacts []CanaryTestArtifact
    db.Where("expires_at < ?", now).Find(&expiredArtifacts)

    for _, artifact := range expiredArtifacts {
        // Delete from S3
        _, err := s3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
            Bucket: aws.String(artifact.S3Bucket),
            Key:    aws.String(artifact.S3Key),
        })
        if err != nil {
            log.Error("Failed to delete S3 artifact", "error", err, "key", artifact.S3Key)
            continue
        }

        // Delete database record
        db.Delete(&artifact)
    }
    log.Info("Deleted expired artifacts", "count", len(expiredArtifacts))

    // 4. Alert on orphaned environments (not cleaned up after 7 days)
    var orphanedEnvs []CanaryEnvironment
    db.Where("cleaned_up_at IS NULL AND provisioned_at < ?", now.Add(-7*24*time.Hour)).
        Find(&orphanedEnvs)

    if len(orphanedEnvs) > 0 {
        // Send alert to Slack
        sendSlackAlert(fmt.Sprintf("⚠️ Found %d orphaned canary environments", len(orphanedEnvs)))

        // Attempt cleanup
        for _, env := range orphanedEnvs {
            if err := forceCleanupEnvironment(db, env.ID); err != nil {
                log.Error("Failed to cleanup orphaned environment", "env_id", env.ID, "error", err)
            }
        }
    }

    return nil
}
```

**Schedule in Code:**

```go
// services/ctl-api/internal/app/canary/worker/worker.go

func (w *Worker) RegisterCleanupSchedule(ctx context.Context) error {
    // Register daily cleanup job
    _, err := w.temporalClient.ScheduleClient().Create(ctx, client.ScheduleOptions{
        ID: "canary-data-cleanup",
        Spec: client.ScheduleSpec{
            CronExpressions: []string{"0 3 * * *"},  // Daily at 3am UTC
        },
        Action: &client.ScheduleWorkflowAction{
            Workflow: "CleanupCanaryDataWorkflow",
            TaskQueue: "canary",
        },
    })

    return err
}
```

#### 8.7 Useful Queries for Monitoring

```sql
-- Find flaky tests (pass rate between 30% and 70%)
SELECT
    scenario,
    test_type,
    pass_rate,
    flakiness_score,
    total_runs
FROM canary_test_reliability
WHERE window_start >= NOW() - INTERVAL '7 days'
  AND pass_rate BETWEEN 30 AND 70
ORDER BY flakiness_score DESC;

-- Find tests with consistently high failure rates
SELECT
    scenario,
    COUNT(*) as failure_count,
    AVG(duration_ms) as avg_duration_ms
FROM canary_test_results
WHERE passed = false
  AND created_at >= NOW() - INTERVAL '30 days'
GROUP BY scenario
HAVING COUNT(*) > 5
ORDER BY failure_count DESC;

-- Cleanup orphaned test environments (provisioned > 24 hours ago, not cleaned up)
SELECT
    id,
    org_id,
    provisioned_at,
    cleanup_status
FROM canary_environments
WHERE cleaned_up_at IS NULL
  AND provisioned_at < NOW() - INTERVAL '24 hours'
  AND cleanup_status = 'pending';

-- View artifacts for failed tests
SELECT
    r.scenario,
    r.passed,
    r.error_message,
    a.artifact_type,
    a.s3_url
FROM canary_test_results r
JOIN canary_test_artifacts a ON a.test_result_id = r.id
WHERE r.passed = false
  AND r.created_at >= NOW() - INTERVAL '7 days'
ORDER BY r.created_at DESC;

-- Schedule health check (ensure schedules are running)
SELECT
    schedule_name,
    is_active,
    last_run_at,
    next_run_at,
    NOW() - last_run_at as time_since_last_run
FROM canary_schedules
WHERE is_active = true
  AND (last_run_at IS NULL OR last_run_at < NOW() - INTERVAL '12 hours')
ORDER BY last_run_at ASC;
```

### 9. Monitoring & Alerting

#### 9.1 Metrics

**DataDog Metrics:**
```go
// Test execution metrics
canary.scheduled.total_tests (gauge)
canary.scheduled.passed_tests (gauge)
canary.scheduled.failed_tests (gauge)
canary.scheduled.duration (timing)
canary.scheduled.failures (count)

// Playwright-specific metrics
canary.playwright.tests_passed (gauge)
canary.playwright.tests_failed (gauge)
canary.playwright.test_duration (timing)
canary.playwright.screenshot_uploads (count)
canary.playwright.video_uploads (count)

// Tags:
// - schedule: quick-smoke, full-suite, etc.
// - environment: local, stage, prod
// - scenario: org_lifecycle, app_sync, dashboard_org_creation, etc.
// - test_type: cli, playwright
```

#### 9.2 Alerting

**Slack Notifications:**
- Send to `#canary-alerts` on test failures (both CLI and Playwright)
- Include failed test details and Temporal UI link
- For Playwright failures, include artifact links (screenshots, videos, traces)
- Daily summary of all test runs with breakdown by test type

**DataDog Monitors:**
- Alert if `canary.scheduled.failed_tests > 0` for 2 consecutive runs
- Alert if `canary.playwright.tests_failed > 0` for critical UI flows
- Alert if scheduled test doesn't run within expected window
- Alert if canary Temporal cluster is down

**PagerDuty:**
- Only for production verification failures (critical path)
- Playwright failures in `dashboard_org_creation` (critical user journey)
- Weekend comprehensive failures are non-critical (Slack only)

**Notification Format:**
```
🔴 Canary Test Failed: full-suite

CLI Tests: 3/4 passed
  ✅ org_lifecycle
  ✅ app_sync
  ✅ install_deploy
  ❌ component_build (timeout)

Playwright Tests: 2/3 passed
  ✅ dashboard_org_creation
  ❌ dashboard_install_deploy (assertion failed)
  ✅ dashboard_app_detail

🔍 View in Temporal: https://temporal-canary-web.nuon.co/...
📸 Screenshots: s3://canary-artifacts/screenshots/...
🎬 Videos: s3://canary-artifacts/videos/...
```

### 10. Local Development Workflow

#### 10.1 Setup

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

#### 10.2 Running Tests Locally

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

### 11. Production Deployment

#### 11.1 Infrastructure

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

#### 11.2 Worker Deployment

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

#### 11.3 Enable Schedules

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
- [ ] Create database migration for all canary tables
  - [ ] `canary_scheduled_runs`
  - [ ] `canary_test_results`
  - [ ] `canary_schedule_history`
  - [ ] `canary_environments`
  - [ ] `canary_workflow_executions`
  - [ ] `canary_test_artifacts`
  - [ ] `canary_test_reliability`
  - [ ] `canary_schedules`
- [ ] Run migrations in local and staging databases
- [ ] Create `services/ctl-api/internal/app/canary/` directory structure
- [ ] Define Go models for all canary tables
- [ ] Implement canary worker with dedicated Temporal client
- [ ] Add canary namespace to `cmd/worker.go`
- [ ] Test worker startup and connection

**Deliverable:** Canary worker can connect to dedicated Temporal instance with database tables ready

### Phase 2: Basic Test Execution (Week 2)
- [ ] Implement CLI executor activity
- [ ] Implement setup activities (create account, token)
- [ ] Implement cleanup activities (delete resources)
- [ ] Implement result storage activities
  - [ ] `StoreScheduledTestResults` (transactional)
  - [ ] `TrackCanaryEnvironment`
  - [ ] `CleanupCanaryEnvironment`
- [ ] Create simple org lifecycle test
- [ ] Test end-to-end: provision → test → cleanup → store results
- [ ] Verify data in all database tables

**Deliverable:** Can execute basic org CRUD test via Temporal workflow with full result storage

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

### Phase 5: Observability & Data Management (Week 5)
- [ ] Add DataDog metrics for CLI and Playwright tests
- [ ] Implement Slack notifications with artifact links
- [ ] Create DataDog dashboards showing both test types
- [ ] Set up alerts for failures and flaky tests
- [ ] Add result viewing commands (nuonctl)
- [ ] Implement flaky test detection algorithm
- [ ] Create background job for calculating test reliability
- [ ] Implement data retention policies
  - [ ] Automated cleanup job (runs daily at 3am)
  - [ ] S3 artifact expiration and deletion
  - [ ] Orphaned environment alerting and cleanup
- [ ] Add monitoring queries for database health
- [ ] Create Slack alerts for orphaned environments

**Deliverable:** Full visibility into test execution, failures, and automated data lifecycle management

### Phase 6: Dashboard Testing with Playwright (Week 6)
- [ ] Set up Playwright in `services/dashboard-ui/e2e/`
- [ ] Create Playwright configuration and authentication fixture
- [ ] Implement `dashboard_org_creation` test
- [ ] Implement `dashboard_app_detail` test
- [ ] Implement `dashboard_install_deploy` test
- [ ] Implement `dashboard_error_handling` test
- [ ] Create `PlaywrightRunner` activity in canary worker
- [ ] Add S3 artifact upload on test failure
- [ ] Integrate Playwright tests into E2ETestSuite workflow
- [ ] Update schedules to include Playwright scenarios
- [ ] Test end-to-end: Playwright execution + artifact storage

**Deliverable:** Dashboard UI testing integrated with canary system

### Phase 7: Production Rollout (Week 7)
- [ ] Deploy canary Temporal to production
- [ ] Deploy canary worker to production
- [ ] Enable quick-smoke schedule (CLI + critical Playwright)
- [ ] Monitor for 1 week
- [ ] Enable full-suite schedule (all CLI + Playwright tests)
- [ ] Enable weekend-comprehensive schedule

**Deliverable:** Canary tests (CLI + Playwright) running in production

## Success Metrics

### Test Coverage

**CLI Testing:**
- ✅ Org CRUD operations
- ✅ App sync workflow
- ✅ Install deployment
- ✅ Component builds
- ⏳ Release flow (Phase 8)
- ⏳ Multi-cloud testing (Phase 8)

**Dashboard Testing (Playwright):**
- ✅ Organization creation flow
- ✅ App detail page navigation
- ✅ Install deployment via UI
- ✅ Error handling and validation
- ⏳ User settings and preferences (Phase 8)
- ⏳ Multi-org switching (Phase 8)

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
- **2026-01-29**: Integrated Playwright dashboard testing into comprehensive E2E testing system
  - Added Section 6: Dashboard Testing with Playwright
  - Updated schedule definitions to include both CLI and Playwright test scenarios
  - Enhanced metrics and alerting to cover Playwright-specific events
  - Added Phase 6 for Playwright implementation
  - Updated success metrics, example outputs, and directory structure
- **2026-01-29**: Added Section 7: Handling Long-Running Tests
  - Multi-level timeout architecture (workflow, activity, scenario-specific)
  - Activity heartbeats for progress tracking and timeout prevention
  - Parallel execution strategies to reduce total runtime
  - Child workflows for extremely long tests
  - Graceful degradation and timeout handling
  - Resource management for Playwright browser tests
  - Monitoring and visibility for long-running test detection
  - Deep dive on Temporal's durability model: event sourcing, worker crash recovery, continue-as-new
  - Workflow querying and signaling for runtime control (pause/resume/cancel)
  - Activity checkpointing for resumable long tests
  - Comparison with other orchestration tools (GitHub Actions, Jenkins, Step Functions)
- **2026-01-29**: Removed Redis/ElastiCache references (not needed)
  - Modern Temporal uses PostgreSQL for both persistence and visibility
  - Simplified infrastructure with single RDS database
  - Reduced cost estimate to ~$40-80/month
- **2026-01-29**: Expanded database schema with comprehensive tables
  - Added `canary_environments` for tracking test account cleanup
  - Added `canary_workflow_executions` for Temporal workflow debugging
  - Added `canary_test_artifacts` for managing S3 screenshots/videos/traces
  - Added `canary_test_reliability` for flaky test detection
  - Added `canary_schedules` for active schedule configuration management
  - Enhanced result storage with transactional writes across all tables
  - Added flaky test detection algorithm and monitoring queries
  - Added orphaned environment cleanup queries
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
│       ├── dashboard/
│       │   ├── run_playwright.go       # Playwright test runner
│       │   ├── artifact_uploader.go    # S3 artifact upload
│       │   └── report_parser.go        # Parse JSON test results
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
├── rds.tf            # PostgreSQL for persistence + visibility
├── ecs.tf
├── dns.tf
└── README.md

bins/nuonctl/cmd/
└── canary_schedule.go

services/dashboard-ui/e2e/
├── fixtures/
│   └── auth.ts                         # Authentication fixture
├── org-creation.spec.ts                # Org creation tests
├── app-detail.spec.ts                  # App detail page tests
├── install-deploy.spec.ts              # Install deployment tests
├── error-handling.spec.ts              # Error handling tests
└── playwright.config.ts                # Playwright configuration
```

## Appendix B: Example Test Output

```
$ nctl canary schedule start full-suite
✅ Started canary schedule: full-suite

$ nctl canary results --schedule full-suite --limit 1

Canary Test Run: full-suite
Started:  2026-01-29 16:00:00 UTC
Completed: 2026-01-29 16:28:15 UTC
Duration: 28m15s
Environment: stage

CLI Test Results:
  ✅ org_lifecycle (2m15s)
  ✅ app_sync (8m42s)
  ✅ install_deploy (12m18s)
  ❌ component_build (1m17s)
     Error: Build timed out after 15 minutes

Dashboard Test Results (Playwright):
  ✅ dashboard_org_creation (45s)
     Tests: 2 passed, 0 failed
  ✅ dashboard_app_detail (1m23s)
     Tests: 3 passed, 0 failed
  ❌ dashboard_install_deploy (2m18s)
     Tests: 2 passed, 1 failed
     Error: Assertion failed - Deploy button not found
     Artifacts:
       📸 Screenshot: s3://canary-artifacts/screenshots/2026-01-29/install-deploy-failure.png
       🎬 Video: s3://canary-artifacts/videos/2026-01-29/install-deploy-failure.webm
       🔍 Trace: https://trace.playwright.dev/?trace=s3://...

Summary: 5/7 tests passed (71%)

View Details: https://temporal-canary-web.nuon.co/namespaces/canary/workflows/canary_full-suite_1738166400
```
