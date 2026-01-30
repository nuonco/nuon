# Playwright Tests in Canary System

This document shows how to integrate Playwright E2E tests into the canary Temporal workflow system for complete test execution tracking.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│              Canary Temporal Workflow                        │
│                                                              │
│  E2ETestSuite Workflow                                      │
│    ├─ Setup Activities                                      │
│    │   ├─ CreateCanaryAccount                              │
│    │   ├─ CreateAPIToken                                   │
│    │   └─ CreateTestOrg                                    │
│    │                                                        │
│    ├─ CLI Test Activities                                  │
│    │   ├─ TestOrgCreate (nuon orgs create)                │
│    │   └─ TestAppSync (nuon apps sync)                    │
│    │                                                        │
│    ├─ Playwright Test Activities  ← NEW                   │
│    │   ├─ RunPlaywrightTest (org-creation.spec.ts)       │
│    │   ├─ RunPlaywrightTest (app-detail.spec.ts)         │
│    │   └─ RunPlaywrightTest (install-deploy.spec.ts)     │
│    │                                                        │
│    └─ Cleanup Activities                                   │
│        └─ DeleteTestResources                              │
│                                                              │
│  Results → canary_test_results database                    │
│  Screenshots → S3 bucket (on failure)                      │
│  Metrics → DataDog                                         │
│  Alerts → Slack                                            │
└─────────────────────────────────────────────────────────────┘
```

---

## 1. Playwright Test Structure

### Directory Layout

```
services/dashboard-ui/
├── e2e/                          # Playwright E2E tests
│   ├── fixtures/                 # Shared test fixtures
│   │   └── auth.ts              # Authentication helpers
│   ├── org-creation.spec.ts     # Org creation flow
│   ├── app-detail.spec.ts       # App detail page
│   ├── install-deploy.spec.ts   # Install deployment
│   └── error-handling.spec.ts   # Error states
├── playwright.config.ts          # Playwright configuration
└── package.json
```

### Playwright Configuration

```typescript
// services/dashboard-ui/playwright.config.ts

import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,  // Sequential for canary tests
  forbidOnly: true,
  retries: 1,  // Retry once on failure
  workers: 1,  // Single worker for consistent results
  timeout: 60000,  // 60 second timeout per test

  // Output configuration for canary integration
  reporter: [
    ['json', { outputFile: 'playwright-report/results.json' }],
    ['html', { outputFolder: 'playwright-report/html', open: 'never' }],
  ],

  use: {
    // Dashboard URL from environment
    baseURL: process.env.DASHBOARD_URL || 'http://localhost:4000',

    // Trace and screenshots for debugging
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',

    // Set viewport
    viewport: { width: 1280, height: 720 },
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
})
```

### Authentication Fixture

```typescript
// services/dashboard-ui/e2e/fixtures/auth.ts

import { test as base } from '@playwright/test'

type AuthFixtures = {
  authenticatedPage: Page
}

// Fixture that provides authenticated page context
export const test = base.extend<AuthFixtures>({
  authenticatedPage: async ({ page }, use) => {
    // Get auth token from environment (set by canary activity)
    const authToken = process.env.CANARY_AUTH_TOKEN
    const orgId = process.env.CANARY_ORG_ID

    if (!authToken) {
      throw new Error('CANARY_AUTH_TOKEN not set')
    }

    // Set authentication cookie
    await page.context().addCookies([
      {
        name: 'appSession',
        value: authToken,
        domain: 'localhost',
        path: '/',
        httpOnly: true,
        secure: false,
        sameSite: 'Lax',
      },
    ])

    // Set org context cookie
    if (orgId) {
      await page.context().addCookies([
        {
          name: 'orgId',
          value: orgId,
          domain: 'localhost',
          path: '/',
        },
      ])
    }

    await use(page)
  },
})

export { expect } from '@playwright/test'
```

### Example Test

```typescript
// services/dashboard-ui/e2e/org-creation.spec.ts

import { test, expect } from './fixtures/auth'

test.describe('Organization Creation Flow', () => {
  test('should create organization via dashboard UI', async ({ authenticatedPage: page }) => {
    // Navigate to dashboard
    await page.goto('/')

    // Wait for dashboard to load
    await expect(page.locator('h1')).toBeVisible()

    // Click create organization button
    const createButton = page.locator('button:has-text("Create Organization")')
    await createButton.click()

    // Fill in organization name
    const orgName = `test-org-${Date.now()}`
    await page.fill('input[name="name"]', orgName)

    // Optional: Select sandbox mode
    await page.check('input[name="use_sandbox_mode"]')

    // Submit form
    await page.click('button[type="submit"]:has-text("Create")')

    // Wait for success toast
    await expect(page.locator('text=Organization created')).toBeVisible({ timeout: 10000 })

    // Verify redirect to org page
    await expect(page).toHaveURL(/\/orgs\/org[a-z0-9]+/)

    // Verify org name appears in page
    await expect(page.locator(`text=${orgName}`)).toBeVisible()

    // Verify org appears in org switcher
    await page.click('[data-testid="org-switcher"]')
    await expect(page.locator(`[data-testid="org-option"]:has-text("${orgName}")`)).toBeVisible()
  })

  test('should validate required fields', async ({ authenticatedPage: page }) => {
    await page.goto('/')

    // Click create organization
    await page.click('button:has-text("Create Organization")')

    // Try to submit without filling name
    await page.click('button[type="submit"]:has-text("Create")')

    // Should show validation error
    await expect(page.locator('text=Organization name is required')).toBeVisible()
  })

  test('should handle API errors gracefully', async ({ authenticatedPage: page }) => {
    // Mock API to return error
    await page.route('**/v1/orgs', (route) => {
      route.fulfill({
        status: 500,
        body: JSON.stringify({
          error: 'Internal server error',
          description: 'Database connection failed',
        }),
      })
    })

    await page.goto('/')
    await page.click('button:has-text("Create Organization")')
    await page.fill('input[name="name"]', 'test-org')
    await page.click('button[type="submit"]:has-text("Create")')

    // Should show error message
    await expect(page.locator('text=Failed to create organization')).toBeVisible()
  })
})
```

---

## 2. Canary Worker Integration

### Playwright Activity Implementation

```go
// services/ctl-api/internal/app/canary/worker/activities/dashboard/run_playwright.go

package dashboard

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "path/filepath"
    "time"

    "go.temporal.io/sdk/activity"
)

type PlaywrightRunner struct {
    dashboardPath string  // Path to dashboard-ui directory
    s3Client      *s3.Client
    s3Bucket      string
}

type RunPlaywrightRequest struct {
    TestFile        string            `json:"test_file"`       // e.g., "e2e/org-creation.spec.ts"
    DashboardURL    string            `json:"dashboard_url"`   // e.g., "http://localhost:4000"
    AuthToken       string            `json:"auth_token"`      // Canary account auth token
    OrgID           string            `json:"org_id"`          // Canary org ID
    Environment     map[string]string `json:"environment"`     // Additional env vars
    Timeout         time.Duration     `json:"timeout"`
}

type PlaywrightTestResult struct {
    TestFile      string                   `json:"test_file"`
    Success       bool                     `json:"success"`
    TestsPassed   int                      `json:"tests_passed"`
    TestsFailed   int                      `json:"tests_failed"`
    TestsSkipped  int                      `json:"tests_skipped"`
    Duration      time.Duration            `json:"duration"`
    Tests         []PlaywrightTest         `json:"tests"`
    Screenshots   []string                 `json:"screenshots"`    // S3 URLs
    Videos        []string                 `json:"videos"`         // S3 URLs
    TraceFiles    []string                 `json:"trace_files"`    // S3 URLs
    ErrorMessage  string                   `json:"error_message,omitempty"`
}

type PlaywrightTest struct {
    Title    string        `json:"title"`
    Status   string        `json:"status"`  // "passed", "failed", "skipped"
    Duration time.Duration `json:"duration"`
    Error    *TestError    `json:"error,omitempty"`
}

type TestError struct {
    Message  string `json:"message"`
    Stack    string `json:"stack"`
    Location string `json:"location"`
}

type PlaywrightJSONReport struct {
    Suites []struct {
        Title string `json:"title"`
        Tests []struct {
            Title  string `json:"title"`
            Status string `json:"status"`
            Duration int `json:"duration"`
            Error  *struct {
                Message string `json:"message"`
                Stack   string `json:"stack"`
            } `json:"error,omitempty"`
        } `json:"tests"`
    } `json:"suites"`
    Stats struct {
        Total   int `json:"total"`
        Passed  int `json:"passed"`
        Failed  int `json:"failed"`
        Skipped int `json:"skipped"`
    } `json:"stats"`
}

// @temporal-gen activity
// @activity-queue "default"
func (r *PlaywrightRunner) RunPlaywrightTest(ctx context.Context, req *RunPlaywrightRequest) (*PlaywrightTestResult, error) {
    logger := activity.GetLogger(ctx)
    startTime := time.Now()

    logger.Info("Running Playwright test",
        "test_file", req.TestFile,
        "dashboard_url", req.DashboardURL)

    // Prepare environment variables
    env := []string{
        fmt.Sprintf("DASHBOARD_URL=%s", req.DashboardURL),
        fmt.Sprintf("CANARY_AUTH_TOKEN=%s", req.AuthToken),
        fmt.Sprintf("CANARY_ORG_ID=%s", req.OrgID),
        "NODE_ENV=test",
        "CI=true",
    }

    // Add additional environment variables
    for k, v := range req.Environment {
        env = append(env, fmt.Sprintf("%s=%s", k, v))
    }

    // Create command
    cmd := exec.CommandContext(ctx, "npx", "playwright", "test", req.TestFile)
    cmd.Dir = r.dashboardPath
    cmd.Env = append(os.Environ(), env...)

    // Execute Playwright tests
    output, err := cmd.CombinedOutput()

    logger.Info("Playwright execution completed",
        "exit_code", cmd.ProcessState.ExitCode(),
        "output_length", len(output))

    // Parse JSON report
    reportPath := filepath.Join(r.dashboardPath, "playwright-report", "results.json")
    report, parseErr := r.parsePlaywrightReport(reportPath)

    result := &PlaywrightTestResult{
        TestFile: req.TestFile,
        Duration: time.Since(startTime),
    }

    if parseErr != nil {
        logger.Error("Failed to parse Playwright report", "error", parseErr)
        result.ErrorMessage = fmt.Sprintf("Failed to parse report: %v", parseErr)
        return result, nil
    }

    // Map report to result
    result.TestsPassed = report.Stats.Passed
    result.TestsFailed = report.Stats.Failed
    result.TestsSkipped = report.Stats.Skipped
    result.Success = report.Stats.Failed == 0 && err == nil

    // Extract individual test results
    for _, suite := range report.Suites {
        for _, test := range suite.Tests {
            testResult := PlaywrightTest{
                Title:    test.Title,
                Status:   test.Status,
                Duration: time.Duration(test.Duration) * time.Millisecond,
            }

            if test.Error != nil {
                testResult.Error = &TestError{
                    Message: test.Error.Message,
                    Stack:   test.Error.Stack,
                }
            }

            result.Tests = append(result.Tests, testResult)
        }
    }

    // Upload artifacts to S3 on failure
    if !result.Success {
        logger.Info("Test failed, uploading artifacts to S3")

        screenshots, _ := r.uploadArtifacts(ctx, "playwright-report/**/*.png", "screenshots")
        videos, _ := r.uploadArtifacts(ctx, "playwright-report/**/*.webm", "videos")
        traces, _ := r.uploadArtifacts(ctx, "playwright-report/**/*.zip", "traces")

        result.Screenshots = screenshots
        result.Videos = videos
        result.TraceFiles = traces
    }

    logger.Info("Playwright test completed",
        "success", result.Success,
        "passed", result.TestsPassed,
        "failed", result.TestsFailed,
        "duration", result.Duration)

    return result, nil
}

func (r *PlaywrightRunner) parsePlaywrightReport(reportPath string) (*PlaywrightJSONReport, error) {
    data, err := ioutil.ReadFile(reportPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read report: %w", err)
    }

    var report PlaywrightJSONReport
    if err := json.Unmarshal(data, &report); err != nil {
        return nil, fmt.Errorf("failed to parse JSON: %w", err)
    }

    return &report, nil
}

func (r *PlaywrightRunner) uploadArtifacts(ctx context.Context, pattern string, artifactType string) ([]string, error) {
    // Find all files matching pattern
    files, err := filepath.Glob(filepath.Join(r.dashboardPath, pattern))
    if err != nil {
        return nil, err
    }

    var s3URLs []string

    for _, file := range files {
        // Upload to S3
        key := fmt.Sprintf("canary-artifacts/%s/%s/%s",
            artifactType,
            time.Now().Format("2006-01-02"),
            filepath.Base(file))

        url, err := r.uploadFileToS3(ctx, file, key)
        if err != nil {
            return nil, err
        }

        s3URLs = append(s3URLs, url)
    }

    return s3URLs, nil
}

func (r *PlaywrightRunner) uploadFileToS3(ctx context.Context, filePath, key string) (string, error) {
    // Implementation: Upload file to S3 and return public URL
    // ... S3 upload logic ...
    return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", r.s3Bucket, key), nil
}
```

---

## 3. Workflow Integration

### E2E Test Suite Workflow

```go
// services/ctl-api/internal/app/canary/worker/e2e_test_suite.go

package worker

import (
    "fmt"
    "time"

    "go.temporal.io/sdk/workflow"

    "github.com/nuonco/nuon/pkg/types/workflows/canary"
)

func (w *Workflows) E2ETestSuite(ctx workflow.Context, req *canary.E2ETestSuiteRequest) (*canary.E2ETestSuiteResponse, error) {
    logger := workflow.GetLogger(ctx)
    startTime := workflow.Now(ctx)

    response := &canary.E2ETestSuiteResponse{
        CanaryID:    req.CanaryID,
        Results:     make(map[string]*canary.TestResult),
        TotalTests:  0,
        PassedTests: 0,
        FailedTests: 0,
    }

    // Activity options
    activityOpts := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,  // Playwright tests can take time
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 2,  // Retry once on failure
        },
    }
    ctx = workflow.WithActivityOptions(ctx, activityOpts)

    // Step 1: Setup canary environment
    var setupResult *SetupResult
    err := workflow.ExecuteActivity(ctx, "SetupCanaryEnvironment", req).Get(ctx, &setupResult)
    if err != nil {
        return nil, fmt.Errorf("setup failed: %w", err)
    }

    logger.Info("Canary environment setup complete",
        "account_id", setupResult.AccountID,
        "org_id", setupResult.OrgID,
        "api_token", "***")

    // Step 2: Run CLI tests
    if contains(req.TestScenarios, "org_lifecycle") {
        response.TotalTests++
        var result *canary.TestResult
        err := workflow.ExecuteActivity(ctx, "TestOrgLifecycle", &TestOrgRequest{
            APIToken: setupResult.APIToken,
            OrgID:    setupResult.OrgID,
        }).Get(ctx, &result)

        response.Results["org_lifecycle"] = result
        if result.Passed {
            response.PassedTests++
        } else {
            response.FailedTests++
        }
    }

    // Step 3: Run Playwright tests
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

        // Convert Playwright result to canary TestResult
        testResult := &canary.TestResult{
            Scenario:  test.Scenario,
            Passed:    playwrightResult.Success,
            Duration:  playwrightResult.Duration,
            Details: map[string]interface{}{
                "test_file":     playwrightResult.TestFile,
                "tests_passed":  playwrightResult.TestsPassed,
                "tests_failed":  playwrightResult.TestsFailed,
                "tests_skipped": playwrightResult.TestsSkipped,
                "tests":         playwrightResult.Tests,
                "screenshots":   playwrightResult.Screenshots,
                "videos":        playwrightResult.Videos,
                "traces":        playwrightResult.TraceFiles,
            },
        }

        if err != nil || !playwrightResult.Success {
            testResult.Error = playwrightResult.ErrorMessage
            if err != nil {
                testResult.Error = fmt.Sprintf("%v", err)
            }
            response.FailedTests++
        } else {
            response.PassedTests++
        }

        response.Results[test.Scenario] = testResult
    }

    // Step 4: Cleanup
    err = workflow.ExecuteActivity(ctx, "CleanupCanaryEnvironment", &CleanupRequest{
        AccountID: setupResult.AccountID,
        OrgID:     setupResult.OrgID,
    }).Get(ctx, nil)

    if err != nil {
        logger.Warn("Cleanup failed", "error", err)
    }

    response.Duration = workflow.Now(ctx).Sub(startTime)

    // Store results
    err = workflow.ExecuteActivity(ctx, "StoreTestResults", response).Get(ctx, nil)
    if err != nil {
        logger.Warn("Failed to store results", "error", err)
    }

    logger.Info("E2E test suite completed",
        "total", response.TotalTests,
        "passed", response.PassedTests,
        "failed", response.FailedTests,
        "duration", response.Duration)

    return response, nil
}
```

---

## 4. Database Storage

### Store Playwright Results

```go
// services/ctl-api/internal/app/canary/worker/activities/storage/store_results.go

package storage

import (
    "context"
    "encoding/json"

    "go.temporal.io/sdk/activity"
    "gorm.io/gorm"
)

type ResultStorage struct {
    db *gorm.DB
}

// @temporal-gen activity
// @activity-queue "default"
func (s *ResultStorage) StoreTestResults(ctx context.Context, response *canary.E2ETestSuiteResponse) error {
    logger := activity.GetLogger(ctx)

    // Store main test run
    run := &CanaryScheduledRun{
        ID:           response.CanaryID,
        ScheduleName: "e2e-suite",
        StartedAt:    response.StartTime,
        CompletedAt:  response.EndTime,
        DurationMs:   int(response.Duration.Milliseconds()),
        Environment:  response.Environment,
        TotalTests:   response.TotalTests,
        PassedTests:  response.PassedTests,
        FailedTests:  response.FailedTests,
    }

    // Convert results to JSONB
    resultsJSON, _ := json.Marshal(response.Results)
    run.Results = resultsJSON

    if err := s.db.Create(run).Error; err != nil {
        logger.Error("Failed to store test run", "error", err)
        return err
    }

    // Store individual test results
    for scenario, result := range response.Results {
        testResult := &CanaryTestResult{
            RunID:    response.CanaryID,
            Scenario: scenario,
            Passed:   result.Passed,
            Duration: int(result.Duration.Milliseconds()),
        }

        if result.Error != "" {
            testResult.ErrorMessage = &result.Error
        }

        // Store details as JSONB
        detailsJSON, _ := json.Marshal(result.Details)
        testResult.Details = detailsJSON

        if err := s.db.Create(testResult).Error; err != nil {
            logger.Warn("Failed to store individual test result", "scenario", scenario, "error", err)
        }
    }

    logger.Info("Test results stored successfully", "canary_id", response.CanaryID)

    return nil
}
```

---

## 5. Viewing Results

### Query Test Results

```sql
-- Get latest Playwright test runs
SELECT
    csr.id,
    csr.schedule_name,
    csr.started_at,
    csr.duration_ms,
    csr.passed_tests,
    csr.failed_tests,
    ctr.scenario,
    ctr.passed,
    ctr.error_message,
    ctr.details->>'screenshots' as screenshots,
    ctr.details->>'videos' as videos
FROM canary_scheduled_runs csr
JOIN canary_test_results ctr ON ctr.run_id = csr.id
WHERE ctr.scenario LIKE 'dashboard_%'
ORDER BY csr.started_at DESC
LIMIT 20;

-- Get Playwright test pass rate by scenario
SELECT
    scenario,
    COUNT(*) as total_runs,
    SUM(CASE WHEN passed THEN 1 ELSE 0 END) as passed_count,
    ROUND(AVG(CASE WHEN passed THEN 1.0 ELSE 0.0 END) * 100, 2) as pass_rate,
    AVG(duration_ms) as avg_duration_ms
FROM canary_test_results
WHERE scenario LIKE 'dashboard_%'
  AND created_at > NOW() - INTERVAL '7 days'
GROUP BY scenario
ORDER BY pass_rate ASC;

-- Get recent Playwright test failures with artifacts
SELECT
    ctr.scenario,
    ctr.error_message,
    ctr.details->>'test_file' as test_file,
    ctr.details->>'screenshots' as screenshots,
    ctr.details->>'videos' as videos,
    ctr.details->>'traces' as traces,
    ctr.created_at
FROM canary_test_results ctr
WHERE ctr.scenario LIKE 'dashboard_%'
  AND ctr.passed = false
  AND ctr.created_at > NOW() - INTERVAL '24 hours'
ORDER BY ctr.created_at DESC;
```

### nuonctl Commands

```bash
# View recent Playwright test results
nctl canary results --type playwright --limit 10

# View specific test scenario results
nctl canary results --scenario dashboard_org_creation --limit 5

# View failed tests with artifacts
nctl canary results --failed-only --with-artifacts

# Download screenshots from failed test
nctl canary artifacts download --run-id canary-123 --type screenshots
```

---

## 6. Temporal UI View

When you open the Temporal UI for a test run:

```
Workflow: E2ETestSuite
ID: canary-e2e-suite-1738166400
Status: ✅ Completed
Duration: 8m 23s

Activities:
  1. SetupCanaryEnvironment          [✅ 45s]
  2. TestOrgLifecycle                [✅ 12s]
  3. RunPlaywrightTest               [✅ 2m 34s]
     - Test: e2e/org-creation.spec.ts
     - Passed: 3/3 tests
     - Screenshots: 0
  4. RunPlaywrightTest               [❌ 3m 12s]
     - Test: e2e/app-detail.spec.ts
     - Passed: 2/3 tests
     - Failed: 1 test
     - Screenshots: 2 (uploaded to S3)
     - Videos: 1 (uploaded to S3)
     - Traces: 1 (uploaded to S3)
  5. RunPlaywrightTest               [✅ 1m 48s]
     - Test: e2e/install-deploy.spec.ts
     - Passed: 4/4 tests
  6. CleanupCanaryEnvironment        [✅ 23s]
  7. StoreTestResults                [✅ 2s]

Result: 2/3 Playwright tests passed (9/10 individual test cases)
```

---

## 7. Notifications

### Slack Notification on Failure

```go
// services/ctl-api/internal/app/canary/worker/activities/notifications/slack.go

func (n *Notifier) SendPlaywrightFailureNotification(ctx context.Context, result *PlaywrightTestResult) error {
    if result.Success {
        return nil // Only notify on failures
    }

    // Build list of failed tests
    var failedTests []string
    for _, test := range result.Tests {
        if test.Status == "failed" {
            failedTests = append(failedTests, fmt.Sprintf("  • %s", test.Title))
        }
    }

    // Build artifact links
    var artifactLinks []string
    for _, screenshot := range result.Screenshots {
        artifactLinks = append(artifactLinks, fmt.Sprintf("📸 <https://s3.amazonaws.com/%s|Screenshot>", screenshot))
    }
    for _, video := range result.Videos {
        artifactLinks = append(artifactLinks, fmt.Sprintf("🎬 <https://s3.amazonaws.com/%s|Video>", video))
    }
    for _, trace := range result.TraceFiles {
        artifactLinks = append(artifactLinks, fmt.Sprintf("🔍 <https://trace.playwright.dev/?trace=%s|Trace>", trace))
    }

    message := fmt.Sprintf(`
🔴 *Playwright Test Failed*

*Test File:* %s
*Failed Tests:* %d/%d
*Duration:* %s

*Failed Test Cases:*
%s

*Debug Artifacts:*
%s

<https://temporal-canary-web.nuon.co/namespaces/canary/workflows/%s|View in Temporal UI>
    `,
        result.TestFile,
        result.TestsFailed,
        result.TestsPassed+result.TestsFailed,
        result.Duration,
        strings.Join(failedTests, "\n"),
        strings.Join(artifactLinks, "\n"),
        activity.GetInfo(ctx).WorkflowExecution.ID,
    )

    return n.slackClient.SendMessage(ctx, "#canary-alerts", message)
}
```

---

## 8. Complete Flow Example

### Running the System

```bash
# 1. Start infrastructure
nctl scripts exec reset-dependencies
nctl scripts exec reset-dependencies-canary

# 2. Start dashboard (for local testing)
cd services/dashboard-ui
npm run dev

# 3. Start canary worker
cd services/ctl-api
go run main.go worker --namespace canary

# 4. Trigger E2E test suite
curl -X POST http://localhost:8081/v1/general/canary/test/e2e-suite \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "test_scenarios": [
      "org_lifecycle",
      "dashboard_org_creation",
      "dashboard_app_detail"
    ],
    "dashboard_url": "http://localhost:4000"
  }'
```

### Expected Output

```json
{
  "canary_id": "canary-e2e-123",
  "total_tests": 3,
  "passed_tests": 2,
  "failed_tests": 1,
  "duration": "8m23s",
  "results": {
    "org_lifecycle": {
      "passed": true,
      "duration": "12s"
    },
    "dashboard_org_creation": {
      "passed": true,
      "duration": "2m34s",
      "details": {
        "test_file": "e2e/org-creation.spec.ts",
        "tests_passed": 3,
        "tests_failed": 0
      }
    },
    "dashboard_app_detail": {
      "passed": false,
      "duration": "3m12s",
      "error": "Test failed: timeout waiting for element",
      "details": {
        "test_file": "e2e/app-detail.spec.ts",
        "tests_passed": 2,
        "tests_failed": 1,
        "screenshots": [
          "https://s3.../canary-artifacts/screenshots/2026-01-29/failure-1.png"
        ],
        "videos": [
          "https://s3.../canary-artifacts/videos/2026-01-29/failure-1.webm"
        ]
      }
    }
  }
}
```

---

## Key Benefits

✅ **Complete Integration**: Playwright tests run within Temporal workflows
✅ **Historical Tracking**: All test runs stored in canary database
✅ **Artifact Management**: Screenshots/videos uploaded to S3 on failures
✅ **Unified Monitoring**: Same alerts, metrics, dashboards as CLI tests
✅ **Debugging**: Temporal UI shows full execution trace
✅ **Scheduling**: Use same cron scheduling as other canary tests
✅ **Reporting**: Query results via SQL or nuonctl commands

This gives you complete visibility into your dashboard E2E tests with the same infrastructure as your CLI tests!
