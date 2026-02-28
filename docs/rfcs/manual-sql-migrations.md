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

**Runtime index diffing is fragile.** The `Indexes()` interface requires every model to declare its expected indexes in Go. On each deploy the system queries `pg_indexes`, diffs expected vs actual, and issues `CREATE INDEX` and `DROP INDEX` statements live. This couples deploy-time behavior to runtime introspection and makes index changes invisible in code review.

**No migration history for schema changes.** Unlike the custom data migrations in `all.go` (which are tracked in a `migrations` table), AutoMigrate and index changes leave no audit trail. There is no way to correlate a schema change to a specific deploy or commit.

**Deploy-time surprises.** Both AutoMigrate and index diffing execute DDL during the startup command. A schema change that takes 30 seconds locally can lock a production table for minutes. There is no circuit breaker, no rollback plan, and no preview of what will happen.

---

## Proposed Approach

Use **Atlas `schema diff`** as a **developer tool** (not a runtime dependency) to generate SQL migration statements. Paste the output into the existing `Migration` system in `all.go`.

### How it works

```
┌─────────────────────────────────────────────────────────┐
│                   Developer Workflow                     │
│                                                         │
│  1. Make model changes in Go structs                     │
│                                                         │
│  2. nuonctl database migration generate                  │
│     --name "094-install-add-new-field"                   │
│                                                         │
│  3. Review the generated SQL and scaffolded migration    │
│                                                         │
│  4. Commit, PR, code review, merge                       │
└─────────────────────────────────────────────────────────┘
```

### The migration pipeline after

The `Exec()` method in `exec.go` simplifies to:

| Step | What it does |
|------|-------------|
| 1. Join Tables | Unchanged (GORM M2M setup) |
| 2. Views | `CREATE OR REPLACE VIEW` (unchanged) |
| 3. Global Migrations | All DDL + data migrations, tracked and ordered |

AutoMigrate (step 2 today) and index diffing (step 3 today) are removed.

### What happens to `Indexes()` methods

The ~100 models with `Indexes()` implementations become dead code and are removed. Their index definitions are "frozen" into the base schema. Existing databases already have these indexes. New databases get them from the initial migration.

A one-time "baseline" migration captures the full current schema (tables + indexes) as the starting point for new environments.

### Examples

#### Adding a column

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
    Fn:   m.Migration094InstallAddNewField,
}

// 3. Implement the migration function
func (m *Migrator) Migration094InstallAddNewField(db *gorm.DB) error {
    return db.Exec(`ALTER TABLE installs ADD COLUMN new_field TEXT NOT NULL DEFAULT '';`).Error
}
```

#### Adding an index

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
// In all.go
{
    Name: "095-install-index-new-field",
    Fn:   m.Migration095InstallIndexNewField,
}

// Migration function
func (m *Migrator) Migration095InstallIndexNewField(db *gorm.DB) error {
    return db.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_installs_new_field ON installs (new_field, deleted_at);`).Error
}
```

Note: `CONCURRENTLY` is now possible because we control the exact SQL. The runtime index diffing could not use concurrent index creation.

---

## Atlas as a diffing tool

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
--exclude "_view_*"
```

**No `dev-url` required.** When diffing two live databases, Atlas inspects both directly. The `dev-url` flag is only needed when diffing against SQL or HCL files.

### Developer tooling: `nuonctl database migration`

Rather than a standalone script, the workflow is exposed as a new `nuonctl` subcommand:

```bash
# Generate a migration diff between current main and your local changes
nuonctl database migration generate --name "094-install-add-new-field"
```

Under the hood, this command:

1. Uses the local Postgres instance for development
2. Runs current `main` migrations against `base_db`
3. Runs current branch migrations against `dev_db`
4. Runs `atlas schema diff` between them (excluding views)
5. Scaffolds the migration function, appends the entry to `all.go`, and outputs the diff for the developer to review

This keeps Atlas as an implementation detail — developers interact with `nuonctl` and never need to install or invoke Atlas directly.

---

## Alternative considered: full Atlas provider

### What it is

The [Atlas GORM Provider](https://atlasgo.io/guides/orms/gorm) (`ariga.io/atlas-provider-gorm`) can introspect GORM models at compile time and generate Atlas migration files automatically.

The full Atlas approach would:

1. Replace `all.go` with an Atlas-managed versioned migration directory
2. Use `atlas migrate diff` with the GORM provider to auto-generate migrations from model changes
3. Use `atlas migrate apply` at deploy time instead of the custom `Migrator`
4. Manage the migration tracking table (`atlas_schema_revisions`) instead of the custom `migrations` table

### Why we are not doing this

**Views are a second-class citizen.** Atlas's GORM provider does not introspect `Views()` methods. Atlas supports views via a separate API, but view diffing in `schema diff` is a **Pro-tier feature** requiring authentication. Even with Pro, the `AlwaysReapply` flag (drop and recreate on every deploy) is Go-side logic that Atlas cannot represent.

**Index plugins do not map cleanly.** The `Indexes()` interface supports features that the GORM provider does not fully capture.

**Join table setup is GORM-specific.** The `JoinTables()` interface calls `db.SetupJoinTable()`, which is a GORM runtime operation. Atlas does not know about this.

### Summary of trade-offs

| Concern | Full Atlas | Atlas as Diff Tool (proposed) |
|---------|-----------|------------------------------|
| Views management | Unsupported (needs workaround) | Unchanged (`Views()` stays) |
| Index control | Limited by provider | Full SQL control |
| Migration tracking | Replace with `atlas_schema_revisions` | Keep existing `migrations` table |
| DataDog metrics | Rebuild | Unchanged |
| `AlwaysRun` migrations | Unsupported | Unchanged |
| Deploy safety (unique constraint) | Different mechanism | Unchanged |
| ClickHouse support | Not supported | Unchanged |
| Runtime dependency | Yes | No (dev-time only) |
| Migration effort | High | Low |

---

## Migration plan

### Phase 1: Baseline

1. Generate a complete schema dump of the current production database (tables + indexes, excluding views)
2. Store as a reference "initial migration"
3. For **new** environments, add a single bootstrap migration that creates the full schema

### Phase 2: Remove AutoMigrate

1. Remove `applyGormMigrations()` from `exec.go`
2. All new schema changes go through `all.go` as explicit SQL
3. GORM models remain the source of truth for Go code (reads/writes). They just no longer drive DDL.

### Phase 3: Remove index diffing

1. Remove `applyIndexes()` from `exec.go`
2. All new index changes go through `all.go` as explicit SQL

### Phase 4: Developer tooling

1. Add `nuonctl database migration` subcommand (CLI and web app)
2. Document the workflow in the wiki
3. Optionally add a CI check that GORM model changes have a corresponding migration entry

### Phase 5: Admin dashboard

1. Add a migrations view showing which migrations have been applied, when, and against which environment
2. Add the ability to reapply a specific migration
3. Add the ability to view the raw SQL for each migration

---

## Risks and mitigations

| Risk | Mitigation |
|------|-----------|
| Developer forgets to add migration for model change | CI check: if `internal/app/*.go` changes, require corresponding `all.go` entry |
| Generated SQL has issues | Developer reviews Atlas output before committing; PR review catches problems |
| New environment setup is slow with many migrations | Collapse old migrations periodically into a new baseline |
| Existing databases diverge from expected schema | One-time schema validation script to detect drift |

---

## Open questions

1. **Should we vendor Atlas or expect developers to install it?** Leaning toward bundling Atlas within the `nuonctl database migration` command (e.g., via Docker) so there is no separate install requirement.
2. **How do we handle the ClickHouse migration side?** ClickHouse has its own migrator with cluster-aware DDL. Same approach applies but with different SQL syntax. Atlas supports ClickHouse but the diff workflow would need a separate script.
3. **Should the baseline migration be a single large SQL file or broken up by model?** Single file is simpler for bootstrap; broken up is easier to reason about.
