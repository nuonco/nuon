# Admin Dashboard Service

The **Admin Dashboard** is an internal administrative interface for the Nuon platform, providing read-only views into the system's data for debugging, monitoring, and operational purposes.

## Service Overview

This is a **server-side rendered (SSR)** web interface built with:
- **Go + Gin** - HTTP server and routing
- **Templ** - Type-safe Go templating engine (https://templ.guide/)
- **TailwindCSS** - Utility-first CSS framework
- **templui** - Component library for Templ (https://templui.io/)

The dashboard is designed to be **simple, fast, and maintainable** without the complexity of a full frontend framework.

## Architecture Philosophy

**Key Principles**:
- **Server-side rendering** - No JavaScript frameworks, pure SSR with Templ
- **Component-based** - Reusable templui components for UI consistency
- **Direct database access** - Use GORM directly, never HTTP API calls
- **Admin operations** - Primarily viewing data, with write operations for admin tools
- **Minimal dependencies** - Leverages Go's standard library and simple tooling
- **Dark theme** - Matches Nuon brand with custom Tailwind configuration

## ⚠️ CRITICAL WARNINGS FOR AI ASSISTANTS

### 1. NEVER Run `templ generate` Manually

**DO NOT EVER run `templ generate` in this directory.** The templ CLI has path issues that will break the generated imports.

- ❌ `templ generate` - **NEVER DO THIS**
- ✅ Edit `.templ` files and let the build process handle generation
- The `_templ.go` files are auto-generated during the build/compilation process
- Running `templ generate` manually will create broken import paths that are difficult to fix

### 2. DO NOT Add Unnecessary Comments

**Stop adding obvious comments everywhere.** The code should be self-explanatory.

❌ **Bad - Unnecessary comments**:
```go
// Get the organization from context
org := ctx.Org

// Query the database for the organization
result := s.db.Where("id = ?", orgID).First(&org)

// Return the organization
return org
```

✅ **Good - Clean code without noise**:
```go
var org app.Org
if err := s.db.Where("id = ?", orgID).First(&org).Error; err != nil {
    return nil, err
}
return &org, nil
```

**When comments ARE appropriate**:
- Complex business logic that isn't obvious
- Non-obvious workarounds or edge cases
- Important architectural decisions
- Public API documentation

**When comments are NOT needed** (most of the time):
- Obvious variable assignments
- Standard CRUD operations
- Self-explanatory function calls
- Anything the code itself clearly expresses

## Technology Stack

### Templating: Templ

Templ (https://templ.guide/) is a templating language for Go that compiles to Go code:

**Key Features**:
- Type-safe templates with Go syntax
- Compile-time checking (no runtime template parsing errors)
- Component composition with `{ children... }` pattern
- Automatic escaping and security
- Props via Go structs

**Workflow**:
```bash
# 1. Edit .templ files
vim service/views/my_page.templ

# 2. Build the application (generates _templ.go files automatically)
go build ./services/ctl-api/...

# The _templ.go files are auto-generated during compilation
```

**CRITICAL**: DO NOT run `templ generate` manually - it will break import paths. Let the build process handle generation.

### Component Library: templui

templui (https://templui.io/) provides pre-built, accessible UI components:

**Installed Components** (in `/components/` directory):
- `table/` - Data tables with headers, rows, cells
- `card/` - Container cards with header, content, footer
- `button/` - Action buttons
- `icon/` - Icon system
- `badge/` - Status badges
- `copybutton/` - Copy-to-clipboard functionality
- `tooltip/` - Hover tooltips
- Additional components as needed

**Installation**:
```bash
cd services/ctl-api/internal/app/admin_dashboard
templui install <component-name>
```

Components are installed locally and can be customized if needed.

### Styling: TailwindCSS

**Custom Theme** (`assets/css/input.css`):
- Dark theme with Nuon brand colors
- Custom color palette (cyan, purple, success/error/warning/info states)
- Monospace font (JetBrains Mono) for technical data
- Sans-serif font (Inter) for UI text

**Theme Colors**:
```css
--cyan: #00FFFF (brand color, used for links and accents)
--purple: #9333EA (secondary brand color)
--success: Green tones (for active/success states)
--error: Red tones (for error states)
--warning: Yellow tones (for warning states)
--info: Blue tones (for info states)
```

**Build Process**:
```bash
# Compile Tailwind CSS
cd services/ctl-api/internal/app/admin_dashboard
npx tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --watch
```

## Project Structure

```
services/ctl-api/internal/app/admin-dashboard/
├── service/              # HTTP handlers and business logic
│   ├── views/            # Templ templates (42 .templ files + generated _templ.go)
│   │   ├── layout.templ                         # Base layout with navigation
│   │   ├── index.templ                          # Homepage
│   │   ├── orgs.templ / orgs_table.templ        # Org list + HTMX table component
│   │   ├── org_detail.templ                     # Org detail with graph
│   │   ├── accounts.templ / accounts_table.templ
│   │   ├── account_detail.templ                 # Account + audit logs
│   │   ├── account_audit_logs_table.templ       # HTMX audit log table
│   │   ├── account_installs_table.templ
│   │   ├── installs.templ / installs_table*.templ
│   │   ├── install_detail.templ                 # Install detail with activity
│   │   └── *_templ.go                           # Generated Go code (DO NOT EDIT)
│   ├── service.go                    # Service struct, FX injection, route registration
│   ├── index.go                      # Homepage handler
│   ├── orgs.go / orgs_table.go       # Org list + HTMX polling endpoint
│   ├── org_detail.go                 # Org detail + graph generation
│   ├── org_status.go                 # Status badge polling
│   ├── org_tags.go                   # Tag management (UPDATE/DELETE)
│   ├── org_support_users.go          # Support user management
│   ├── accounts.go / accounts_table.go
│   ├── account_detail.go             # Account detail + UNION activity queries
│   ├── account_audit_logs_table.go
│   ├── account_installs_table.go
│   ├── installs.go                   # Global installs list
│   ├── installs_table_global.go / installs_table.go
│   ├── install_detail.go             # Install detail with nested UNION queries
│   ├── install_component_status.go   # Component status badge
│   ├── install_runner_status.go      # Runner status badge
│   ├── install_sandbox_status.go     # Sandbox status badge
│   ├── install_drift_status.go       # Drift status badge
│   └── livez.go                      # Health check
├── components/           # templui components (table, card, button, badge, icon, etc.)
├── assets/               # Static assets
│   ├── css/
│   │   ├── input.css     # Tailwind source (edit this)
│   │   └── output.css    # Compiled CSS (generated, do not edit)
│   └── favicon.svg
└── utils/
    └── templui.go        # TwMerge utility for class merging
```

## Development Patterns

### Creating a New Page

**Step 1: Create Template**

Create a new `.templ` file in `service/views/`:

```templ
// service/views/my_page.templ
package views

import (
    "github.com/nuonco/nuon/services/ctl-api/internal/app"
    "github.com/nuonco/nuon/services/ctl-api/internal/app/admin_dashboard/components/card"
)

templ MyPage(data *app.SomeModel) {
    @Layout("My Page Title") {
        <div class="w-full max-w-4xl mx-auto p-6">
            @card.Card() {
                @card.Header() {
                    @card.Title() {
                        My Page
                    }
                }
                @card.Content() {
                    <p>{ data.Name }</p>
                }
            }
        }
    }
}
```

**Step 2: Create Handler**

Create a handler file in `service/`:

```go
// service/my_page.go
package service

import (
    "context"
    "fmt"
    "net/http"

    "github.com/a-h/templ"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    "github.com/nuonco/nuon/services/ctl-api/internal/app"
    "github.com/nuonco/nuon/services/ctl-api/internal/app/admin_dashboard/service/views"
)

// Public handler - handles HTTP concerns
func (s *service) MyPage(c *gin.Context) {
    data, err := s.getMyData(c)
    if err != nil {
        s.l.Error("failed to get data", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
        return
    }

    component := views.MyPage(data)
    templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

// Private method - handles business logic
func (s *service) getMyData(ctx context.Context) (*app.SomeModel, error) {
    var data app.SomeModel

    res := s.db.WithContext(ctx).
        Where("some_condition = ?", true).
        First(&data)

    if res.Error != nil {
        return nil, fmt.Errorf("unable to get data: %w", res.Error)
    }

    return &data, nil
}
```

**Step 3: Register Route**

Add route in `service/service.go`:

```go
func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
    api.Static("/assets", "./internal/app/admin_dashboard/assets")

    api.GET("/", s.Index)
    api.GET("/my-page", s.MyPage)  // Add new route

    s.l.Info("admin-dashboard routes registered")
    return nil
}
```

**Step 4: Build Application**

```bash
# DO NOT run templ generate manually
# Just build the application and it will generate the Go files automatically
go build ./services/ctl-api/...
```

**Step 5: Format Go Code**

```bash
go fmt ./services/ctl-api/internal/app/admin_dashboard/...
```

### Handler Pattern

**Two-Method Pattern** (recommended for pages with database queries):

```go
// Public handler - HTTP concerns only
func (s *service) PageName(c *gin.Context) {
    data, err := s.getPageData(c)
    if err != nil {
        s.l.Error("failed to get data", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error message"})
        return
    }

    component := views.PageName(data)
    templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

// Private method - business logic
func (s *service) getPageData(ctx context.Context) (*app.Model, error) {
    // Database queries, data processing, etc.
    return data, nil
}
```

**Single-Method Pattern** (for simple pages):

```go
func (s *service) SimplePage(c *gin.Context) {
    component := views.SimplePage("Some static data")
    templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
```

### Templ Component Patterns

**Using templui Components**:

```templ
// Import components you need
import (
    "github.com/nuonco/nuon/services/ctl-api/internal/app/admin_dashboard/components/card"
    "github.com/nuonco/nuon/services/ctl-api/internal/app/admin_dashboard/components/table"
)

templ MyComponent() {
    @card.Card() {
        @card.Header() {
            @card.Title() { Title Text }
            @card.Description() { Description text }
        }
        @card.Content() {
            @table.Table() {
                @table.Header() {
                    @table.Row() {
                        @table.Head() { Column 1 }
                        @table.Head() { Column 2 }
                    }
                }
                @table.Body() {
                    @table.Row() {
                        @table.Cell() { Data 1 }
                        @table.Cell() { Data 2 }
                    }
                }
            }
        }
    }
}
```

**Passing Props to Components**:

```templ
// Components accept props via structs
@table.Cell(table.CellProps{Class: "font-mono text-xs"}) {
    { data.ID }
}

@card.Card(card.Props{Class: "max-w-2xl"}) {
    { children... }
}
```

**Conditional Rendering**:

```templ
templ ConditionalExample(showContent bool) {
    if showContent {
        <p>Content is visible</p>
    } else {
        <p>Content is hidden</p>
    }
}
```

**Loops**:

```templ
templ LoopExample(items []string) {
    <ul>
        for _, item := range items {
            <li>{ item }</li>
        }
    </ul>
}
```

**Helper Functions**:

```templ
// Define helper functions in the same file
templ StatusBadge(status app.OrgStatus) {
    <span class={ statusClass(status) }>
        { string(status) }
    </span>
}

// Go function for dynamic class generation
func statusClass(status app.OrgStatus) string {
    switch status {
    case app.OrgStatusActive:
        return "bg-success text-success-foreground"
    case app.OrgStatusError:
        return "bg-error text-error-foreground"
    default:
        return "bg-muted text-muted-foreground"
    }
}
```

### HTMX Two-Handler Pattern

Every list page uses two handlers: a **page handler** (full layout) and a **table handler** (HTMX polling, returns only the table component). The table handler reuses the same private data method:

```go
// Page handler — full layout
func (s *service) Orgs(c *gin.Context) {
    search, tagFilters, page := parseOrgParams(c)
    orgs, totalPages, err := s.getOrgs(c.Request.Context(), search, tagFilters, page)
    // ...
    component := views.Orgs(orgs, totalPages, page, search, tagFilters)
    templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

// Table handler — HTMX polling, returns table component only (no layout)
func (s *service) OrgsTable(c *gin.Context) {
    search, tagFilters, page := parseOrgParams(c)
    orgs, totalPages, err := s.getOrgs(c.Request.Context(), search, tagFilters, page)
    // ...
    component := views.OrgsTable(orgs, totalPages, page, search, tagFilters)
    templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
```

In the template, the table div uses `hx-get` to poll the table endpoint and `hx-include` to pass filter params:

```templ
<div hx-get="/orgs/table"
     hx-trigger="change from:[name='tag']"
     hx-target="#orgs-table"
     hx-swap="outerHTML"
     hx-include="[name='search'],[name='tag']">
```

JavaScript in the template listens for `htmx:afterSwap` to update the browser URL with `history.replaceState()`, enabling bookmarking and back-button behavior.

### Pagination Pattern

All list queries use consistent pagination:

```go
const perPageConst = 8

func (s *service) getOrgs(ctx context.Context, search string, tagFilters []string, page int) ([]*app.Org, int, error) {
    query := s.db.WithContext(ctx).Model(&app.Org{})

    // Apply filters first, then count
    var totalCount int64
    query.Count(&totalCount)

    totalPages := int(math.Ceil(float64(totalCount) / float64(perPageConst)))
    if totalPages == 0 {
        totalPages = 1  // Always min 1 page
    }

    offset := (page - 1) * perPageConst
    query.Order("created_at desc").Limit(perPageConst).Offset(offset).Find(&results)
}
```

**Rules:**
- Always at least 1 page (even for empty results)
- Search matches both name (`ILIKE '%term%'`) and exact ID (`id = ?`)
- Tag filters use PostgreSQL `&&` array overlap operator (OR logic): `WHERE tags && CAST(? AS text[])`

### Aggregation / Count Queries

Use embedded count subqueries with a wrapper struct, then convert back to the base type:

```go
type OrgWithCounts struct {
    app.Org
    AppCount     int `gorm:"column:app_count"`
    InstallCount int `gorm:"column:install_count"`
}

var results []OrgWithCounts
s.db.Select("orgs.*, " +
    "(SELECT COUNT(*) FROM apps WHERE apps.org_id = orgs.id AND apps.deleted_at IS NULL) as app_count, " +
    "(SELECT COUNT(*) FROM installs WHERE installs.org_id = orgs.id) as install_count").
    Find(&results)

// Convert wrapper → base type (copy count into base model field)
orgs := make([]*app.Org, len(results))
for i := range results {
    results[i].Org.AppCount = results[i].AppCount
    orgs[i] = &results[i].Org
}
```

### Parallel Fetching Pattern

Detail pages fetch multiple data sources in parallel using `errgroup`:

```go
func (s *service) OrgDetail(c *gin.Context) {
    orgID := c.Param("id")

    var org *app.Org
    var installs []*app.Install
    var recentApp *app.App

    g, gCtx := errgroup.WithContext(c.Request.Context())
    g.Go(func() error { var err error; org, err = s.getOrg(gCtx, orgID); return err })
    g.Go(func() error { var err error; installs, _, err = s.getInstallsForOrg(gCtx, orgID, 1); return err })
    g.Go(func() error { var err error; recentApp, err = s.getMostRecentApp(gCtx, orgID); return err })
    if err := g.Wait(); err != nil {
        // handle error
        return
    }
}
```

### UNION-Based Activity Log Queries

Activity logs across multiple entity types use raw SQL UNION ALL with dynamic query building:

```go
var queries []string
var queryParams []interface{}

for _, entityType := range entityTypes {
    switch entityType {
    case "runner_job":
        queries = append(queries, `SELECT id, 'runner_job' as entity_type, ... FROM runner_jobs WHERE install_id = ?`)
        queryParams = append(queryParams, installID)
    case "workflow":
        queries = append(queries, `SELECT id, 'workflow' as entity_type, ... FROM install_workflows WHERE install_id = ?`)
        queryParams = append(queryParams, installID)
    }
}

query := strings.Join(queries, " UNION ALL ") + " ORDER BY created_at DESC"
countQuery := "SELECT COUNT(*) FROM (" + query + ") as entries"

s.db.Raw(query+" LIMIT ? OFFSET ?", append(queryParams, limit, offset)...).Scan(&entries)
```

Default entity types if none specified (e.g., `["runner_job", "workflow"]`). Date range defaults to 30 days. Use `Unscoped()` when the query should include soft-deleted records.

### Styling Patterns

**Layout Containers**:
```html
<!-- Full-width page with max-width constraint -->
<div class="w-full max-w-7xl mx-auto p-6">
    <!-- Content -->
</div>

<!-- Narrower content (for detail pages) -->
<div class="w-full max-w-4xl mx-auto p-6">
    <!-- Content -->
</div>
```

**Technical Text Styling**:
```html
<!-- IDs, timestamps, technical data -->
<span class="font-mono text-xs">{ id }</span>
<span class="font-mono text-sm">{ timestamp }</span>

<!-- Uppercase type/status labels -->
<span class="font-mono text-xs uppercase">{ type }</span>
```

**Status Badges**:
```html
<span class="px-2 py-1 rounded text-xs font-mono uppercase bg-success-bg text-success border border-success-border">
    active
</span>
```

**Links**:
```html
<!-- Primary links (navigation) -->
<a href="/path" class="text-cyan hover:underline">Link Text</a>

<!-- Muted links (breadcrumbs) -->
<a href="/path" class="text-muted-foreground hover:text-cyan transition-colors">
    ← Back
</a>
```

**Spacing**:
- `p-6` - Standard padding for cards and containers
- `space-y-4` - Vertical spacing between elements
- `mb-6` - Bottom margin for sections

## Common Tasks

### Adding a New Component

```bash
# Install from templui
cd services/ctl-api/internal/app/admin_dashboard
templui install badge

# Component is now available at:
# components/badge/badge.templ
```

### Updating Tailwind Styles

```bash
# 1. Edit input.css
vim assets/css/input.css

# 2. Rebuild CSS (with watch mode for development)
npx tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --watch
```

### Debugging Templ Templates

**Common Issues**:

1. **Template not rendering**: Check that `templ generate` was run
2. **Compilation errors**: Check `.templ` file syntax (proper imports, brackets)
3. **Props not working**: Verify component prop struct fields match
4. **Styling not applied**: Ensure Tailwind classes are in output.css

**Debug Steps**:
```bash
# DO NOT run templ generate manually - it breaks imports
# Just build to regenerate templates automatically
go build ./services/ctl-api/...

# Rebuild CSS
npx tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css
```

### Running Locally

```bash
# Start full stack (includes ctl-api with admin dashboard)
cd /path/to/nuon
./run-nuonctl.sh dev --dev-all

# Or run just ctl-api
./run-nuonctl.sh dev --dev=ctl-api

# Access admin dashboard
# Default port: 8085
open http://localhost:8085/
```

**Default Routes**:
- `GET /` - Index/homepage
- `GET /livez` - Health check
- `GET /orgs` - Organizations table
- `GET /orgs/:id` - Organization detail page

## Design Guidelines

### UI Patterns

**Navigation**:
- Top navigation bar with site title and main links
- Breadcrumb navigation for detail pages
- Consistent link styling (cyan color, hover effects)

**Data Display**:
- Use tables for lists of items
- Use cards for detail views
- Monospace fonts for IDs and technical data
- Color-coded status badges for states

**Layout**:
- Center-aligned content with max-width constraints
- Consistent padding (p-6 for most containers)
- Responsive design (mobile-friendly)

### Color Usage

- **Cyan (`text-cyan`)**: Primary brand color, links, active states
- **Purple**: Secondary brand color (used sparingly)
- **Muted (`text-muted-foreground`)**: Secondary text, breadcrumbs
- **Status colors**: Success (green), error (red), warning (yellow), info (blue)

### Typography

- **Headers**: Bold, larger text for titles
- **Body**: Regular Inter font for readability
- **Technical**: JetBrains Mono for IDs, timestamps, code
- **Labels**: Muted color for form labels and metadata

## Integration with ctl-api

### Service Registration

The admin dashboard service is registered in ctl-api's main application:

**Location**: `services/ctl-api/admin.go`

The admin dashboard routes are mounted on the admin API server (separate from public/runner APIs).

**Port Configuration**:
- Production: Configured via environment variables
- Local development: Default port 8085

### Database Access

The admin dashboard has direct access to the ctl-api database via GORM.

**CRITICAL RULE: Always Use GORM Directly**

❌ **NEVER make HTTP API calls** from admin-dashboard handlers:
```go
// ❌ BAD - Do NOT do this
func (s *service) UpdateSomething(c *gin.Context) {
    // Making HTTP call to localhost API - FAILS IN PRODUCTION
    resp, err := http.Post("http://localhost:8081/v1/resource", ...)
}
```

✅ **ALWAYS use GORM directly** for database operations:
```go
// ✅ GOOD - Use GORM directly
func (s *service) UpdateSomething(c *gin.Context) {
    ctx := c.Request.Context()

    // Get data from database
    var data app.Model
    if err := s.db.WithContext(ctx).Where("id = ?", id).First(&data).Error; err != nil {
        // Handle error
        return
    }

    // Update directly in database
    data.Field = newValue
    if err := s.db.WithContext(ctx).Model(&data).Updates(&data).Error; err != nil {
        // Handle error
        return
    }
}
```

**Why This Matters:**
- HTTP calls to `localhost` work locally but **fail in production deployment**
- Admin-dashboard runs in the same process as ctl-api and has direct database access
- Direct GORM operations are faster and more reliable than HTTP calls
- Handlers should operate on the database layer, not the API layer

**Common Database Patterns:**

```go
// SELECT query
func (s *service) getData(ctx context.Context, id string) (*app.Model, error) {
    var data app.Model
    res := s.db.WithContext(ctx).
        Where("id = ?", id).
        First(&data)
    return &data, res.Error
}

// UPDATE operation
func (s *service) updateData(ctx context.Context, id string, newValue string) error {
    return s.db.WithContext(ctx).
        Model(&app.Model{}).
        Where("id = ?", id).
        Update("field", newValue).
        Error
}

// Multiple updates (use struct or map)
func (s *service) updateMultipleFields(ctx context.Context, model *app.Model) error {
    return s.db.WithContext(ctx).
        Model(model).
        Select("field1", "field2").  // Specify which fields to update
        Updates(model).
        Error
}
```

**Read-Write Operations:**
- Admin-dashboard CAN perform both read and write operations when needed
- Always use transactions for multi-step operations
- Use proper error handling and logging
- Consider the impact of write operations (they're admin-only tools)

### Using ctl-api Models

Import models from `internal/app`:

```go
import "github.com/nuonco/nuon/services/ctl-api/internal/app"

// Available models:
// app.Org, app.Account, app.Install, app.App, etc.
```

## Best Practices

### Templ Development

1. **NEVER run `templ generate` manually** - let `go build` handle generation
2. **Never edit `_templ.go` files** - they're auto-generated
3. **Use type-safe props** - define struct types for component props
4. **Extract complex logic** - use helper functions for conditional classes

### Handler Development

1. **NEVER use HTTP API calls** - Always use GORM directly for database operations
2. **Separate concerns** - HTTP logic in handlers, business logic in private methods
3. **Use proper error handling** - log errors, return user-friendly messages
4. **Follow naming conventions** - handlers are PascalCase, private methods are camelCase
5. **Limit database queries** - use reasonable limits (e.g., 100 items)

### Performance

1. **Server-side rendering is fast** - no client-side framework overhead
2. **Use database indexes** - queries should be fast even with large datasets
3. **Limit result sets** - use LIMIT and pagination for large tables
4. **Preload relationships wisely** - only preload what you need

### Security

1. **Admin-only operations** - All routes require admin authentication (handled by ctl-api)
2. **Use GORM directly** - Never make HTTP calls to localhost APIs
3. **Parameterized queries only** - Always use GORM's query builders with placeholders
4. **Escape output** - templ handles this automatically
5. **Careful with write operations** - Admin tools can modify data; consider impact

## Testing

### Manual Testing

```bash
# Start the service
./run-nuonctl.sh dev --dev=ctl-api

# Test routes
curl http://localhost:8085/livez
curl http://localhost:8085/orgs
curl http://localhost:8085/orgs/{org_id}

# Check in browser
open http://localhost:8085/
```

### Verification Checklist

- ✅ Page loads without errors
- ✅ Navigation links work
- ✅ Data displays correctly
- ✅ Styling matches theme
- ✅ No console errors in browser
- ✅ Responsive design works on mobile
- ✅ Tables are readable
- ✅ Status badges show correct colors

## Current Features

Pagination, search/filter, detail pages for orgs/accounts/installs, HTMX live-updating tables, status badge polling, tag management, support user management, and org dependency graphs are all implemented.

**Keep it simple**: Only add features when needed. The dashboard should remain lightweight and focused on operational needs.

## Troubleshooting

### Templ Generation Fails

**Issue**: `_templ.go` files have broken imports or compilation errors

**Solutions**:
- **DO NOT run `templ generate` manually** - it will break import paths
- Check `.templ` file syntax (missing brackets, imports)
- Verify Go imports are correct
- Let the build process handle generation: `go build ./services/ctl-api/...`

### Styles Not Applied

**Issue**: Tailwind classes not working

**Solutions**:
- Rebuild CSS: `npx tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css`
- Check class names (Tailwind v3 syntax)
- Verify CSS file is served correctly (check `/assets/css/output.css`)

### Page Not Found

**Issue**: Route returns 404

**Solutions**:
- Verify route is registered in `RegisterAdminDashboardRoutes()`
- Check handler method is exported (PascalCase)
- Restart service after adding routes

### Data Not Displaying

**Issue**: Template renders but data is missing

**Solutions**:
- Check database query in private method
- Verify model has data (check logs)
- Ensure template props match handler data
- Check for nil pointer dereferences

## Resources

- **Templ Documentation**: https://templ.guide/
- **templui Components**: https://templui.io/
- **TailwindCSS Docs**: https://tailwindcss.com/
- **Gin Framework**: https://gin-gonic.com/

## Notes for AI Assistants

When working on the admin dashboard:

1. **🚨 NEVER run `templ generate` manually** - It breaks import paths. Let the build process handle it.
2. **🚨 NEVER use HTTP API calls** - Always use GORM directly for all database operations
3. **🚨 DO NOT add unnecessary comments** - Code should be self-explanatory. Only comment non-obvious logic.
4. **Always read this file first** to understand architecture and patterns
5. **Use established patterns** for handlers, templates, and styling
6. **Follow the two-method handler pattern** for consistency
7. **Use templui components** rather than building custom UI
8. **Direct database access only** - Use `s.db` for all data operations
9. **Test locally** before considering the task complete
10. **Keep it simple** - avoid over-engineering

**Critical Reminders:**
- HTTP calls to `localhost` or the admin API will work in development but **fail in production**
- The admin-dashboard shares the same database connection as ctl-api and operates directly on the database layer using GORM
- Running `templ generate` manually creates broken import paths - never do it
- Stop adding obvious comments like "Get the org" or "Query the database" - the code is clear enough

The admin dashboard is intentionally minimal and focused. Prefer simplicity over features.
