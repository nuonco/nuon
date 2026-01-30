# Example: Testing `nuon orgs create foo`

This document shows a complete example of testing the `nuon orgs create foo` CLI command.

## Test Flow

```
┌─────────────────────────────────────────────────────────────┐
│                   OrgCreateTest Workflow                     │
│                                                              │
│  1. Setup: Create canary account & API token                │
│  2. Execute: nuon orgs create test-org-{timestamp}          │
│  3. Parse: Extract org ID from CLI output                   │
│  4. Verify: Check org exists via API                        │
│  5. Verify: Check org exists in database                    │
│  6. Cleanup: Delete test org                                │
│  7. Verify: Confirm org was deleted                         │
│                                                              │
│  Result: Pass/Fail with detailed error information          │
└─────────────────────────────────────────────────────────────┘
```

---

## 1. Activity: Run CLI Command

```go
// services/ctl-api/internal/app/canary/worker/activities/cli/run_command.go

package cli

import (
    "context"
    "fmt"
    "os/exec"
    "strings"
    "time"

    "go.temporal.io/sdk/activity"
)

type Executor struct {
    apiURL   string
    apiToken string
}

type CommandRequest struct {
    Command []string `json:"command"`  // ["orgs", "create", "foo"]
    Timeout time.Duration `json:"timeout"`
}

type CommandResult struct {
    Command    []string      `json:"command"`
    Output     string        `json:"output"`
    ExitCode   int           `json:"exit_code"`
    Success    bool          `json:"success"`
    Duration   time.Duration `json:"duration"`
    ErrorMsg   string        `json:"error_msg,omitempty"`
}

// @temporal-gen activity
// @activity-queue "default"
func (e *Executor) RunCLICommand(ctx context.Context, req *CommandRequest) (*CommandResult, error) {
    logger := activity.GetLogger(ctx)
    startTime := time.Now()

    logger.Info("Running CLI command",
        "command", strings.Join(req.Command, " "),
        "api_url", e.apiURL)

    // Create command context with timeout
    cmdCtx, cancel := context.WithTimeout(ctx, req.Timeout)
    defer cancel()

    // Build nuon command
    cmd := exec.CommandContext(cmdCtx, "nuon", req.Command...)

    // Set environment variables for CLI
    cmd.Env = []string{
        fmt.Sprintf("NUON_API_URL=%s", e.apiURL),
        fmt.Sprintf("NUON_API_TOKEN=%s", e.apiToken),
        "PATH=" + os.Getenv("PATH"),
        "HOME=" + os.Getenv("HOME"),
    }

    // Capture combined stdout/stderr
    output, err := cmd.CombinedOutput()

    result := &CommandResult{
        Command:  req.Command,
        Output:   string(output),
        ExitCode: 0,
        Success:  true,
        Duration: time.Since(startTime),
    }

    if err != nil {
        result.Success = false
        result.ErrorMsg = err.Error()

        if exitErr, ok := err.(*exec.ExitError); ok {
            result.ExitCode = exitErr.ExitCode()
        } else {
            result.ExitCode = -1
        }

        logger.Error("CLI command failed",
            "command", strings.Join(req.Command, " "),
            "exit_code", result.ExitCode,
            "output", result.Output,
            "error", err)
    } else {
        logger.Info("CLI command succeeded",
            "command", strings.Join(req.Command, " "),
            "duration", result.Duration,
            "output_length", len(result.Output))
    }

    return result, nil
}
```

---

## 2. Activity: Parse CLI Output

```go
// services/ctl-api/internal/app/canary/worker/activities/cli/parse_output.go

package cli

import (
    "context"
    "encoding/json"
    "fmt"
    "regexp"
    "strings"

    "go.temporal.io/sdk/activity"
)

type ParsedOrgOutput struct {
    OrgID   string `json:"org_id"`
    OrgName string `json:"org_name"`
}

// @temporal-gen activity
// @activity-queue "default"
func (e *Executor) ParseOrgCreateOutput(ctx context.Context, output string) (*ParsedOrgOutput, error) {
    logger := activity.GetLogger(ctx)

    // The CLI might output in different formats:
    // 1. JSON format (if --output json flag used)
    // 2. Human-readable format

    // Try JSON first
    if strings.HasPrefix(strings.TrimSpace(output), "{") {
        var result struct {
            ID   string `json:"id"`
            Name string `json:"name"`
        }
        if err := json.Unmarshal([]byte(output), &result); err == nil {
            return &ParsedOrgOutput{
                OrgID:   result.ID,
                OrgName: result.Name,
            }, nil
        }
    }

    // Parse human-readable format
    // Example output:
    // ✅ Created organization: test-org-1738166400
    // Organization ID: org123abc456def789ghi
    // Name: test-org-1738166400

    // Extract org ID
    orgIDPattern := regexp.MustCompile(`Organization ID:\s*(\w+)`)
    if matches := orgIDPattern.FindStringSubmatch(output); len(matches) > 1 {
        orgID := matches[1]

        // Extract org name
        namePattern := regexp.MustCompile(`Name:\s*(.+)`)
        var orgName string
        if matches := namePattern.FindStringSubmatch(output); len(matches) > 1 {
            orgName = strings.TrimSpace(matches[1])
        }

        logger.Info("Parsed org create output",
            "org_id", orgID,
            "org_name", orgName)

        return &ParsedOrgOutput{
            OrgID:   orgID,
            OrgName: orgName,
        }, nil
    }

    // If we can't parse, return error
    return nil, fmt.Errorf("unable to parse org ID from output: %s", output)
}
```

---

## 3. Activity: Verify Org via API

```go
// services/ctl-api/internal/app/canary/worker/activities/validation/verify_org.go

package validation

import (
    "context"
    "fmt"
    "time"

    "go.temporal.io/sdk/activity"
    "gorm.io/gorm"

    "github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type Validator struct {
    db        *gorm.DB
    apiClient *api.Client
}

type VerifyOrgRequest struct {
    OrgID      string `json:"org_id"`
    OrgName    string `json:"org_name"`
    AccountID  string `json:"account_id"`
}

type VerifyOrgResult struct {
    Exists       bool      `json:"exists"`
    Name         string    `json:"name"`
    Status       string    `json:"status"`
    CreatedAt    time.Time `json:"created_at"`
    HasAdminRole bool      `json:"has_admin_role"`
}

// @temporal-gen activity
// @activity-queue "default"
func (v *Validator) VerifyOrgExistsInAPI(ctx context.Context, req *VerifyOrgRequest) (*VerifyOrgResult, error) {
    logger := activity.GetLogger(ctx)

    logger.Info("Verifying org via API",
        "org_id", req.OrgID,
        "org_name", req.OrgName)

    // Query API for org
    org, err := v.apiClient.GetOrg(ctx, req.OrgID)
    if err != nil {
        logger.Error("Org not found via API", "org_id", req.OrgID, "error", err)
        return &VerifyOrgResult{
            Exists: false,
        }, nil
    }

    // Verify name matches
    if org.Name != req.OrgName {
        return nil, fmt.Errorf("org name mismatch: expected %s, got %s", req.OrgName, org.Name)
    }

    logger.Info("Org verified via API",
        "org_id", req.OrgID,
        "name", org.Name,
        "status", org.Status)

    return &VerifyOrgResult{
        Exists:    true,
        Name:      org.Name,
        Status:    org.Status,
        CreatedAt: org.CreatedAt,
    }, nil
}

// @temporal-gen activity
// @activity-queue "default"
func (v *Validator) VerifyOrgExistsInDB(ctx context.Context, req *VerifyOrgRequest) (*VerifyOrgResult, error) {
    logger := activity.GetLogger(ctx)

    logger.Info("Verifying org in database",
        "org_id", req.OrgID)

    var org app.Org
    result := v.db.WithContext(ctx).
        Where("id = ?", req.OrgID).
        First(&org)

    if result.Error != nil {
        if result.Error == gorm.ErrRecordNotFound {
            logger.Error("Org not found in database", "org_id", req.OrgID)
            return &VerifyOrgResult{
                Exists: false,
            }, nil
        }
        return nil, fmt.Errorf("database error: %w", result.Error)
    }

    // Verify account has admin role for this org
    var roleCount int64
    v.db.WithContext(ctx).
        Table("account_roles").
        Joins("JOIN roles ON account_roles.role_id = roles.id").
        Where("account_roles.account_id = ?", req.AccountID).
        Where("roles.org_id = ?", req.OrgID).
        Where("roles.role_type = ?", app.RoleTypeOrgAdmin).
        Count(&roleCount)

    hasAdminRole := roleCount > 0

    logger.Info("Org verified in database",
        "org_id", req.OrgID,
        "name", org.Name,
        "has_admin_role", hasAdminRole)

    return &VerifyOrgResult{
        Exists:       true,
        Name:         org.Name,
        Status:       org.Status,
        CreatedAt:    org.CreatedAt,
        HasAdminRole: hasAdminRole,
    }, nil
}
```

---

## 4. Activity: Cleanup Org

```go
// services/ctl-api/internal/app/canary/worker/activities/cleanup/delete_org.go

package cleanup

import (
    "context"
    "fmt"
    "time"

    "go.temporal.io/sdk/activity"
    "gorm.io/gorm"

    "github.com/nuonco/nuon/services/ctl-api/internal/app"
    orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
)

type Cleaner struct {
    db          *gorm.DB
    orgsHelpers *orgshelpers.Helpers
}

type DeleteOrgRequest struct {
    OrgID string `json:"org_id"`
}

type DeleteOrgResult struct {
    Deleted   bool   `json:"deleted"`
    Method    string `json:"method"`  // "hard_delete" or "force_delete"
    ErrorMsg  string `json:"error_msg,omitempty"`
}

// @temporal-gen activity
// @activity-queue "default"
func (c *Cleaner) DeleteTestOrg(ctx context.Context, req *DeleteOrgRequest) (*DeleteOrgResult, error) {
    logger := activity.GetLogger(ctx)

    logger.Info("Deleting test org", "org_id", req.OrgID)

    // Use hard delete helper to completely remove test org
    if err := c.orgsHelpers.HardDeleteOrg(ctx, req.OrgID); err != nil {
        logger.Error("Failed to hard delete org",
            "org_id", req.OrgID,
            "error", err)

        return &DeleteOrgResult{
            Deleted:  false,
            Method:   "hard_delete",
            ErrorMsg: err.Error(),
        }, nil
    }

    logger.Info("Successfully deleted test org", "org_id", req.OrgID)

    return &DeleteOrgResult{
        Deleted: true,
        Method:  "hard_delete",
    }, nil
}

// @temporal-gen activity
// @activity-queue "default"
func (c *Cleaner) VerifyOrgDeleted(ctx context.Context, req *DeleteOrgRequest) (bool, error) {
    logger := activity.GetLogger(ctx)

    var count int64
    result := c.db.WithContext(ctx).
        Model(&app.Org{}).
        Where("id = ?", req.OrgID).
        Count(&count)

    if result.Error != nil {
        return false, fmt.Errorf("database error: %w", result.Error)
    }

    deleted := count == 0

    logger.Info("Verified org deletion status",
        "org_id", req.OrgID,
        "deleted", deleted)

    return deleted, nil
}
```

---

## 5. Workflow: Org Create Test

```go
// services/ctl-api/internal/app/canary/worker/tests/org_create_test.go

package tests

import (
    "fmt"
    "time"

    "go.temporal.io/sdk/temporal"
    "go.temporal.io/sdk/workflow"

    "github.com/nuonco/nuon/pkg/types/workflows/canary"
    "github.com/nuonco/nuon/services/ctl-api/internal/app/canary/worker/activities/cli"
    "github.com/nuonco/nuon/services/ctl-api/internal/app/canary/worker/activities/cleanup"
    "github.com/nuonco/nuon/services/ctl-api/internal/app/canary/worker/activities/validation"
)

type OrgCreateTestRequest struct {
    CanaryID    string `json:"canary_id"`
    APIToken    string `json:"api_token"`
    AccountID   string `json:"account_id"`
    Environment string `json:"environment"`
}

// @temporal-gen workflow
// @execution-timeout 10m
// @task-timeout 2m
// @task-queue "default"
// @namespace "canary"
func OrgCreateTest(ctx workflow.Context, req *OrgCreateTestRequest) (*canary.TestResult, error) {
    logger := workflow.GetLogger(ctx)
    startTime := workflow.Now(ctx)

    result := &canary.TestResult{
        Scenario:  "org_lifecycle",
        Passed:    false,
        StartTime: startTime,
    }

    // Configure activity options
    activityOpts := workflow.ActivityOptions{
        StartToCloseTimeout: 2 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval:    1 * time.Second,
            BackoffCoefficient: 2.0,
            MaximumInterval:    30 * time.Second,
            MaximumAttempts:    3,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, activityOpts)

    // Generate unique org name with timestamp
    timestamp := workflow.Now(ctx).Unix()
    orgName := fmt.Sprintf("test-org-%s-%d", req.CanaryID, timestamp)

    logger.Info("Starting org create test", "org_name", orgName)

    // Step 1: Create org via CLI
    var createResult *cli.CommandResult
    err := workflow.ExecuteActivity(ctx, "RunCLICommand", &cli.CommandRequest{
        Command: []string{"orgs", "create", orgName},
        Timeout: 1 * time.Minute,
    }).Get(ctx, &createResult)

    if err != nil {
        result.Error = fmt.Sprintf("Failed to execute CLI command: %v", err)
        result.Duration = workflow.Now(ctx).Sub(startTime)
        return result, nil
    }

    if !createResult.Success {
        result.Error = fmt.Sprintf("CLI command failed (exit %d): %s",
            createResult.ExitCode,
            createResult.Output)
        result.Duration = workflow.Now(ctx).Sub(startTime)
        return result, nil
    }

    logger.Info("Org create command succeeded",
        "output", createResult.Output,
        "duration", createResult.Duration)

    // Step 2: Parse org ID from output
    var parsedOutput *cli.ParsedOrgOutput
    err = workflow.ExecuteActivity(ctx, "ParseOrgCreateOutput", createResult.Output).Get(ctx, &parsedOutput)

    if err != nil {
        result.Error = fmt.Sprintf("Failed to parse org ID from output: %v", err)
        result.Duration = workflow.Now(ctx).Sub(startTime)
        return result, nil
    }

    orgID := parsedOutput.OrgID
    logger.Info("Parsed org ID from output", "org_id", orgID, "org_name", parsedOutput.OrgName)

    // Step 3: Verify org exists via API
    var apiVerifyResult *validation.VerifyOrgResult
    err = workflow.ExecuteActivity(ctx, "VerifyOrgExistsInAPI", &validation.VerifyOrgRequest{
        OrgID:     orgID,
        OrgName:   orgName,
        AccountID: req.AccountID,
    }).Get(ctx, &apiVerifyResult)

    if err != nil {
        result.Error = fmt.Sprintf("Failed to verify org via API: %v", err)
        result.Duration = workflow.Now(ctx).Sub(startTime)
        // Still attempt cleanup
        workflow.ExecuteActivity(ctx, "DeleteTestOrg", &cleanup.DeleteOrgRequest{OrgID: orgID})
        return result, nil
    }

    if !apiVerifyResult.Exists {
        result.Error = "Org not found via API after creation"
        result.Duration = workflow.Now(ctx).Sub(startTime)
        return result, nil
    }

    logger.Info("Org verified via API", "org_id", orgID)

    // Step 4: Verify org exists in database
    var dbVerifyResult *validation.VerifyOrgResult
    err = workflow.ExecuteActivity(ctx, "VerifyOrgExistsInDB", &validation.VerifyOrgRequest{
        OrgID:     orgID,
        OrgName:   orgName,
        AccountID: req.AccountID,
    }).Get(ctx, &dbVerifyResult)

    if err != nil {
        result.Error = fmt.Sprintf("Failed to verify org in database: %v", err)
        result.Duration = workflow.Now(ctx).Sub(startTime)
        // Still attempt cleanup
        workflow.ExecuteActivity(ctx, "DeleteTestOrg", &cleanup.DeleteOrgRequest{OrgID: orgID})
        return result, nil
    }

    if !dbVerifyResult.Exists {
        result.Error = "Org not found in database after creation"
        result.Duration = workflow.Now(ctx).Sub(startTime)
        return result, nil
    }

    if !dbVerifyResult.HasAdminRole {
        result.Error = "Account does not have admin role for created org"
        result.Duration = workflow.Now(ctx).Sub(startTime)
        // Still attempt cleanup
        workflow.ExecuteActivity(ctx, "DeleteTestOrg", &cleanup.DeleteOrgRequest{OrgID: orgID})
        return result, nil
    }

    logger.Info("Org verified in database with admin role", "org_id", orgID)

    // Step 5: Cleanup - Delete the test org
    var deleteResult *cleanup.DeleteOrgResult
    err = workflow.ExecuteActivity(ctx, "DeleteTestOrg", &cleanup.DeleteOrgRequest{
        OrgID: orgID,
    }).Get(ctx, &deleteResult)

    if err != nil {
        result.Error = fmt.Sprintf("Failed to delete test org: %v", err)
        result.Duration = workflow.Now(ctx).Sub(startTime)
        return result, nil
    }

    if !deleteResult.Deleted {
        result.Error = fmt.Sprintf("Failed to delete test org: %s", deleteResult.ErrorMsg)
        result.Duration = workflow.Now(ctx).Sub(startTime)
        return result, nil
    }

    logger.Info("Test org deleted", "org_id", orgID)

    // Step 6: Verify org was deleted
    var isDeleted bool
    err = workflow.ExecuteActivity(ctx, "VerifyOrgDeleted", &cleanup.DeleteOrgRequest{
        OrgID: orgID,
    }).Get(ctx, &isDeleted)

    if err != nil {
        result.Error = fmt.Sprintf("Failed to verify org deletion: %v", err)
        result.Duration = workflow.Now(ctx).Sub(startTime)
        return result, nil
    }

    if !isDeleted {
        result.Error = "Org still exists in database after deletion"
        result.Duration = workflow.Now(ctx).Sub(startTime)
        return result, nil
    }

    logger.Info("Verified org was deleted", "org_id", orgID)

    // All steps passed!
    result.Passed = true
    result.Duration = workflow.Now(ctx).Sub(startTime)
    result.Details = map[string]interface{}{
        "org_id":              orgID,
        "org_name":            orgName,
        "cli_output":          createResult.Output,
        "cli_duration":        createResult.Duration.String(),
        "api_verification":    apiVerifyResult,
        "db_verification":     dbVerifyResult,
        "deletion_method":     deleteResult.Method,
    }

    logger.Info("Org create test completed successfully",
        "org_id", orgID,
        "total_duration", result.Duration)

    return result, nil
}
```

---

## 6. Example Test Execution

### Local Test Run

```bash
# Terminal 1: Start canary worker
cd services/ctl-api
go run main.go worker --namespace canary

# Terminal 2: Trigger test
curl -X POST http://localhost:8081/v1/general/canary/test/org-create \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "canary_id": "test-local-123",
    "environment": "local"
  }'
```

### Expected Temporal UI View

```
Workflow: OrgCreateTest
ID: canary-org-create-test-local-123
Status: ✅ Completed
Duration: 8.3 seconds

Activities:
  1. RunCLICommand                [✅ 1.2s] - nuon orgs create test-org-123-1738166400
  2. ParseOrgCreateOutput         [✅ 0.1s] - Extracted: orgrok933tcyzji01s7us3aeo3
  3. VerifyOrgExistsInAPI         [✅ 2.1s] - Found org in API
  4. VerifyOrgExistsInDB          [✅ 1.8s] - Found org in DB with admin role
  5. DeleteTestOrg                [✅ 2.4s] - Hard deleted org
  6. VerifyOrgDeleted             [✅ 0.7s] - Confirmed org deleted

Result:
{
  "scenario": "org_lifecycle",
  "passed": true,
  "duration": "8.3s",
  "details": {
    "org_id": "orgrok933tcyzji01s7us3aeo3",
    "org_name": "test-org-123-1738166400",
    "cli_output": "✅ Created organization: test-org-123-1738166400\nOrganization ID: orgrok933tcyzji01s7us3aeo3",
    "cli_duration": "1.2s",
    "api_verification": {
      "exists": true,
      "name": "test-org-123-1738166400",
      "status": "active",
      "created_at": "2026-01-29T16:00:00Z"
    },
    "db_verification": {
      "exists": true,
      "has_admin_role": true
    },
    "deletion_method": "hard_delete"
  }
}
```

---

## 7. Example Failure Scenarios

### Failure 1: CLI Command Failed

```
Activity: RunCLICommand
Status: ✅ Completed (but command failed)
Duration: 1.1s

Result:
{
  "command": ["orgs", "create", "test-org-123-1738166400"],
  "output": "Error: unauthorized - invalid API token",
  "exit_code": 1,
  "success": false,
  "duration": "1.1s",
  "error_msg": "exit status 1"
}

Test Result:
{
  "scenario": "org_lifecycle",
  "passed": false,
  "error": "CLI command failed (exit 1): Error: unauthorized - invalid API token",
  "duration": "1.1s"
}
```

### Failure 2: Org Not Found After Creation

```
Activity: RunCLICommand
Status: ✅ Completed
Duration: 1.2s
Result: Success

Activity: ParseOrgCreateOutput
Status: ✅ Completed
Duration: 0.1s
Result: org_id = orgrok933tcyzji01s7us3aeo3

Activity: VerifyOrgExistsInAPI
Status: ✅ Completed
Duration: 2.3s
Result: exists = false

Test Result:
{
  "scenario": "org_lifecycle",
  "passed": false,
  "error": "Org not found via API after creation",
  "duration": "3.6s"
}
```

### Failure 3: Org Not Deleted Properly

```
Activity: RunCLICommand through DeleteTestOrg
Status: ✅ All completed successfully
Duration: 7.6s

Activity: VerifyOrgDeleted
Status: ✅ Completed
Duration: 0.8s
Result: isDeleted = false

Test Result:
{
  "scenario": "org_lifecycle",
  "passed": false,
  "error": "Org still exists in database after deletion",
  "duration": "8.4s"
}
```

---

## 8. Debugging Failed Tests

### View in Temporal UI

```
1. Open http://localhost:8234 (local) or https://temporal-canary-web.nuon.co (prod)
2. Navigate to namespace: canary
3. Find workflow: OrgCreateTest
4. Click on failed workflow run
5. View activity results and errors
6. Check workflow history for detailed execution flow
```

### Query Test Results from Database

```sql
-- Get latest org create test results
SELECT
    id,
    schedule_name,
    started_at,
    duration_ms,
    passed_tests,
    failed_tests,
    results->>'org_lifecycle' as org_test_result
FROM canary_scheduled_runs
WHERE results ? 'org_lifecycle'
ORDER BY started_at DESC
LIMIT 10;

-- Get all failed org create tests
SELECT
    ctr.scenario,
    ctr.error_message,
    ctr.details,
    ctr.created_at,
    csr.schedule_name
FROM canary_test_results ctr
JOIN canary_scheduled_runs csr ON ctr.run_id = csr.id
WHERE ctr.scenario = 'org_lifecycle'
  AND ctr.passed = false
ORDER BY ctr.created_at DESC;
```

### View Metrics in DataDog

```
# Test pass rate
avg:canary.test.passed{scenario:org_lifecycle} by {environment}

# Test duration
avg:canary.test.duration{scenario:org_lifecycle} by {environment}

# Test failure count
sum:canary.test.failures{scenario:org_lifecycle} by {environment}
```

---

## 9. Extending the Test

### Add More Verification Steps

```go
// Step: Verify org has default settings
var settingsResult *validation.VerifyOrgSettingsResult
err = workflow.ExecuteActivity(ctx, "VerifyOrgDefaultSettings", &validation.VerifyOrgSettingsRequest{
    OrgID: orgID,
}).Get(ctx, &settingsResult)

// Step: Verify org roles were created
var rolesResult *validation.VerifyOrgRolesResult
err = workflow.ExecuteActivity(ctx, "VerifyOrgRolesCreated", &validation.VerifyOrgRolesRequest{
    OrgID: orgID,
}).Get(ctx, &rolesResult)

// Expected roles: OrgAdmin, Installer, Runner
if len(rolesResult.Roles) != 3 {
    result.Error = fmt.Sprintf("Expected 3 roles, got %d", len(rolesResult.Roles))
    result.Duration = workflow.Now(ctx).Sub(startTime)
    return result, nil
}
```

### Add Update/Rename Test

```go
// Step: Update org name via CLI
newOrgName := fmt.Sprintf("%s-updated", orgName)
var updateResult *cli.CommandResult
err = workflow.ExecuteActivity(ctx, "RunCLICommand", &cli.CommandRequest{
    Command: []string{"orgs", "update", orgID, "--name", newOrgName},
    Timeout: 1 * time.Minute,
}).Get(ctx, &updateResult)

if !updateResult.Success {
    result.Error = fmt.Sprintf("Failed to update org: %s", updateResult.Output)
    // Still attempt cleanup...
    return result, nil
}

// Verify name was updated
var verifyUpdateResult *validation.VerifyOrgResult
err = workflow.ExecuteActivity(ctx, "VerifyOrgExistsInAPI", &validation.VerifyOrgRequest{
    OrgID:   orgID,
    OrgName: newOrgName,
}).Get(ctx, &verifyUpdateResult)

if verifyUpdateResult.Name != newOrgName {
    result.Error = fmt.Sprintf("Org name not updated: expected %s, got %s",
        newOrgName, verifyUpdateResult.Name)
    return result, nil
}
```

---

## 10. Complete Test Output Example

```json
{
  "test_result": {
    "scenario": "org_lifecycle",
    "passed": true,
    "start_time": "2026-01-29T16:00:00Z",
    "end_time": "2026-01-29T16:00:08Z",
    "duration": "8.3s",
    "error": null,
    "details": {
      "org_id": "orgrok933tcyzji01s7us3aeo3",
      "org_name": "test-org-123-1738166400",
      "cli_output": "✅ Created organization: test-org-123-1738166400\nOrganization ID: orgrok933tcyzji01s7us3aeo3\nName: test-org-123-1738166400\nStatus: active\nCreated: 2026-01-29 16:00:00 UTC",
      "cli_duration": "1.2s",
      "api_verification": {
        "exists": true,
        "name": "test-org-123-1738166400",
        "status": "active",
        "created_at": "2026-01-29T16:00:00Z",
        "has_admin_role": true
      },
      "db_verification": {
        "exists": true,
        "name": "test-org-123-1738166400",
        "status": "active",
        "created_at": "2026-01-29T16:00:00Z",
        "has_admin_role": true
      },
      "deletion_method": "hard_delete",
      "deletion_verified": true,
      "steps_executed": [
        {
          "step": "RunCLICommand",
          "duration": "1.2s",
          "success": true
        },
        {
          "step": "ParseOrgCreateOutput",
          "duration": "0.1s",
          "success": true
        },
        {
          "step": "VerifyOrgExistsInAPI",
          "duration": "2.1s",
          "success": true
        },
        {
          "step": "VerifyOrgExistsInDB",
          "duration": "1.8s",
          "success": true
        },
        {
          "step": "DeleteTestOrg",
          "duration": "2.4s",
          "success": true
        },
        {
          "step": "VerifyOrgDeleted",
          "duration": "0.7s",
          "success": true
        }
      ]
    }
  }
}
```

---

This example shows the complete flow from CLI command execution to verification and cleanup!
