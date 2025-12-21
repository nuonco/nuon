# Canvas View Visual Layout

## Page Structure

```
┌─────────────────────────────────────────────────────────────────────────┐
│ Breadcrumbs: Home > Apps > [App] > Branches > [Branch] > Canvas        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  Branch Name - Workflow Canvas                                          │
│  Horizontal visualization of the branch workflow pipeline               │
│                                                                           │
└─────────────────────────────────────────────────────────────────────────┘
```

## Horizontal Pipeline Visualization

```
┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐
│  Fetch Repository│ ───> │  Build Config    │ ───> │Build Changed     │ ───> │  Update Installs │
│                  │      │                  │      │  Components      │      │                  │
│  ✓ Completed     │      │  ✓ Completed     │      │  ◷ Running       │      │  ⏸ Pending       │
│                  │      │                  │      │                  │      │                  │
│  📝 abc1234      │      │                  │      │  📦 3 components │      │  ☁️ 5 installs   │
│                  │      │                  │      │     changed      │      │     affected     │
│                  │      │                  │      │                  │      │                  │
│  ⏱ 20 seconds    │      │  ⏱ 40 seconds    │      │  ⏳ Running...   │      │                  │
└──────────────────┘      └──────────────────┘      └──────────────────┘      └──────────────────┘
  Green border              Green border              Blue border              Gray border
  Green background          Green background          Blue background          Gray background
```

## Status Legend

```
┌─────────────────────────────────────────────────────────────────────────┐
│ Status Legend:                                                            │
│                                                                           │
│  ● Completed    ● Running    ● Pending    ● Failed                      │
└─────────────────────────────────────────────────────────────────────────┘
```

## Mock Data Notice

```
┌─────────────────────────────────────────────────────────────────────────┐
│ ℹ️  Mock Data Preview                                                    │
│                                                                           │
│ This is a preview with mock workflow data. In production, this will      │
│ display real workflow stages and their status for branch [branch-id]    │
└─────────────────────────────────────────────────────────────────────────┘
```

## Navigation Flow

```
Branch Detail Page
├── Workflows Section Header
│   └── "View Canvas" button (top right)
│       └── Navigates to: /[org]/apps/[app]/branches/[branch]/canvas
│
└── Canvas Page
    ├── Breadcrumbs (shows full navigation path)
    ├── Page Header
    ├── Horizontal Workflow Pipeline (scrollable)
    ├── Status Legend
    └── Mock Data Notice
```

## Stage Card Structure

```
┌─────────────────────────────────┐
│  Stage Name           [Status]  │  ← Header with status badge
├─────────────────────────────────┤
│                                 │
│  📝 Commit Hash                 │  ← Metadata section
│  📦 Components Changed          │     (conditional rendering)
│  ☁️ Installs Affected           │
│                                 │
├─────────────────────────────────┤
│  ⏱ Execution Time               │  ← Footer (if completed)
│  or                             │
│  ⏳ Running... (animated)        │  ← Footer (if running)
└─────────────────────────────────┘
```

## Connector Arrows

```
Stage 1  ──────>  Stage 2  ──────>  Stage 3  ──────>  Stage 4
         Active             Active             Inactive
         (Blue)             (Blue)             (Gray)
```

- Active arrows: Blue color (stage is completed or currently running)
- Inactive arrows: Gray color (stage hasn't started yet)

## Responsive Behavior

- Horizontal scroll enabled for smaller screens
- Each stage card: min-width 280px, max-width 280px
- Cards maintain consistent height
- Full pipeline visible on wide screens
- Scroll to see all stages on narrow screens

## Color Scheme

### Completed Stages
- Border: Green-400 / Green-500 (dark mode)
- Background: Green-50/30 / Green-950/20 (dark mode)

### Running Stages
- Border: Blue-400 / Blue-500 (dark mode)
- Background: Blue-50/30 / Blue-950/20 (dark mode)
- Animated pulse on icon

### Pending Stages
- Border: Cool-Grey-300 / Dark-Grey-600 (dark mode)
- Background: Cool-Grey-50/30 / Dark-Grey-800/20 (dark mode)

### Failed Stages
- Border: Red-400 / Red-500 (dark mode)
- Background: Red-50/30 / Red-950/20 (dark mode)
