# Admin Dashboard (HTMX)

Internal admin dashboard for Nuon operations. This is a server-rendered Go + templ + HTMX web app, completely separate from the React `dashboard-ui` SPA.

- **Port**: 8087 (configured via `admin_dashboard_http_port` in `internal/config.go`)
- **URL**: `http://localhost:8087` when running locally
- **Auth**: Admin middleware using `X-Nuon-Admin-Email` header (see `internal/middlewares/admin/`)

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Server | Go (Gin) |
| Templates | [templ](https://templ.guide) — type-safe Go HTML templates |
| Interactivity | [HTMX 2.0](https://htmx.org) — server-rendered HTML fragments |
| Styling | Tailwind CSS v4 with OKLCH color theme |
| Components | [templui](https://templui.com) — templ component library |
| Fonts | Inter + JetBrains Mono (Google Fonts) |

## Directory Structure

```
admin-dashboard/
├── service/              ← Go handlers (Gin)
│   ├── service.go        ← FX params, constructor, route registration
│   ├── views/            ← templ templates (pages + fragments)
│   │   ├── layout.templ  ← shared HTML shell, nav, ActivePage enum
│   │   ├── gen.go        ← package declaration (templ generate target)
│   │   └── *.templ       ← one per page/table/badge
│   └── *.go              ← one file per handler (+ private query methods)
├── components/           ← reusable templui primitives
│   ├── table/            ├── badge/
│   ├── pagination/       ├── copybutton/
│   ├── button/           ├── card/
│   ├── input/            ├── search/
│   ├── selectbox/        ├── tagsinput/
│   ├── dropdown/         ├── popover/
│   ├── tooltip/          ├── datepicker/
│   ├── calendar/         ├── graphviewer/
│   ├── status/           ├── progress/
│   ├── toast/            ├── icon/
│   ├── journey-progress/ └── aspectratio/
├── assets/               ← static files served at /assets
│   ├── css/
│   │   ├── input.css     ← Tailwind source with Nuon OKLCH theme
│   │   └── output.css    ← generated (do not edit directly)
│   ├── js/               ← component JS (minified)
│   └── favicon.svg
├── utils/                ← shared templ utilities (TwMerge)
└── .templui.json         ← templui config (component paths, module name)
```

## Handler Pattern

Every handler follows the same shape. Example from `orgs_table.go`:

```go
func (s *service) OrgsTable(c *gin.Context) {
    ctx := c.Request.Context()
    search := c.Query("search")
    page := getPageFromQuery(c)

    orgs, totalPages, err := s.getOrgs(ctx, search, nil, page)
    if err != nil {
        s.l.Error("failed to get orgs for table", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organizations"})
        return
    }

    component := views.OrgsTable(orgs, page, totalPages, search, filteredTags)
    templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
```

Private query methods (e.g. `getOrgs`) live in the same handler file.

## Page vs Fragment Pattern

**Full pages** wrap content in `views.Layout()` which provides the HTML shell, nav bar, and assets:

```templ
templ Orgs(...) {
    @Layout("Organizations", ActivePageOrgs) {
        <div class="max-w-7xl mx-auto px-6 py-8">
            // page content with embedded table div
        </div>
    }
}
```

**HTMX fragments** return bare HTML (no `Layout` wrapper) for `hx-swap`:

```go
func (s *service) OrgsTable(c *gin.Context) {
    // ... fetch data ...
    component := views.OrgsTable(orgs, page, totalPages, search, filteredTags)
    templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
```

The fragment template has no `@Layout` call — it returns just the table HTML that HTMX swaps in.

## HTMX Patterns Used

| Pattern | Example | Notes |
|---------|---------|-------|
| Polling | `hx-trigger="every 20s"` | Status badges auto-refresh |
| Swap | `hx-swap="outerHTML"` | Replace the triggering element |
| Include | `hx-include="[name='search'],[name='tag']"` | Send form values with request |
| Target | `hx-target="#table-container"` | Swap into a specific element |
| Trigger on change | `hx-trigger="change from:[name='tag']"` | React to filter changes |
| Redirect after POST | `c.Redirect(http.StatusSeeOther, "/orgs/"+orgID)` | PRG pattern for mutations |

The two-handler pattern is central: every list page has a **page handler** (full layout) and a **table handler** (bare fragment for HTMX polling/filtering). Both call the same private query method.

## Route Registration

All routes are registered in `service.go` → `RegisterAdminDashboardRoutes()`. The other four `Register*Routes` methods return nil — admin-dashboard only uses the dashboard route context.

## Available Service Dependencies

The service struct has access to these (injected via FX):

| Field | Type | Use |
|-------|------|-----|
| `db` | `*gorm.DB` (Postgres) | Primary data queries |
| `chDB` | `*gorm.DB` (ClickHouse) | Log stream queries |
| `temporalClient` | `temporalclient.Client` | Temporal workflow inspection |
| `queueClient` | `*queueclient.Client` | Queue operations |
| `appsHelpers` | `*appshelpers.Helpers` | App domain logic |
| `orgsHelpers` | `*orgshelpers.Helpers` | Org domain logic |
| `acctClient` | `*account.Client` | Account operations |
| `authzClient` | `*authz.Client` | RBAC operations |
| `codecs` | `[]converter.PayloadCodec` | Temporal payload decoding (gzip, large, s3) |
| `cfg` | `*internal.Config` | App config (e.g. `cfg.AppURL`) |
| `l` | `*zap.Logger` | Structured logging |
| `v` | `*validator.Validate` | Input validation |
| `mw` | `metrics.Writer` | Metrics |

## Recipes

### Adding a New Page

1. **Create the handler** in `service/<page_name>.go`:

   ```go
   func (s *service) MyPage(c *gin.Context) {
       ctx := c.Request.Context()
       data, err := s.getMyData(ctx)
       if err != nil {
           s.l.Error("failed to get data", zap.Error(err))
           c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
           return
       }
       component := views.MyPage(data)
       templ.Handler(component).ServeHTTP(c.Writer, c.Request)
   }
   ```

2. **Create the templ template** in `service/views/<page_name>.templ`:

   ```templ
   package views

   templ MyPage(data *app.SomeType) {
       @Layout("Page Title", ActivePageMyPage) {
           <div class="max-w-7xl mx-auto px-6 py-8">
               // page content
           </div>
       }
   }
   ```

3. **Add the ActivePage constant** in `service/views/layout.templ` if needed, and add a nav link in the `<nav>` section.

4. **Register the route** in `service.go` → `RegisterAdminDashboardRoutes()`:

   ```go
   api.GET("/my-page", s.MyPage)
   ```

### Adding a New Table Fragment

1. **Create the handler** in `service/<table_name>_table.go`:

   ```go
   func (s *service) MyTable(c *gin.Context) {
       ctx := c.Request.Context()
       page := getPageFromQuery(c)
       items, totalPages, err := s.getMyItems(ctx, page)
       if err != nil {
           s.l.Error("failed to get items", zap.Error(err))
           c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch items"})
           return
       }
       component := views.MyTable(items, page, totalPages)
       templ.Handler(component).ServeHTTP(c.Writer, c.Request)
   }
   ```

2. **Create the templ template** in `service/views/<table_name>_table.templ` — no `Layout()` wrapper:

   ```templ
   package views

   templ MyTable(items []*app.MyItem, page int, totalPages int) {
       <div id="my-table" hx-get="/my-page/table" hx-trigger="every 20s" hx-swap="outerHTML">
           @table.Table(...) {
               // table rows
           }
           @pagination.Pagination(...)
       </div>
   }
   ```

3. **Register the route**: `api.GET("/my-page/table", s.MyTable)`

4. **Reference from the page** templ: use `hx-get="/my-page/table"` to load the fragment.

### Adding a New POST Action

```go
func (s *service) DoAction(c *gin.Context) {
    ctx := c.Request.Context()
    id := c.Param("id")

    if err := c.Request.ParseForm(); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
        return
    }

    // Perform mutation via GORM...

    // Option A: Redirect (PRG pattern)
    c.Redirect(http.StatusSeeOther, "/my-page/"+id)

    // Option B: Return updated fragment for hx-swap
    component := views.UpdatedFragment(updatedData)
    templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
```

Register: `api.POST("/my-page/:id/action", s.DoAction)`

## Build & Code Generation

**Do not run `templ generate`, `tailwindcss`, or `go build` manually.** The `nuonctl` tooling handles all code generation and asset compilation. Just edit `.templ` and `.css` source files and let the build pipeline take care of the rest.

- `*_templ.go` files are generated from `.templ` files — never edit them directly.
- `assets/css/output.css` is generated from `assets/css/input.css` — never edit it directly.

## Key Rules

- **Always use GORM directly** — never make HTTP API calls from handlers. The dashboard shares a process with ctl-api and has direct DB access. HTTP calls to localhost work locally but fail in production.
- **Never edit `_templ.go` files** — they are generated from `.templ` files.
- **Use templui components** from `components/` rather than building custom UI primitives.
- **Follow the two-method handler pattern**: public handler for HTTP concerns, private method for DB queries.

## Current Pages

| Route | Handler | Description |
|-------|---------|-------------|
| `/` | `Index` | Dashboard home |
| `/orgs` | `Orgs` | Organization list with search + tag filters |
| `/orgs/table` | `OrgsTable` | HTMX fragment: org table |
| `/orgs/:id` | `OrgDetail` | Org detail with installs, tags, support users |
| `/orgs/:id/status` | `OrgStatus` | HTMX fragment: org status badge |
| `/orgs/:id/installs/table` | `InstallsTable` | HTMX fragment: org installs |
| `/accounts` | `Accounts` | Account list |
| `/accounts/table` | `AccountsTable` | HTMX fragment: account table |
| `/accounts/:id` | `AccountDetail` | Account detail with installs + audit logs |
| `/accounts/:id/installs/table` | `AccountInstallsTable` | HTMX fragment |
| `/accounts/:id/audit-logs/table` | `AccountAuditLogsTable` | HTMX fragment |
| `/installs` | `Installs` | Global install list |
| `/installs/table` | `InstallsTableGlobal` | HTMX fragment: global installs |
| `/installs/:id` | `InstallDetail` | Install detail with status, deployments, workflows |
| `/installs/:id/status/runner` | `InstallRunnerStatus` | HTMX fragment: runner status badge |
| `/installs/:id/status/sandbox` | `InstallSandboxStatus` | HTMX fragment |
| `/installs/:id/status/component` | `InstallComponentStatus` | HTMX fragment |
| `/installs/:id/status/drift` | `InstallDriftStatus` | HTMX fragment |
| `/installs/:id/active-deployments/table` | `InstallActiveDeploymentsTable` | HTMX fragment |
| `/installs/:id/activity/table` | `InstallActivityTable` | HTMX fragment |
| `/installs/:id/workflows/table` | `InstallWorkflowsTable` | HTMX fragment |
| `/workflows` | `Workflows` | Workflow list |
| `/workflows/table` | `WorkflowsTable` | HTMX fragment |
| `/workflows/:workflow_id` | `WorkflowDetail` | Workflow detail |
| `/queues` | `Queues` | Queue list |
| `/queues/table` | `QueuesTable` | HTMX fragment |
| `/queues/:id` | `QueueDetail` | Queue detail with emitters + signals |
| `/queues/:id/emitters/table` | `QueueEmittersTable` | HTMX fragment |
| `/queues/:id/signals/table` | `QueueSignalsTable` | HTMX fragment |
| `/queues/:id/signals/:signal_id` | `QueueSignalDetail` | Signal detail |
| `/queues/:id/emitters/:emitter_id` | `QueueEmitterDetail` | Emitter detail |
| `/queue-signals` | `QueueSignals` | Global signal list |
| `/queue-signals/table` | `QueueSignalsGlobalTable` | HTMX fragment |
| `/signal-catalog` | `SignalCatalog` | Signal type catalog |
| `/signal-catalog/:signal_type` | `SignalCatalogDetail` | Signal type detail |
| `/log-streams` | `LogStreamViewer` | Log stream viewer (ClickHouse) |
| `/log-streams/:log_stream_id` | `LogStreamDetail` | Log stream detail |
| `/log-streams/:log_stream_id/logs/table` | `LogStreamLogsTable` | HTMX fragment |
| `/temporal-workflows` | `TemporalWorkflowViewer` | Temporal workflow inspector |

### POST Routes (Mutations)

| Route | Handler | Description |
|-------|---------|-------------|
| `POST /orgs/:id/tags` | `UpdateOrgTags` | Update org tags (returns updated header fragment) |
| `POST /orgs/:id/tags/remove/:tag` | `RemoveSingleTag` | Remove single tag |
| `POST /orgs/:id/support-users/add` | `AddSupportUsers` | Add support users to org |
| `POST /queues/:id/restart` | `RestartQueue` | Restart a queue |

## Common Query Patterns

### Pagination

```go
const perPage = 8

func (s *service) getItems(ctx context.Context, page int) ([]*app.Item, int, error) {
    var totalCount int64
    query := s.db.WithContext(ctx).Model(&app.Item{})
    query.Count(&totalCount)

    totalPages := int(math.Ceil(float64(totalCount) / float64(perPage)))
    if totalPages == 0 {
        totalPages = 1
    }

    offset := (page - 1) * perPage
    var items []*app.Item
    query.Order("created_at desc").Limit(perPage).Offset(offset).Find(&items)
    return items, totalPages, nil
}
```

### Aggregation with Count Subqueries

```go
type OrgWithCounts struct {
    app.Org
    AppCount     int `gorm:"column:app_count"`
    InstallCount int `gorm:"column:install_count"`
}

var results []OrgWithCounts
s.db.Select("orgs.*, "+
    "(SELECT COUNT(*) FROM apps WHERE apps.org_id = orgs.id AND apps.deleted_at = 0) as app_count, "+
    "(SELECT COUNT(*) FROM installs WHERE installs.org_id = orgs.id AND installs.deleted_at = 0) as install_count").
    Find(&results)

orgs := make([]*app.Org, len(results))
for i := range results {
    results[i].Org.AppCount = results[i].AppCount
    orgs[i] = &results[i].Org
}
```

### Parallel Fetching (Detail Pages)

```go
g, gCtx := errgroup.WithContext(c.Request.Context())
g.Go(func() error { var err error; org, err = s.getOrg(gCtx, orgID); return err })
g.Go(func() error { var err error; installs, _, err = s.getInstallsForOrg(gCtx, orgID, 1); return err })
if err := g.Wait(); err != nil { ... }
```
