---
name: No SQL Window Functions
description: Flags new SQL window functions in Postgres views and queries, which force Postgres to read the whole partition or table on every lookup
severity-default: high
tools: [Grep, Read, glob, Bash]
globs: ["services/ctl-api/**/*.sql", "services/ctl-api/**/*.go", "pkg/**/*.go"]
---

Do not add SQL window functions (`... OVER (...)`) to Postgres views or to queries that run on request, workflow, or worker paths. Flag every window function the change adds or moves into a new view version.

## Why

Postgres cannot push a filter on the outer query into a window function, except on its `PARTITION BY` columns. A view that numbers rows with a window function computes that window over every row it covers, even when the caller asks for one ID. The GORM views plugin routes every query and preload of a model with `UseView()` through its view, so one window function there slows every read of that model.

`component_config_connections_view_v3` joined to `SELECT id, row_number() OVER (PARTITION BY app_id ORDER BY created_at) FROM app_configs`. Each lookup sorted every app config in the database: about 125ms for one row on prod, and the slowest table on the install overview path. The v4 view replaced it with a correlated count scoped to the app and took the same lookup to about 6ms.

## How to find them

```bash
git diff origin/main...HEAD -U0 -- '*.sql' '*.go' | rg -i '^\+.*\b(over\s*\(|row_number|rank\s*\(|dense_rank|ntile|lag\s*\(|lead\s*\(|first_value|last_value|nth_value|percent_rank|cume_dist)'
```

Also check new view SQL files under `services/ctl-api/internal/pkg/db/viewsql/` in full, since a new view version is a new file.

## What to suggest instead

- **Per-row numbering or ordinals:** a correlated count scoped to an indexed column, such as `(SELECT count(*) FROM t t2 WHERE t2.parent_id = t.parent_id AND (t2.created_at, t2.id) <= (t.created_at, t.id))`. Confirm an index covers the scoping column and the ordering columns.
- **Values that never change after insert,** such as a version number: store them as a column when the row is written.
- **Latest row per group:** `DISTINCT ON (group_col) ... ORDER BY group_col, created_at DESC`, or a pinned foreign key to the latest row.
- **Running totals or counts:** compute them in Go over rows the query already loads.

## Do not flag

- One-shot data migrations under `services/ctl-api/internal/pkg/db/psql/migrations/` that do not create or replace a view.
- ClickHouse queries.
- Window functions that already exist and that the change does not touch.

## Reporting format

For each finding, report:
1. **File:Line** of the window function
2. **Where it runs:** view, request path, worker, or migration
3. **Current SQL**
4. **Suggested replacement** from the list above
5. **Index** the replacement relies on, and whether it exists
