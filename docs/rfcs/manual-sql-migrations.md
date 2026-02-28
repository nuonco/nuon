# RFC: Manual SQL Migrations via Atlas Diffing

**Status:** Draft
**Authors:** Nuon Platform Team
**Created:** 2026-02-28

---

## Summary

Replace GORM's `AutoMigrate` and the runtime index-diffing step with explicit, version-controlled SQL migrations generated using Atlas `schema diff`. Views and the existing migration tracking system remain unchanged.

---

## Problem

The current migration pipeline runs **five steps on every deploy** against every registered model (~106 models):

| Step | What it does | Risk |
|------|-------------|------|
| 1. Join Tables | `db.SetupJoinTable()` for M2M | Low |
| 2. AutoMigrate | Creates/alters tables from GORM struct tags | **High** |
| 3. Indexes | Diffs expected vs actual indexes, adds/removes | **Medium** |
| 4. Views | `CREATE OR REPLACE VIEW` | Low |
| 5. Custom Migrations | Tracked, named, one-shot data migrations | Low |

The problems concentrate in steps 2 and 3:

**AutoMigrate is a black box.** GORM inspects every struct, compares it to the live schema, and issues `ALTER TABLE` statements at deploy time. There is no review step, no dry-run, and no way to know what DDL will execute in production before it runs. A developer adding a `gorm:"notnull"` tag can unknowingly trigger a full-table rewrite on a million-row table.

**AutoMigrate is additive-only.** By design, GORM AutoMigrate will **never** drop columns, drop tables, rename columns, or change column types. If you remove a field from a struct, the column stays in the database forever. Rename a field and you get a new column while the old one becomes a ghost. This means:
- Schema drift accumulates over time (live DB has columns the code doesn't know about)
- Destructive migrations (dropping unused columns, renaming) are impossible without manual SQL
- There is no way to generate a complete diff — AutoMigrate can only tell you what to *add*, not what to remove or change

This is the core reason GORM itself cannot serve as a migration diffing tool. Atlas, which compares two full schemas and generates both `ADD` and `DROP`/`ALTER` statements, fills this gap.

**Runtime index diffing is fragile.** The `Indexes()` interface requires every model to declare its expected indexes in Go. On each deploy the system queries `pg_indexes`, diffs expected vs actual, and issues `CREATE INDEX` / `DROP INDEX` statements live. This couples deploy-time behavior to runtime introspection and makes index changes invisible in code review.

**No migration history for schema changes.** Unlike the custom data migrations in `all.go` (which are tracked in a `migrations` table), AutoMigrate and index changes leave no audit trail. There is no way to correlate a schema change to a specific deploy or commit.

**Deploy-time surprises.** Both AutoMigrate and index diffing execute DDL during the startup command. A schema change that takes 30 seconds locally can lock a production table for minutes. There is no circuit breaker, no rollback plan, and no preview of what will happen.

---

## Proposed Approach

Use **Atlas `schema diff`** as a **developer tool** (not a runtime dependency) to generate SQL migration statements. Paste the output into the existing `Migration` system in `all.go`.

### How It Works

```
┌─────────────────────────────────────────────────────────┐
│                   Developer Workflow                     │
│                                                         │
│  1. Spin up two local Postgres databases                │
│     - base_db: run current main migrations              │
│     - dev_db:  run current main migrations + changes    │
│                                                         │
│  2. atlas schema diff                                   │
│     --from "postgres://localhost/base_db"               │
│     --to   "postgres://localhost/dev_db"                │
│                                                         │
│  3. Review the generated SQL                            │
│                                                         │
│  4. Add to all.go as a new Migration{SQL: "..."}        │
│                                                         │
│  5. Commit, PR, code review, merge                      │
└─────────────────────────────────────────────────────────┘
```

### What Changes

| Component | Before | After |
|-----------|--------|-------|
| **Table creation/alteration** | GORM AutoMigrate (runtime) | Explicit SQL in `all.go` (tracked) |
| **Index management** | Runtime diffing via `Indexes()` methods | Explicit `CREATE INDEX` SQL in `all.go` (tracked) |
| **Views** | `CREATE OR REPLACE VIEW` via `Views()` | **No change** - stays as-is |
| **Data migrations** | `Migration{Fn/SQL}` in `all.go` | **No change** - stays as-is |
| **Migration tracking** | `migrations` table for custom only | `migrations` table for **all** DDL |
| **Atlas** | Not used | Dev-time diffing tool only |

### The Migration Pipeline After

The `Exec()` method in `exec.go` simplifies to:

| Step | What it does |
|------|-------------|
| 1. Join Tables | Unchanged (GORM M2M setup) |
| 2. Views | `CREATE OR REPLACE VIEW` - unchanged |
| 3. Global Migrations | All DDL + data migrations, tracked and ordered |

AutoMigrate (step 2 today) and index diffing (step 3 today) are removed.

### What Happens to `Indexes()` Methods

The ~100 models with `Indexes()` implementations become dead code and are removed. Their index definitions are "frozen" into the base schema — existing databases already have these indexes, and new databases get them from the initial migration.

A one-time "baseline" migration captures the full current schema (tables + indexes) as the starting point for new environments.

### Example: Adding a Column

**Today** (implicit):
```go
// Just add a field to the struct. AutoMigrate figures it out.
type Install struct {
    // ...
    NewField string `gorm:"notnull;default:''"`
}
```

**After** (explicit):
```go
// 1. Add the field to the struct (for GORM reads/writes)
type Install struct {
    // ...
    NewField string `gorm:"notnull;default:''"`
}

// 2. Add the migration to all.go
{
    Name: "094-install-add-new-field",
    SQL:  `ALTER TABLE installs ADD COLUMN new_field text NOT NULL DEFAULT '';`,
}
```

### Example: Adding an Index

**Today** (implicit):
```go
func (i *Install) Indexes(db *gorm.DB) []migrations.Index {
    return []migrations.Index{
        // add new entry here, runtime diffing picks it up
        {
            Name:    indexes.Name(db, &Install{}, "new_field"),
            Columns: []string{"new_field", "deleted_at"},
        },
    }
}
```

**After** (explicit):
```go
{
    Name: "095-install-index-new-field",
    SQL:  `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_installs_new_field ON installs (new_field, deleted_at);`,
}
```

Note: `CONCURRENTLY` is now possible because we control the exact SQL. The runtime index diffing could not use concurrent index creation.

---

## Atlas as a Diffing Tool

Atlas is used **only at development time** to generate the SQL. It is never a runtime dependency.

```bash
# Install (one-time)
curl -sSf https://atlasgo.sh | sh

# Generate diff between two database states
atlas schema diff \
  --from "postgres://localhost:5432/base_db?sslmode=disable" \
  --to   "postgres://localhost:5432/dev_db?sslmode=disable" \
  --exclude ".*_view_.*"
```

The `--exclude` flag uses glob patterns to skip database objects. Since views are managed by the existing `Views()` plugin, we exclude them from the diff. The pattern format is `schema.object[type=resource_type]`:

```bash
# Exclude all views from the diff
--exclude "public.*[type=view]"

# Or by naming convention
--exclude "*_view_*"
```

**No `dev-url` required.** When diffing two live databases, Atlas inspects both directly. The `dev-url` flag is only needed when diffing against SQL/HCL files.

### Scripting the Workflow

This can be wrapped in a script (e.g., `bins/nuonctl/scripts/generate-migration.sh`):

1. Start a fresh Postgres container
2. Run current `main` migrations against `base_db`
3. Run current branch migrations against `dev_db`
4. Run `atlas schema diff` between them
5. Print the SQL for the developer to review and add to `all.go`

---

## Alternative Considered: Full Atlas Provider

### What It Is

The [Atlas GORM Provider](https://atlasgo.io/guides/orms/gorm) (`ariga.io/atlas-provider-gorm`) can introspect GORM models at compile time and generate Atlas migration files automatically. The full Atlas approach would:

1. Replace `all.go` with Atlas-managed versioned migration directory
2. Use `atlas migrate diff` with the GORM provider to auto-generate migrations from model changes
3. Use `atlas migrate apply` at deploy time instead of the custom `Migrator`
4. Manage the migration tracking table (`atlas_schema_revisions`) instead of the custom `migrations` table

### Why We Are Not Doing This

**Views are a second-class citizen.** Atlas's GORM provider doesn't introspect `Views()` methods — it reads struct definitions. Atlas does support views via a separate `ViewDef` API, but view diffing in `schema diff` is a **Pro-tier feature** requiring authentication. Even with Pro, the `AlwaysReapply` flag (drop and recreate on every deploy) is pure Go-side logic that Atlas cannot represent. We would need to either:
- Maintain views as separate "post-migration" hooks outside Atlas (fragmenting the pipeline)
- Embed raw SQL views into Atlas migration files (losing the versioned `AlwaysReapply` behavior)
- Pay for Atlas Pro and still rewrite view management logic

None of these options are better than what we have today with 19+ SQL view files managed by the `Views()` plugin.

**Index plugins don't map cleanly.** The `Indexes()` interface supports features that the GORM provider doesn't fully capture:
- Partial indexes with `WHERE` clauses (`Option` field)
- Index type hints (`USING btree`, `USING gin`)
- Comment annotations
- The existing naming convention (`idx_{table}_{name}`) would need to be reconciled with Atlas's naming

**Join table setup is GORM-specific.** The `JoinTables()` interface calls `db.SetupJoinTable()`, which is a GORM runtime operation. Atlas doesn't know about this.

**Custom migration semantics don't translate.** Atlas has no equivalent for:
- `AlwaysRun` migrations with hourly deduplication (used for idempotent fixups)
- `Migration{Fn: ...}` callbacks that run arbitrary Go code (used for data backfills)
- `Migration{SQLFn: ...}` that generate SQL dynamically at runtime
- The `in_progress` / `applied` / `error` status model with DataDog event emission

These would all need to be reimplemented or abandoned.

**It replaces working infrastructure.** The existing `Migrator` handles:
- Migration tracking with `in_progress`/`applied`/`error` status
- `AlwaysRun` migrations with hourly deduplication
- DataDog metrics emission per migration
- Concurrent deploy safety via unique constraint on migration name
- Per-model custom data migrations via `Fn` callbacks
- ClickHouse migrations with cluster-aware DDL templates

Replacing all of this with Atlas means rebuilding or abandoning each of these capabilities.

**It's an all-or-nothing migration.** Moving to Atlas requires migrating the existing `migrations` tracking table to `atlas_schema_revisions`, establishing a baseline, and ensuring every existing database in production is compatible with the new tracking. The existing system works — the only problem is AutoMigrate and runtime index diffing.

### Summary of Trade-offs

| Concern | Full Atlas | Atlas as Diff Tool (Proposed) |
|---------|-----------|------------------------------|
| Views management | Unsupported - needs workaround | Unchanged - `Views()` plugin stays |
| Index control | Limited by GORM provider | Full SQL control |
| Migration tracking | Replace with `atlas_schema_revisions` | Keep existing `migrations` table |
| DataDog metrics | Rebuild | Unchanged |
| `AlwaysRun` migrations | Unsupported | Unchanged |
| Deploy safety (unique constraint) | Different mechanism | Unchanged |
| ClickHouse support | Not supported | Unchanged |
| Runtime dependency | Yes (`atlas migrate apply`) | No (dev-time only) |
| Migration to adopt | High (baseline + table migration) | Low (add entries to `all.go`) |

---

## Migration Plan

### Phase 1: Baseline

1. Generate a complete schema dump of the current production database (tables + indexes, excluding views)
2. Store as a reference "baseline" — not as a migration (existing databases already have this schema)
3. For **new** environments, add a single bootstrap migration that creates the full schema

### Phase 2: Remove AutoMigrate

1. Remove `applyGormMigrations()` from `exec.go`
2. All new schema changes go through `all.go` as explicit SQL
3. GORM models remain the source of truth for Go code (reads/writes) — they just no longer drive DDL

### Phase 3: Remove Index Diffing

1. Remove `applyIndexes()` from `exec.go`
2. Remove all `Indexes()` method implementations from models (~100 files)
3. All new index changes go through `all.go` as explicit SQL

### Phase 4: Developer Tooling

1. Add `generate-migration.sh` script to `bins/nuonctl/scripts/`
2. Document the workflow in the wiki
3. Optionally add CI check that GORM model changes have a corresponding migration entry

---

## Risks and Mitigations

| Risk | Mitigation |
|------|-----------|
| Developer forgets to add migration for model change | CI check: if `internal/app/*.go` changes, require corresponding `all.go` entry |
| Generated SQL has issues | Developer reviews Atlas output before committing; PR review catches problems |
| New environment setup is slow with many migrations | Collapse old migrations periodically into a new baseline |
| Existing databases diverge from expected schema | One-time schema validation script to detect drift |

---

## Open Questions

1. **Should we vendor Atlas or expect developers to install it?** Leaning toward a Docker-based approach in the script so there is no local install requirement.
2. **How do we handle the ClickHouse migration side?** ClickHouse has its own migrator with cluster-aware DDL. Same approach applies but with different SQL syntax. Atlas supports ClickHouse but the diff workflow would need a separate script.
3. **Should the baseline migration be a single large SQL file or broken up by model?** Single file is simpler for bootstrap; broken up is easier to reason about.
