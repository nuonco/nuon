# Visual Guide: Canvas Features

## Feature 1: Side Panel for Step Details

### Step 1: Find a Step with Details
```
Bottom Detail Pane:
┌────────────────────────────────────────────────────┐
│ Detailed Steps (5)                                 │
├────────────────────────────────────────────────────┤
│ 1 │ ✓ Clone repository              8s  [✓] [→]  │ ← Arrow button
│ 2 │ ✓ Checkout branch               3s  [✓] [→]  │
│ 3 │ ⚙ Build Docker image           --  [⚙] [→]  │
│ 4 │ ○ Push to registry              --  [○]      │
└────────────────────────────────────────────────────┘
                                              ↑
                                    Click this arrow
```

### Step 2: Panel Opens with Full Details
```
Main Canvas                          Side Panel (slides in →)
┌─────────────────────────────┐     ┌────────────────────────┐
│                             │     │ Step Details      [×]  │
│   Workflow Pipeline         │     ├────────────────────────┤
│   [Stage 1] → [Stage 2] → ...│     │                        │
│                             │     │ ✓ Build Docker image   │
│                             │     │ Status: Running        │
│                             │     │ Time: 45s              │
│                             │     │                        │
│                             │     │ Substeps (3)           │
│   [Stage details below...]  │     │ ├─ ✓ Create context   │
│                             │     │ ├─ ⚙ Execute Docker..│
│                             │     │ └─ ○ Tag image        │
│                             │     │                        │
│                             │     │ Logs (4 lines)         │
│                             │     │ ┌──────────────────┐   │
│                             │     │ │ INFO: Starting...│   │
│                             │     │ │ INFO: Base img...│   │
│                             │     │ │ INFO: Installing.│   │
│                             │     │ │ INFO: Building...│   │
│                             │     │ └──────────────────┘   │
└─────────────────────────────┘     └────────────────────────┘
  ← Dimmed with backdrop              ← Full details visible
```

## Feature 2: Collapsible Parallel Stages

### Collapsed View (Default for >4 installs)
```
Canvas - Update Installs Stage:
┌────────────────────────────────┐
│  Update Installs               │
│  ╭──────────────────────────╮  │
│  │ ✓  Update Installs       │  │  ← Install 1 (visible)
│  │    install-5             │  │
│  ╰──────────────────────────╯  │
│  ╭──────────────────────────╮  │
│  │ ✗  Update Installs       │  │  ← Install 2 (visible)
│  │    install-6             │  │
│  ╰──────────────────────────╯  │
│  ╭──────────────────────────╮  │
│  │ ✓  Update Installs       │  │  ← Install 3 (visible)
│  │    install-7             │  │
│  ╰──────────────────────────╯  │
│  ╭──────────────────────────╮  │
│  │ ⚙  Update Installs       │  │  ← Install 4 (visible)
│  │    install-8             │  │
│  ╰──────────────────────────╯  │
│  ╭──────────────────────────╮  │
│  │    ⌄  +2 more            │  │  ← Expand button
│  ╰──────────────────────────╯  │
└────────────────────────────────┘
              ↑
         Click to expand
```

### Expanded View (Shows all installs)
```
Canvas - Update Installs Stage:
┌────────────────────────────────┐
│  Update Installs               │
│  ╭──────────────────────────╮  │
│  │ ✓  Update Installs       │  │  ← Install 1
│  │    install-5             │  │
│  ╰──────────────────────────╯  │
│  ╭──────────────────────────╮  │
│  │ ✗  Update Installs       │  │  ← Install 2
│  │    install-6             │  │
│  ╰──────────────────────────╯  │
│  ╭──────────────────────────╮  │
│  │ ✓  Update Installs       │  │  ← Install 3
│  │    install-7             │  │
│  ╰──────────────────────────╯  │
│  ╭──────────────────────────╮  │
│  │ ⚙  Update Installs       │  │  ← Install 4
│  │    install-8             │  │
│  ╰──────────────────────────╯  │
│  ╭──────────────────────────╮  │
│  │ ○  Update Installs       │  │  ← Install 5 (now visible)
│  │    install-9             │  │
│  ╰──────────────────────────╯  │
│  ╭──────────────────────────╮  │
│  │ ○  Update Installs       │  │  ← Install 6 (now visible)
│  │    install-10            │  │
│  ╰──────────────────────────╯  │
│  ╭──────────────────────────╮  │
│  │    ⌃  Collapse           │  │  ← Collapse button
│  ╰──────────────────────────╯  │
└────────────────────────────────┘
              ↑
      Click to collapse
```

## Status Icons Legend
```
✓  Completed     (Green)
⚙  Running       (Blue, animated spin)
✗  Failed        (Red)
○  Pending       (Grey)
→  View Details  (Arrow button)
```

## Interaction Flow

### Side Panel Flow:
1. User clicks stage → Details appear in bottom pane
2. User sees steps with arrow buttons [→]
3. User clicks arrow → Side panel slides in from right
4. User views logs, substeps, errors
5. User closes via X, ESC, or backdrop click
6. Panel slides out smoothly

### Expand/Collapse Flow:
1. Stage has 6 parallel installs
2. Canvas shows only 4 by default
3. "+2 more" button appears at bottom
4. User clicks → All 6 installs visible
5. "Collapse" button appears
6. User clicks → Back to showing 4

## Responsive Behavior

### Side Panel:
- Desktop: Half-width panel (can expand to full)
- Tablet: 3/4 width panel
- Mobile: Full-width panel

### Parallel Stages:
- All breakpoints: Same collapse behavior
- Threshold remains 4 regardless of screen size
- Horizontal scroll on canvas for overflow

## Accessibility Features

### Side Panel:
- ✓ Keyboard navigable (ESC to close)
- ✓ Focus trap within panel
- ✓ ARIA labels on buttons
- ✓ Screen reader announcements

### Expand/Collapse:
- ✓ Clear button text ("+2 more" vs "Collapse")
- ✓ Keyboard accessible
- ✓ Focus management on expand/collapse
- ✓ Visual indicators (chevron up/down)
