# Dashboard UX Testing Options

This document explores different approaches for testing the Nuon dashboard user experience within the canary testing framework.

## Current State

### Existing Test Infrastructure
- ✅ **Unit Tests**: Vitest + React Testing Library (~100+ test files)
- ✅ **Mock Service Worker**: MSW for API mocking in tests
- ✅ **Component Stories**: Ladle for component development
- ❌ **Browser E2E Tests**: None currently
- ❌ **Visual Regression**: None currently

### Dashboard Architecture Context
- **Framework**: Next.js 15 with App Router
- **Language**: TypeScript + React
- **Authentication**: Auth0
- **API Communication**: Server actions + API routes
- **Local Dev**: Runs on `http://localhost:4000`

---

## Option 1: Browser Automation (Playwright) ⭐ RECOMMENDED

### Overview
Use Playwright to automate real browser interactions - clicking buttons, filling forms, navigating pages, verifying UI elements.

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│           Canary Temporal Workflow                           │
│                                                              │
│  Activities:                                                 │
│  1. SetupCanaryEnvironment (account, token, org)            │
│  2. StartDashboardServer (or point to existing)             │
│  3. RunPlaywrightTest (browser automation)                  │
│  4. CaptureScreenshots (on failure)                         │
│  5. CleanupTestData                                         │
└─────────────────────────────────────────────────────────────┘
```

### Implementation

#### Install Playwright
```bash
cd services/dashboard-ui
npm install -D @playwright/test
npx playwright install
```

#### Configuration
```typescript
// services/dashboard-ui/playwright.config.ts

import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,  // Run tests sequentially for canary
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,  // Single worker for canary tests

  reporter: [
    ['html'],
    ['json', { outputFile: 'playwright-report/results.json' }],
  ],

  use: {
    baseURL: process.env.DASHBOARD_URL || 'http://localhost:4000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
})
```

#### Example Test: Create Org via Dashboard
```typescript
// services/dashboard-ui/e2e/org-creation.spec.ts

import { test, expect } from '@playwright/test'

test.describe('Organization Creation', () => {
  let authToken: string

  test.beforeAll(async ({ request }) => {
    // Get auth token from canary activity
    authToken = process.env.CANARY_AUTH_TOKEN
  })

  test('should create organization via dashboard', async ({ page }) => {
    // Set authentication cookie
    await page.context().addCookies([{
      name: 'appSession',
      value: authToken,
      domain: 'localhost',
      path: '/',
    }])

    // Navigate to dashboard
    await page.goto('/')

    // Should see dashboard after auth
    await expect(page).toHaveTitle(/Nuon Dashboard/)

    // Click "Create Organization" button
    await page.click('text=Create Organization')

    // Fill in org name
    const orgName = `test-org-${Date.now()}`
    await page.fill('input[name="name"]', orgName)

    // Submit form
    await page.click('button:has-text("Create")')

    // Wait for success notification
    await expect(page.locator('text=Organization created')).toBeVisible()

    // Verify redirect to org page
    await expect(page).toHaveURL(/\/orgs\/org[a-z0-9]+/)

    // Verify org name appears in UI
    await expect(page.locator(`text=${orgName}`)).toBeVisible()
  })

  test('should show validation error for empty org name', async ({ page }) => {
    await page.goto('/')
    await page.click('text=Create Organization')

    // Submit without filling name
    await page.click('button:has-text("Create")')

    // Should show validation error
    await expect(page.locator('text=Organization name is required')).toBeVisible()
  })
})
```

#### Example Test: App Sync Flow
```typescript
// services/dashboard-ui/e2e/app-sync.spec.ts

import { test, expect } from '@playwright/test'

test.describe('App Sync Flow', () => {
  test('should display app after CLI sync', async ({ page }) => {
    // Prerequisite: CLI has already synced an app (done by canary activity)
    const appId = process.env.CANARY_APP_ID
    const orgId = process.env.CANARY_ORG_ID

    await page.goto(`/${orgId}/apps`)

    // Should see the synced app in list
    await expect(page.locator(`[data-app-id="${appId}"]`)).toBeVisible()

    // Click on app to view details
    await page.click(`[data-app-id="${appId}"]`)

    // Should navigate to app detail page
    await expect(page).toHaveURL(`/${orgId}/apps/${appId}`)

    // Should see app name and components
    await expect(page.locator('h1')).toContainText('test-app')
    await expect(page.locator('text=Components')).toBeVisible()
  })

  test('should trigger component build from dashboard', async ({ page }) => {
    const appId = process.env.CANARY_APP_ID
    const orgId = process.env.CANARY_ORG_ID
    const componentId = process.env.CANARY_COMPONENT_ID

    await page.goto(`/${orgId}/apps/${appId}`)

    // Find component row and click build button
    await page.locator(`[data-component-id="${componentId}"]`).locator('button:has-text("Build")').click()

    // Should show build started notification
    await expect(page.locator('text=Build started')).toBeVisible()

    // Should see building status
    await expect(page.locator(`[data-component-id="${componentId}"]`).locator('text=Building')).toBeVisible({ timeout: 10000 })
  })
})
```

### Canary Integration

```go
// services/ctl-api/internal/app/canary/worker/activities/dashboard/run_playwright_test.go

package dashboard

import (
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "time"

    "go.temporal.io/sdk/activity"
)

type PlaywrightExecutor struct {
    dashboardPath string
}

type PlaywrightTestRequest struct {
    TestFile        string            `json:"test_file"`
    DashboardURL    string            `json:"dashboard_url"`
    Environment     map[string]string `json:"environment"`
    Timeout         time.Duration     `json:"timeout"`
}

type PlaywrightTestResult struct {
    Success     bool              `json:"success"`
    TestsPassed int               `json:"tests_passed"`
    TestsFailed int               `json:"tests_failed"`
    Duration    time.Duration     `json:"duration"`
    Report      *PlaywrightReport `json:"report"`
    Screenshots []string          `json:"screenshots"`
    Videos      []string          `json:"videos"`
}

type PlaywrightReport struct {
    Stats struct {
        Total   int `json:"total"`
        Passed  int `json:"passed"`
        Failed  int `json:"failed"`
        Skipped int `json:"skipped"`
    } `json:"stats"`
    Suites []struct {
        Title string `json:"title"`
        Tests []struct {
            Title  string `json:"title"`
            Status string `json:"status"`
            Error  string `json:"error,omitempty"`
        } `json:"tests"`
    } `json:"suites"`
}

// @temporal-gen activity
// @activity-queue "default"
func (e *PlaywrightExecutor) RunPlaywrightTest(ctx context.Context, req *PlaywrightTestRequest) (*PlaywrightTestResult, error) {
    logger := activity.GetLogger(ctx)
    startTime := time.Now()

    logger.Info("Running Playwright test",
        "test_file", req.TestFile,
        "dashboard_url", req.DashboardURL)

    // Build command
    cmd := exec.CommandContext(ctx, "npx", "playwright", "test", req.TestFile, "--reporter=json")
    cmd.Dir = e.dashboardPath

    // Set environment variables
    env := []string{
        fmt.Sprintf("DASHBOARD_URL=%s", req.DashboardURL),
    }
    for k, v := range req.Environment {
        env = append(env, fmt.Sprintf("%s=%s", k, v))
    }
    cmd.Env = append(cmd.Env, env...)

    // Execute tests
    output, err := cmd.CombinedOutput()

    // Parse JSON report
    var report PlaywrightReport
    if jsonErr := json.Unmarshal(output, &report); jsonErr != nil {
        logger.Warn("Failed to parse Playwright report", "error", jsonErr)
    }

    result := &PlaywrightTestResult{
        Success:     err == nil && report.Stats.Failed == 0,
        TestsPassed: report.Stats.Passed,
        TestsFailed: report.Stats.Failed,
        Duration:    time.Since(startTime),
        Report:      &report,
    }

    // Collect screenshots and videos on failure
    if !result.Success {
        result.Screenshots = e.collectArtifacts("test-results/**/*.png")
        result.Videos = e.collectArtifacts("test-results/**/*.webm")
    }

    logger.Info("Playwright test completed",
        "success", result.Success,
        "passed", result.TestsPassed,
        "failed", result.TestsFailed,
        "duration", result.Duration)

    return result, nil
}
```

### Pros & Cons

**Pros:**
- ✅ Tests real user interactions in actual browser
- ✅ Catches UI/UX bugs that unit tests miss
- ✅ Tests full authentication flow
- ✅ Verifies visual elements and layout
- ✅ Can capture screenshots/videos on failure
- ✅ Industry standard (used by GitHub, Microsoft, etc.)

**Cons:**
- ❌ Slower than unit tests (seconds per test)
- ❌ More brittle (UI changes break tests)
- ❌ Requires running dashboard server
- ❌ Authentication setup complexity
- ❌ Parallel execution challenging

**Best For:**
- Critical user flows (login, org creation, app sync, deploy)
- Regression testing of major features
- Scheduled daily/weekly full test runs

---

## Option 2: API-Based Testing ⭐ GOOD COMPLEMENT

### Overview
Test the APIs that the dashboard calls, verifying backend behavior without browser automation.

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│           Canary Test Workflow                               │
│                                                              │
│  Activities:                                                 │
│  1. SetupCanaryAccount (create account, get token)          │
│  2. TestOrgCRUD (via API endpoints)                         │
│  3. TestAppSync (via API endpoints)                         │
│  4. TestInstallCreation (via API endpoints)                 │
│  5. VerifyDatabaseState (direct DB checks)                  │
└─────────────────────────────────────────────────────────────┘
```

### Implementation

```go
// services/ctl-api/internal/app/canary/worker/activities/api/test_org_crud.go

package api

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "go.temporal.io/sdk/activity"
)

type APITester struct {
    apiURL    string
    apiClient *http.Client
}

type OrgCRUDTestRequest struct {
    APIToken string `json:"api_token"`
    OrgName  string `json:"org_name"`
}

type OrgCRUDTestResult struct {
    CreateSuccess bool          `json:"create_success"`
    GetSuccess    bool          `json:"get_success"`
    UpdateSuccess bool          `json:"update_success"`
    DeleteSuccess bool          `json:"delete_success"`
    OrgID         string        `json:"org_id"`
    Duration      time.Duration `json:"duration"`
    Error         string        `json:"error,omitempty"`
}

// @temporal-gen activity
// @activity-queue "default"
func (a *APITester) TestOrgCRUD(ctx context.Context, req *OrgCRUDTestRequest) (*OrgCRUDTestResult, error) {
    logger := activity.GetLogger(ctx)
    startTime := time.Now()
    result := &OrgCRUDTestResult{}

    // Step 1: Create Org via API
    createResp, err := a.createOrg(ctx, req.APIToken, req.OrgName)
    if err != nil {
        result.Error = fmt.Sprintf("Create failed: %v", err)
        return result, nil
    }
    result.CreateSuccess = true
    result.OrgID = createResp.ID

    logger.Info("Created org via API", "org_id", result.OrgID)

    // Step 2: Get Org via API
    _, err = a.getOrg(ctx, req.APIToken, result.OrgID)
    if err != nil {
        result.Error = fmt.Sprintf("Get failed: %v", err)
        return result, nil
    }
    result.GetSuccess = true

    // Step 3: Update Org via API
    updatedName := fmt.Sprintf("%s-updated", req.OrgName)
    err = a.updateOrg(ctx, req.APIToken, result.OrgID, updatedName)
    if err != nil {
        result.Error = fmt.Sprintf("Update failed: %v", err)
        return result, nil
    }
    result.UpdateSuccess = true

    // Step 4: Delete Org via API
    err = a.deleteOrg(ctx, req.APIToken, result.OrgID)
    if err != nil {
        result.Error = fmt.Sprintf("Delete failed: %v", err)
        return result, nil
    }
    result.DeleteSuccess = true

    result.Duration = time.Since(startTime)

    logger.Info("Org CRUD test completed",
        "org_id", result.OrgID,
        "duration", result.Duration)

    return result, nil
}

func (a *APITester) createOrg(ctx context.Context, token, name string) (*OrgResponse, error) {
    req, _ := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/v1/orgs", a.apiURL),
        strings.NewReader(fmt.Sprintf(`{"name": "%s"}`, name)))

    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
    req.Header.Set("Content-Type", "application/json")

    resp, err := a.apiClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    var orgResp OrgResponse
    if err := json.NewDecoder(resp.Body).Decode(&orgResp); err != nil {
        return nil, err
    }

    return &orgResp, nil
}
```

### Pros & Cons

**Pros:**
- ✅ Fast execution (milliseconds per test)
- ✅ No browser/UI dependencies
- ✅ Easy to parallelize
- ✅ Reliable and stable
- ✅ Can test edge cases easily
- ✅ Direct API verification

**Cons:**
- ❌ Doesn't test actual UI/UX
- ❌ Can't catch visual bugs
- ❌ Doesn't verify user-facing errors
- ❌ Misses client-side validation
- ❌ No authentication flow testing

**Best For:**
- Backend logic verification
- API contract testing
- Performance testing
- Quick smoke tests
- Database state validation

---

## Option 3: Hybrid Approach (API + Playwright) ⭐ BEST OF BOTH

### Overview
Combine API testing for backend verification with selective Playwright tests for critical UI flows.

### Test Distribution

```
┌─────────────────────────────────────────────────────────────┐
│                      Test Pyramid                            │
│                                                              │
│              🔺 Playwright E2E (5-10 tests)                 │
│                 Critical user flows only                     │
│                                                              │
│         🔺🔺🔺 API Tests (30-50 tests)                      │
│              Backend logic & integration                     │
│                                                              │
│     🔺🔺🔺🔺🔺 Unit Tests (100+ tests)                      │
│            Component & function testing                      │
└─────────────────────────────────────────────────────────────┘
```

### Implementation Strategy

**Fast API Tests (Run every 6 hours):**
- Org CRUD
- App sync API
- Install creation API
- Component build API
- Basic health checks

**Playwright UI Tests (Run daily):**
- Login flow
- Org creation via dashboard
- App detail page navigation
- Deploy button interaction
- Error message display

### Workflow Example

```go
// services/ctl-api/internal/app/canary/worker/dashboard_test_suite.go

func (w *Workflows) DashboardTestSuite(ctx workflow.Context, req *DashboardTestRequest) (*DashboardTestResult, error) {
    result := &DashboardTestResult{}

    // Phase 1: Fast API tests (5-10 minutes)
    apiTests := []string{
        "org_crud",
        "app_sync_api",
        "install_creation",
        "component_build",
    }

    for _, test := range apiTests {
        var testResult *APITestResult
        err := workflow.ExecuteActivity(ctx, "RunAPITest", test).Get(ctx, &testResult)

        if err != nil || !testResult.Success {
            result.FailedTests = append(result.FailedTests, test)
        }
    }

    // Phase 2: Selective Playwright tests (only if API tests pass)
    if len(result.FailedTests) == 0 && req.RunUITests {
        playwrightTests := []string{
            "e2e/org-creation.spec.ts",
            "e2e/app-detail-page.spec.ts",
        }

        for _, test := range playwrightTests {
            var testResult *PlaywrightTestResult
            err := workflow.ExecuteActivity(ctx, "RunPlaywrightTest", &PlaywrightTestRequest{
                TestFile: test,
            }).Get(ctx, &testResult)

            if err != nil || !testResult.Success {
                result.FailedTests = append(result.FailedTests, test)
            }
        }
    }

    return result, nil
}
```

### Pros & Cons

**Pros:**
- ✅ Fast feedback from API tests
- ✅ Critical UI flows verified
- ✅ Balanced speed vs coverage
- ✅ Catches both backend and frontend issues
- ✅ Pragmatic test distribution

**Cons:**
- ❌ Two test systems to maintain
- ❌ More complex setup
- ❌ Need to decide what goes where

**Best For:**
- Production canary testing
- Comprehensive coverage
- Balanced test execution time

---

## Option 4: Visual Regression Testing

### Overview
Capture screenshots of UI states and compare against baselines to detect visual regressions.

### Tools
- **Percy** (commercial, CI/CD integration)
- **Chromatic** (commercial, Storybook integration)
- **Playwright Visual Comparisons** (open source)

### Implementation with Playwright

```typescript
// services/dashboard-ui/e2e/visual-regression.spec.ts

import { test, expect } from '@playwright/test'

test('org list page visual snapshot', async ({ page }) => {
  await page.goto('/orgs')

  // Wait for page to load completely
  await page.waitForLoadState('networkidle')

  // Take screenshot and compare
  await expect(page).toHaveScreenshot('org-list-page.png', {
    fullPage: true,
    maxDiffPixels: 100,  // Allow minor differences
  })
})

test('install detail page visual snapshot', async ({ page }) => {
  await page.goto('/orgs/test-org/installs/test-install')
  await page.waitForLoadState('networkidle')

  await expect(page).toHaveScreenshot('install-detail-page.png', {
    fullPage: true,
  })
})
```

### Pros & Cons

**Pros:**
- ✅ Catches unintended visual changes
- ✅ Detects CSS regressions
- ✅ Verifies layout across browsers
- ✅ Good for design system testing

**Cons:**
- ❌ High false positive rate
- ❌ Requires baseline management
- ❌ Slow (screenshot comparison)
- ❌ Flaky with dynamic content
- ❌ Commercial tools can be expensive

**Best For:**
- Design system validation
- Major UI refactors
- Cross-browser testing
- Marketing pages

---

## Option 5: Component Testing with Ladle ✅ ALREADY HAVE

### Overview
You already have Ladle set up for component development. This can be extended for automated testing.

### Current Setup
```bash
npm run dev:ladle  # Component development server
```

### Extension for Testing

```typescript
// services/dashboard-ui/.ladle/test-runner.ts

import { test } from 'vitest'
import { render } from '@testing-library/react'
import * as stories from '../src/components/common/Button.stories'

// Auto-generate tests from stories
Object.entries(stories).forEach(([name, Story]) => {
  if (name === 'default') return

  test(`${name} story renders`, () => {
    const { container } = render(<Story />)
    expect(container).toBeTruthy()
  })
})
```

### Pros & Cons

**Pros:**
- ✅ Already implemented
- ✅ Fast component testing
- ✅ Visual component review
- ✅ Design system validation

**Cons:**
- ❌ Not true E2E testing
- ❌ Doesn't test integration
- ❌ No real data flow

**Best For:**
- Component development
- Design system validation
- Unit-level component testing

---

## Recommended Approach

### Phase 1: Start with API Testing (Week 1-2)
- Implement API test activities in canary worker
- Test critical flows: org, app, install
- Run every 6 hours
- **Pros**: Fast, reliable, immediate value

### Phase 2: Add Playwright for Critical Flows (Week 3-4)
- Implement 3-5 critical UI tests:
  1. Org creation flow
  2. App detail page navigation
  3. Deploy button interaction
- Run daily
- **Pros**: Covers most important user paths

### Phase 3: Expand Coverage (Week 5+)
- Add more Playwright tests for secondary flows
- Add visual regression for design system
- Optimize test execution time

### Hybrid Schedule Example

```yaml
schedules:
  quick-smoke:
    cron: "0 */6 * * *"  # Every 6 hours
    tests:
      - api/org-crud
      - api/app-sync
      - api/install-creation

  full-ui-suite:
    cron: "0 16 * * *"  # Daily at 4pm
    tests:
      - api/*             # All API tests
      - playwright/critical/*  # Critical UI flows

  comprehensive:
    cron: "0 6 * * 6"   # Saturday 6am
    tests:
      - api/*
      - playwright/*      # All UI tests
      - visual-regression/*
```

---

## Cost Comparison

| Approach | Setup Time | Execution Time | Maintenance | Reliability |
|----------|-----------|----------------|-------------|-------------|
| **Playwright** | 2-3 days | 2-5 min/test | Medium | Medium |
| **API Testing** | 1-2 days | <1 sec/test | Low | High |
| **Hybrid** | 3-4 days | 30sec-2min | Medium | High |
| **Visual Regression** | 2-3 days | 3-10 min/test | High | Low |
| **Ladle** | ✅ Already have | <1 sec/test | Low | High |

---

## Implementation Checklist

### For Playwright (Option 1)
- [ ] Install Playwright in dashboard-ui
- [ ] Configure playwright.config.ts
- [ ] Create e2e/ directory
- [ ] Write 3-5 critical flow tests
- [ ] Add Playwright activity to canary worker
- [ ] Handle authentication in tests
- [ ] Set up screenshot/video capture
- [ ] Add to canary test schedule

### For API Testing (Option 2)
- [ ] Create API tester activity
- [ ] Implement HTTP client with auth
- [ ] Write API test scenarios
- [ ] Add database verification
- [ ] Integrate with canary workflow
- [ ] Set up failure reporting

### For Hybrid (Option 3)
- [ ] Implement both Playwright + API testing
- [ ] Define test distribution strategy
- [ ] Create separate schedules
- [ ] Optimize execution order

---

## Next Steps

1. **Decision**: Choose Hybrid approach (API + Playwright)
2. **Start**: Implement API testing first (fast ROI)
3. **Add**: 3-5 Playwright tests for critical flows
4. **Schedule**: Run API tests every 6h, UI tests daily
5. **Iterate**: Expand coverage based on findings

This gives you comprehensive dashboard coverage while maintaining fast feedback loops!
