# Testing Conventions

This document defines Go testing standards across the Nuon monorepo, with a focus
on `services/ctl-api` and shared `pkg/` code. **Postgres integration suites are the
default for database-backed code; in-memory sqlite is a deliberate, narrower tier —
not a substitute.**

## Quick Reference

| What you're testing | Tier | Pattern |
|---|---|---|
| Endpoints, migrations, constraints, cascades, gorm hooks | Postgres integration | `tests.BaseDBTestSuite` + `testseed` |
| Query shape / gorm logic with no Postgres-specific semantics | sqlite unit | in-memory sqlite, runs in plain `go test ./...` |
| Pure logic, no DB | plain funcs | testify; table-driven or `t.Run` prose subtests |
| HTTP handlers | httptest | `httptest.NewRecorder()` + router `ServeHTTP` |
| External HTTP APIs / SDK clients | httptest server or hand-rolled fake | — |
| Generated clients (temporal, etc.) | gomock | controller in `SetupSuite`, `Finish()` in `TearDownTest` |

---

## Database tests: two tiers

### Tier 1 — Postgres integration suites (the default)

Any test that depends on real database behavior — migrations, check constraints,
foreign-key cascades, soft-delete semantics through gorm hooks, HSTORE/JSONB — runs
against actual Postgres via the shared suite machinery in `services/ctl-api/tests/`:

```go
type MyServiceTestSuite struct {
    tests.BaseDBTestSuite
    // fx deps, router, etc.
}

func TestMyServiceSuite(t *testing.T) {
    tests.SkipIfNotIntegration(t) // skips unless INTEGRATION=true
    suite.Run(t, new(MyServiceTestSuite))
}
```

- **Gating**: every integration suite calls `tests.SkipIfNotIntegration(t)`. These
  tests only run when `INTEGRATION=true` and a migrated database is available (the
  separate `testsetup` binary runs migrations).
- **Wiring**: `tests.CtlApiFXOptions()` / `tests.CtlApiFXOptionsWithMocks(...)` for
  dependency injection; `tests.NewTestRouter(...)` for endpoint tests. The shared
  suite lives in a `suite_test.go` with per-endpoint test files beside it — see
  `services/ctl-api/internal/app/installs/service/suite_test.go` for the canonical
  layout.
- **Fixtures**: seed data through `services/ctl-api/tests/testseed`
  (`Build*` / `Create*` / `Ensure*` helpers — see its README). Do not hand-insert
  rows around it.
- **Isolation**: by unique names per test, not by truncating tables between tests.

#### Running integration suites locally

The fx graph validates the full ctl-api config at startup, so the suites need more
than database coordinates. A ready-made, secret-free environment is checked in at
`services/ctl-api/tests/integration.env` — local-dev values plus inert dummies for
required-but-unused fields. From `services/ctl-api`, with the dev dependency stack
running (Postgres + ClickHouse, via mono's `reset-dependencies`):

```bash
source tests/integration.env
go run ./cmd/nuontest    # once per schema change: creates + migrates ctl_api_test
go test -count=1 ./internal/pkg/account/...   # or any suite / ./...
```

Two things to know:

- The env file points at a dedicated `ctl_api_test` database. Never point it at
  `ctl_api` — `nuontest` drops and recreates the named database, and `ctl_api` is
  your live dev data.
- Real credentials never belong in `integration.env`. If a new required config
  field appears, add a dummy value there, not a secret.

### Tier 2 — in-memory sqlite unit tests

For small "does this query select the right rows" tests — typically worker
activities and query helpers — an in-memory sqlite database keeps the test in the
default `go test ./...` run with zero infrastructure:

```go
db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
```

Use this tier only when the logic under test does not depend on Postgres semantics.
Know its limits and be honest about them in the test:

- sqlite is not Postgres. Check constraints using Postgres functions
  (`char_length`), FK cascade behavior, and column type affinities are **not**
  exercised here. If an assertion's correctness rests on one of those, the test
  belongs in Tier 1.
- Model tags carry Postgres-only constructs that break `AutoMigrate` on sqlite,
  which is why existing tests hand-write `CREATE TABLE` DDL. Hand-written DDL is a
  drift risk against the models — keep it minimal, comment why it exists, and
  prefer sharing one schema helper per package over copying DDL between files.
- Seed with `db.Exec` inserts when gorm `BeforeCreate` hooks would fight the test
  (e.g. hooks that overwrite IDs), and say so in a comment.
- Mark all seed/setup helpers with `t.Helper()`.

---

## Assertions

testify everywhere; no stdlib-only assertion style.

- `require` for anything whose failure invalidates the rest of the test
  (especially `require.NoError`); `assert` for value comparisons. Both are commonly
  imported in the same file.
- In suites, `require.NoError(s.T(), err)` and `s.Require().NoError(err)` both
  appear; either is fine — be consistent within a file.

## Naming and structure

- Test files are named after the subject under test, not the package:
  `get_latest_runner_job_test.go`, `create_install_config_test.go`.
- Test functions are behavior sentences: `TestEnsureSecretIsIdempotent`,
  `TestJSONNeverEmitsSecretValues`. Subtest names are lowercase prose:
  `t.Run("an expired token is not a credential", ...)`.
- Both table-driven tests (slice of structs with `name`, setup closure,
  expected-`*` fields) and individually named `t.Run` subtests are established;
  pick whichever reads better for the case count.
- Tests and non-obvious helpers carry a comment explaining **why the test exists**
  — the failure it guards against, the customer impact, or the mechanism it pins
  down. This is a load-bearing house norm, not decoration.
- Never use real customer names in tests or fixtures.

## Mocking

Exactly three sanctioned styles — do not introduce a fourth (no mockery):

1. **gomock** (`github.com/golang/mock/gomock`) for generated clients:
   `s.ctrl = gomock.NewController(s.T())` in `SetupSuite`, `s.ctrl.Finish()` in
   `TearDownTest`; factor repeated expectations into suite helper methods.
2. **Hand-rolled fakes** for SDK boundaries (AWS, etc.): a struct implementing the
   interface plus call counters.
3. **`net/http/httptest`** for anything HTTP: `NewRecorder` + `router.ServeHTTP`
   for handlers (a `makeRequest(method, path, body)` suite helper is the norm),
   `NewServer` for SDK/client tests.

## Hermeticity

- Tests that read environment variables must clear or set every variable the code
  consults (`t.Setenv`), so a developer's shell or CI's own ambient credentials
  can't silently satisfy an assertion.
- Count requests or calls when asserting "this must not retry" — absence of
  behavior needs positive proof.

## Frontend

TypeScript/React testing in `services/dashboard-ui` follows that service's own
tooling (`bun run test`); see the service's CLAUDE.md. This document does not
define frontend conventions.
