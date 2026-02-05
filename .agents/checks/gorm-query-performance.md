---
name: GORM Query Performance
description: Reviews GORM queries for performance anti-patterns like N+1 queries, unbounded loads, and missing optimizations
severity-default: high
tools: [Grep, Read, glob, Bash]
globs: ["services/ctl-api/**/*.go", "pkg/**/*.go"]
ast-grep-hints: [gorm-n-plus-one, gorm-unbounded-preload, gorm-find-without-where]
---

When reviewing GORM queries, check for these performance anti-patterns. Use ast-grep rules as **hints only** (may have false positives)—verify each finding manually.

## Run ast-grep for Initial Hints

```bash
sg scan --rule rules/11-gorm-n-plus-one.yml --rule rules/09-gorm-unbounded-preload.yml --rule rules/10-gorm-find-without-where.yml <path>
```

## N+1 Query Detection

**Pattern**: Database calls inside loops cause N+1 queries.

```go
// ❌ BAD: N+1 query - one query per iteration
for _, item := range items {
    var related Related
    db.Where("item_id = ?", item.ID).First(&related)
}

// ✅ GOOD: Batch query with WHERE IN
itemIDs := make([]string, len(items))
for i, item := range items {
    itemIDs[i] = item.ID
}
var allRelated []Related
db.Where("item_id IN ?", itemIDs).Find(&allRelated)
// Then build a map for O(1) lookups
```

**What to look for:**
- `db.Find()`, `db.First()`, `db.Where()...Find()` inside `for` loops
- Preload operations that could be batched
- Recursive functions that query on each call

## Unbounded Preloads

**Pattern**: `.Preload("Association")` without limits can load massive datasets.

```go
// ❌ BAD: Could load thousands of related records
db.Preload("Builds").Find(&apps)

// ✅ GOOD: Scope the preload with limits or ordering
db.Preload("Builds", func(db *gorm.DB) *gorm.DB {
    return db.Order("created_at DESC").Limit(10)
}).Find(&apps)

// ✅ GOOD: Use Joins for single record associations
db.Joins("LatestBuild").Find(&apps)
```

**When to flag:**
- Preloading has-many associations without scope function
- Preloading associations that grow unboundedly (builds, deploys, logs)
- Nested preloads: `.Preload("A.B.C")`

## Unbounded Find/All Queries

**Pattern**: `Find()` without `Where()`, `Limit()`, or pagination loads entire tables.

```go
// ❌ BAD: Loads all records
var orgs []Org
db.Find(&orgs)

// ✅ GOOD: Add filtering or pagination
db.Where("status = ?", "active").Limit(100).Find(&orgs)
```

## Missing Indexes for Query Patterns

Cross-reference queries with model indexes:

```go
// If you see this query pattern:
db.Where("org_id = ? AND status = ?", orgID, status).Find(&items)

// The model should have a composite index:
func (i *Item) Indexes(db *gorm.DB) []migrations.Index {
    return []migrations.Index{
        {Name: "idx_items_org_status", Columns: []string{"org_id", "status"}},
    }
}
```

## Select Only Needed Columns

**Pattern**: Loading full records when only specific fields are needed.

```go
// ❌ BAD: Loads all columns including large JSONB fields
var builds []Build
db.Where("app_id = ?", appID).Find(&builds)

// ✅ GOOD: Select only needed columns
var builds []Build
db.Select("id", "status", "created_at").Where("app_id = ?", appID).Find(&builds)
```

## Inefficient Count Queries

```go
// ❌ BAD: Loads all records just to count
var items []Item
db.Find(&items)
count := len(items)

// ✅ GOOD: Use Count()
var count int64
db.Model(&Item{}).Where("org_id = ?", orgID).Count(&count)
```

## Transaction Scope Issues

```go
// ❌ BAD: Multiple queries outside transaction that should be atomic
db.Create(&parent)
db.Create(&child) // If this fails, parent is orphaned

// ✅ GOOD: Use transaction
db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&parent).Error; err != nil {
        return err
    }
    return tx.Create(&child).Error
})
```

## Reporting Format

For each finding, report:
1. **File:Line** - Location of the issue
2. **Pattern** - Which anti-pattern was detected
3. **ast-grep hint** - If flagged by ast-grep rule (note: may be false positive)
4. **Current code** - The problematic code snippet
5. **Suggested fix** - Recommended improvement with code example
6. **Impact** - Estimated performance impact (e.g., "O(n) queries → O(1)")
