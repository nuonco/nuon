---
name: GORM Query Path Optimality
description: Verifies that queries and Temporal activity calls use the most direct relationship path through the GORM data model instead of multi-step lookups
severity-default: high
tools: [Grep, Read, glob]
globs: ["services/ctl-api/**/*.go", "pkg/**/*.go"]
---

When reviewing queries or Temporal workflow/activity code, verify that data is fetched via the most direct GORM relationship path rather than through indirect or multi-step lookups.

## How to Trace Relationship Chains

1. **Find the root model struct** in `services/ctl-api/internal/app/`. Look for struct fields with:
   - GORM association tags: `gorm:"constraint:OnDelete:CASCADE;"`, `gorm:"foreignKey:..."`
   - Foreign key fields: fields ending in `ID` (e.g., `AppConfigID string`) paired with a struct field of the referenced type (e.g., `AppConfig AppConfig`)
   - Nested associations: struct fields that themselves have further GORM associations

2. **Map the full chain** from your starting model to the data you need. Example chain:
   ```
   ComponentBuild
     → ComponentConfigConnection (belongs-to via ComponentConfigConnectionID)
       → AppConfig (belongs-to via AppConfigID)
         → PoliciesConfig (has-one, AppPoliciesConfig with AppConfigID FK)
           → Policies (has-many, []AppPolicyConfig)
   ```

3. **Check if existing `Preload()` calls already load the chain**. A single query with chained preloads is cheaper than multiple activity calls:
   ```go
   // ✅ GOOD: Single query loads entire chain
   db.Preload("ComponentConfigConnection").
      Preload("ComponentConfigConnection.AppConfig").
      Preload("ComponentConfigConnection.AppConfig.PoliciesConfig").
      Preload("ComponentConfigConnection.AppConfig.PoliciesConfig.Policies").
      First(&build, "id = ?", buildID)

   // ❌ BAD: Two separate lookups (especially as Temporal activities)
   build := GetComponentBuild(buildID)
   policies := GetPoliciesConfigByAppConfigID(build.ComponentConfigConnection.AppConfigID)
   ```

## Red Flags to Look For

### 1. Chained Activity Calls Where Output Feeds Input

If the second activity's only input comes from the first activity's output, the two can almost always be collapsed:

```go
// ❌ RED FLAG: Second call depends entirely on first call's output
build, _ := activities.AwaitGetComponentBuild(ctx, req{ID: buildID})
appConfigID := build.ComponentConfigConnection.AppConfigID
policies, _ := activities.AwaitGetPoliciesConfigByAppConfigID(ctx, req{AppConfigID: appConfigID})
```

**Fix**: Add preloads to the first query so the second call is unnecessary.

### 2. Fetching "Latest" When a Pinned FK Exists

If the model has a direct foreign key to the related record, do not re-derive the relationship by querying for the latest record:

```go
// ❌ RED FLAG: Re-deriving via ORDER BY when FK is available
db.Preload("Component.App.AppConfigs", func(db *gorm.DB) *gorm.DB {
    return db.Order("created_at DESC").Limit(1)
}).First(&build, "id = ?", buildID)
appConfigID := build.Component.App.AppConfigs[0].ID

// ✅ GOOD: Use the pinned FK directly
// ComponentConfigConnection already has AppConfigID field
build.ComponentConfigConnection.AppConfigID  // Already set at creation time
```

This is also a **correctness bug**: the "latest" config may differ from the config the build was actually created under. The pinned FK captures the exact config version.

### 3. Navigating Through Collection Indexes

Accessing `SomeSlice[0]` after a sorted preload is fragile and indicates the query is taking an indirect path:

```go
// ❌ RED FLAG: Fragile index access
appConfigs := component.App.AppConfigs
if len(appConfigs) == 0 { return }
appConfigID := appConfigs[0].ID

// ✅ GOOD: Direct FK access
appConfigID := build.ComponentConfigConnection.AppConfigID
```

### 4. Multiple DB Queries in a Single Activity

If an activity makes multiple `db.First()` / `db.Find()` calls that traverse related models, consider whether a single query with preloads would suffice:

```go
// ❌ RED FLAG: Multiple queries in one activity
build := db.First(&build, buildID)
config := db.Where("id = ?", build.AppConfigID).First(&config)
policies := db.Where("app_config_id = ?", config.ID).Find(&policies)

// ✅ GOOD: Single query with preloads
db.Preload("ComponentConfigConnection.AppConfig.PoliciesConfig.Policies").
   First(&build, "id = ?", buildID)
```

## Verification Steps

When reviewing a query or activity call:

1. **Identify what data is ultimately needed** (e.g., "I need the policies for this build")
2. **Find the shortest relationship path** in the GORM models from the starting entity to the target data
3. **Check if existing preloads cover the path** — if not, add them
4. **Check if the query uses a pinned FK** or re-derives the relationship dynamically
5. **Count the number of Temporal activity round-trips** — each is expensive; collapse when possible
6. **Verify correctness** — does the query path return the exact related record, or could it return a different version?

## Concrete Example: Helm Build Policy Flow

**Before (suboptimal — two Temporal activity round-trips):**
```
Activity 1: GetComponentBuild(buildID)
  → Returns build with ComponentConfigConnection.AppConfigID
Activity 2: GetPoliciesConfigByAppConfigID(appConfigID)
  → Returns PoliciesConfig with Policies
```

**After (optimal — single activity with preloads):**
```
Activity 1: GetComponentBuild(buildID)
  → Preloads: ComponentConfigConnection.AppConfig.PoliciesConfig.Policies
  → All data available from build.ComponentConfigConnection.AppConfig.PoliciesConfig
```

The `GetComponentBuild` activity already preloads `ComponentConfigConnection.AppConfig.PoliciesConfig.Policies` — the second activity call was unnecessary.

## Reporting Format

For each finding, report:
1. **File:Line** — Location of the suboptimal query path
2. **Current path** — The multi-step lookup being used
3. **Optimal path** — The direct relationship chain through GORM models
4. **Model evidence** — The struct fields and FK relationships that prove the direct path exists
5. **Impact** — Number of eliminated activity calls or DB queries
6. **Correctness risk** — Whether the indirect path could return stale/wrong data
