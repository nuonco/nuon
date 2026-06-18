# gormlint

Static analyzer for GORM queries. Catches silent correctness bugs, performance footguns, and convention violations at compile time using `golang.org/x/tools/go/analysis`.

## Analyzers

### `wheregormignored` — Correctness

Detects struct fields tagged `gorm:"-"` used in `Where()` clauses. GORM silently ignores these fields, so the query runs without the intended filter.

```go
// BAD: ComputedField has gorm:"-", this Where condition is silently dropped
db.Where(Install{ComputedField: "value"}).First(&install)

// GOOD: use a field that GORM maps to a column
db.Where(Install{OrgID: orgID}).First(&install)
```

### `wherezerovalue` — Correctness

Detects literal zero-value fields in `Where()` struct clauses. GORM silently ignores zero values (`0`, `""`, `false`), turning the condition into a no-op.

```go
// BAD: Age=0 is silently dropped, this loads ALL users
db.Where(User{Age: 0}).Find(&users)

// GOOD: use a pointer type or map-based Where
db.Where(map[string]interface{}{"age": 0}).Find(&users)
```

### `rawsqlwhere` — Convention

Detects raw SQL strings in `Where()` clauses. The codebase convention is struct-based Where for type safety and refactor safety.

```go
// BAD: raw SQL string — typos and renames break silently
db.Where("org_id = ?", orgID).First(&install)

// GOOD: struct-based Where — compiler catches field name issues
db.Where(Install{OrgID: orgID}).First(&install)
```

### `hardcodedtablename` — Convention

Detects raw SQL strings in `Joins()`, `Order()`, `Select()`, `Group()`, and `Having()`. Hardcoded table/view names bypass the view plugin's automatic routing and break when view versions change.

```go
// BAD: hardcoded view name breaks when view version changes (v1 → v2)
db.Joins("JOIN install_states_view_v1 ON ...").Find(&installs)

// GOOD: use views.TableOrViewName() for dynamic view routing
db.Joins("JOIN " + views.TableOrViewName(db, &app.InstallState{}, " ON ...")).Find(&installs)
```

### `unboundedpreload` — Performance

Detects `Preload()` calls without a scoping function containing `Limit()`. Without a limit, GORM loads ALL related records — with 10k related rows per parent, this kills performance.

```go
// BAD: loads every Order for every User
db.Preload("Orders").Find(&users)

// BAD: scoping function with Order but no Limit
db.Preload("Orders", func(db *gorm.DB) *gorm.DB {
    return db.Order("created_at DESC")
}).Find(&users)

// GOOD: scoping function with Limit
db.Preload("Orders", func(db *gorm.DB) *gorm.DB {
    return db.Order("created_at DESC").Limit(10)
}).Find(&users)
```

### `unboundedquery` — Performance

Detects `Find()` and `Scan()` calls without `Limit()` anywhere in the query chain. Without a limit, queries can return unbounded result sets.

```go
// BAD: returns every matching row
db.Where(User{OrgID: orgID}).Find(&users)

// GOOD: bounded result set
db.Where(User{OrgID: orgID}).Limit(100).Find(&users)

// GOOD: First/Last/Take implicitly limit to 1 row
db.Where(User{OrgID: orgID}).First(&user)
```

### `missingcontext` — Reliability

Detects GORM query chains starting from a struct field (like `s.db`) that lack `.WithContext(ctx)`. Without context, queries have no request tracing, timeout propagation, or cancellation support.

```go
// BAD: no context — query can't be cancelled, no tracing
s.db.Where(User{Name: "alice"}).First(&user)

// GOOD: context propagated for tracing and cancellation
s.db.WithContext(ctx).Where(User{Name: "alice"}).First(&user)
```

GORM hooks (`BeforeCreate`, `AfterQuery`, etc.) are excluded — their `tx *gorm.DB` parameter already carries context from the parent query.

## Usage

```bash
# Run all analyzers on ctl-api
go run ./bins/gormlint ./services/ctl-api/...

# Run a single analyzer
go run ./bins/gormlint -wheregormignored ./services/ctl-api/...

# JSON output (for CI or tooling)
go run ./bins/gormlint -json ./services/ctl-api/...

# Build and run
go build -o gormlint ./bins/gormlint
./gormlint ./services/ctl-api/...

# Show context lines around findings
./gormlint -c 3 ./services/ctl-api/...
```

## Sample Output

```
services/ctl-api/internal/app/accounts/helpers/update_user_journey_step.go:17:11: unbounded Preload without scoping function; add a func(db *gorm.DB) *gorm.DB with Limit() to prevent loading all related records
services/ctl-api/internal/app/auth/service/device_code.go:132:12: GORM query chain missing WithContext(); add .WithContext(ctx) for request tracing and cancellation support
services/ctl-api/internal/app/auth/service/device_code.go:142:8: GORM query chain missing WithContext(); add .WithContext(ctx) for request tracing and cancellation support
services/ctl-api/internal/app/slack/service/subscribe_modal.go:1506:9: raw SQL string in Joins(); use views.TableOrViewName() or association joins instead of raw SQL join strings
services/ctl-api/internal/app/accounts/service/get_auth_me.go:68:9: Find() without Limit() in query chain; consider adding Limit() to prevent loading all matching records
```

## Findings Summary (ctl-api)

As of the initial run on `services/ctl-api/internal/`:

| Analyzer | Findings | Notes |
|----------|----------|-------|
| `wheregormignored` | 0 | No current violations (the motivating bug was fixed) |
| `wherezerovalue` | 0 | No current violations |
| `rawsqlwhere` | 0 | Where-specific (see `hardcodedtablename` for other SQL strings) |
| `hardcodedtablename` | 73 | Select: 30, Order: 21, Joins: 12, Group: 10 |
| `unboundedquery` | 47 | Find: 23, Scan: 24 |
| `unboundedpreload` | 22 | 20 missing scoping func, 2 scoping without Limit |
| `missingcontext` | 12 | Concentrated in `auth/service/` |

## Testing

```bash
go test ./bins/gormlint/analyzers/...
```

Each analyzer has its own test suite using `analysistest.Run` with `// want` comment directives in test fixtures under `testdata/src/example/`.

## Architecture

Built on `golang.org/x/tools/go/analysis` using `multichecker.Main()`, which provides:

- Per-analyzer enable/disable flags
- `-json` output for CI integration
- `-fix` for auto-fixable issues (future)
- Package pattern arguments (`./...` glob support)
- Parallel analysis across packages
