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
- **Read-only operations** - Dashboard is for viewing data, not modifying it
- **Minimal dependencies** - Leverages Go's standard library and simple tooling
- **Dark theme** - Matches Nuon brand with custom Tailwind configuration

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

# 2. Generate Go code
cd services/ctl-api/internal/app/admin_dashboard
templ generate

# 3. Compiled Go files created automatically
# my_page_templ.go is generated and compiled with the rest of the app
```

**Important**: `.templ` files are source files; `_templ.go` files are generated artifacts.

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
services/ctl-api/internal/app/admin_dashboard/
├── service/              # HTTP handlers and business logic
│   ├── views/            # Templ templates (.templ files)
│   │   ├── layout.templ       # Base layout with navigation
│   │   ├── index.templ        # Homepage
│   │   ├── orgs.templ         # Organizations table
│   │   ├── org_detail.templ   # Organization detail page
│   │   └── *_templ.go         # Generated Go code (DO NOT EDIT)
│   ├── service.go        # Service struct, route registration
│   ├── index.go          # Index page handler
│   ├── orgs.go           # Organizations list handler
│   ├── org_detail.go     # Organization detail handler
│   └── livez.go          # Health check handler
├── components/           # templui components
│   ├── table/            # Table component
│   ├── card/             # Card component
│   ├── button/           # Button component
│   └── */                # Other installed components
├── assets/               # Static assets
│   ├── css/
│   │   ├── input.css     # Tailwind source (edit this)
│   │   └── output.css    # Compiled CSS (generated, do not edit)
│   └── favicon.svg       # Site favicon
└── utils/                # Helper utilities
    └── tailwind.go       # TwMerge utility for class merging
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

**Step 4: Generate Templ Files**

```bash
cd services/ctl-api/internal/app/admin_dashboard
templ generate
```

**Step 5: Format Go Code**

```bash
cd /path/to/nuon
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
# Regenerate templates
cd services/ctl-api/internal/app/admin_dashboard
templ generate

# Check for Go compilation errors
cd /path/to/nuon
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

The admin dashboard has **read-only** access to the ctl-api database via GORM:

```go
// Access via service.db
func (s *service) getData(ctx context.Context) (*app.Model, error) {
    var data app.Model
    res := s.db.WithContext(ctx).
        Where("condition = ?", value).
        Find(&data)
    return &data, res.Error
}
```

**Important**: Only use `SELECT` queries. Do not create, update, or delete records.

### Using ctl-api Models

Import models from `internal/app`:

```go
import "github.com/nuonco/nuon/services/ctl-api/internal/app"

// Available models:
// app.Org, app.Account, app.Install, app.App, etc.
```

## Best Practices

### Templ Development

1. **Always run `templ generate`** after editing `.templ` files
2. **Never edit `_templ.go` files** - they're auto-generated
3. **Use type-safe props** - define struct types for component props
4. **Extract complex logic** - use helper functions for conditional classes

### Handler Development

1. **Separate concerns** - HTTP logic in handlers, business logic in private methods
2. **Use proper error handling** - log errors, return user-friendly messages
3. **Follow naming conventions** - handlers are PascalCase, private methods are camelCase
4. **Limit database queries** - use reasonable limits (e.g., 100 items)

### Performance

1. **Server-side rendering is fast** - no client-side framework overhead
2. **Use database indexes** - queries should be fast even with large datasets
3. **Limit result sets** - use LIMIT and pagination for large tables
4. **Preload relationships wisely** - only preload what you need

### Security

1. **Read-only operations** - dashboard should never modify data
2. **Admin authentication** - routes should require admin auth (handled by ctl-api)
3. **No user input in queries** - use parameterized queries only
4. **Escape output** - templ handles this automatically

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

## Future Enhancements

**Potential additions**:
- Pagination for large datasets
- Search/filter functionality
- More detail pages (apps, installs, components)
- Export functionality (CSV/JSON)
- Real-time data updates (SSE or WebSocket)
- Metric visualizations
- Log viewing
- Deployment status monitoring

**Keep it simple**: Only add features when needed. The dashboard should remain lightweight and focused on operational needs.

## Troubleshooting

### Templ Generation Fails

**Issue**: `templ generate` returns errors

**Solutions**:
- Check `.templ` file syntax (missing brackets, imports)
- Ensure templ CLI is installed: `go install github.com/a-h/templ/cmd/templ@latest`
- Verify Go imports are correct

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

1. **Always read this file first** to understand architecture and patterns
2. **Use established patterns** for handlers, templates, and styling
3. **Run `templ generate`** after creating/editing `.templ` files
4. **Follow the two-method handler pattern** for consistency
5. **Use templui components** rather than building custom UI
6. **Maintain read-only operations** - no data modification
7. **Test locally** before considering the task complete
8. **Keep it simple** - avoid over-engineering

The admin dashboard is intentionally minimal and focused. Prefer simplicity over features.
