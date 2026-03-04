# Branch Workflow Canvas - New Features Summary

## Overview
Added two major features to the branch workflow canvas to improve user experience and information access:
1. **Side Panel for Step Details** - Click on steps to view detailed logs, substeps, and metadata
2. **Collapsible Parallel Install Stages** - Automatically collapse stages with >4 parallel installs for cleaner UI

## Feature 1: Side Panel for Step Details

### What It Does
- Clicking the arrow button (→) on any step in the bottom detail pane opens a slide-in panel from the right
- Shows comprehensive step information including:
  - Step status with visual indicators
  - Execution time
  - Error messages (if failed)
  - Substeps with nested hierarchy
  - Complete logs with syntax highlighting
  - Step ID for reference

### User Experience
- **Panel Behavior:**
  - Slides in from the right side
  - Half-width by default (can be expanded to full screen)
  - Backdrop overlay dims the main content
  - Close via X button, escape key, or clicking outside
  - Smooth animations for open/close

- **Step Access:**
  - Arrow button appears on steps that have details (substeps, logs, or errors)
  - Click the arrow to open the panel
  - Panel shows rich formatting with color-coded status indicators

### Implementation Details
```typescript
// New component: StepDetailSidePanel
// State management in BranchWorkflowCanvas:
const [selectedStep, setSelectedStep] = useState<IWorkflowStep | null>(null)
const [isPanelOpen, setIsPanelOpen] = useState(false)

// Integrated with existing Panel component from dashboard-ui
<Panel
  isVisible={isPanelOpen}
  onClose={handleClosePanel}
  heading={...}
  size="half"
>
  {/* Step details content */}
</Panel>
```

### Components Modified
- `branch-workflow-canvas.tsx`: Added StepDetailSidePanel component
- `CollapsibleStepDetailRow`: Added arrow button and onOpenPanel callback
- `StageDetailSection`: Passes onOpenStepPanel handler to all step rows
- `BranchWorkflowCanvas`: Added state management and handlers

## Feature 2: Collapsible Parallel Install Stages

### What It Does
- Parallel install stages with **MORE THAN 4 installs** are automatically collapsed
- Shows only the first 4 installs by default
- Displays an expand button showing "+X more" (e.g., "+2 more")
- Clicking expands to show all installs
- Clicking "Collapse" returns to showing 4

### Visual Example
**Collapsed (>4 installs):**
```
┌─────────────────────────┐
│ Update Installs         │
│ install-5              │
├─────────────────────────┤
│ Update Installs         │
│ install-6              │
├─────────────────────────┤
│ Update Installs         │
│ install-7              │
├─────────────────────────┤
│ Update Installs         │
│ install-8              │
├─────────────────────────┤
│     ⌄ +2 more          │ ← Expand button
└─────────────────────────┘
```

**Expanded:**
```
┌─────────────────────────┐
│ Update Installs         │
│ install-5              │
├─────────────────────────┤
│ Update Installs         │
│ install-6              │
├─────────────────────────┤
... (all 6 installs shown)
├─────────────────────────┤
│     ⌃ Collapse          │ ← Collapse button
└─────────────────────────┘
```

### User Experience
- **Automatic Threshold:** Only activates when a stage has >4 parallel installs
- **Smooth Animation:** Expands/collapses with CSS transitions
- **Clear Indication:** Button shows exactly how many more installs are hidden
- **Maintains Context:** Selected stage indicator remains visible

### Implementation Details
```typescript
// In WorkflowStageCard component:
const COLLAPSE_THRESHOLD = 4
const shouldShowExpandButton = hasParallelInstalls && 
  stage.parallelInstalls!.length > COLLAPSE_THRESHOLD

const visibleInstalls = isExpanded 
  ? stage.parallelInstalls! 
  : stage.parallelInstalls!.slice(0, COLLAPSE_THRESHOLD)

// State management:
const [expandedParallelStages, setExpandedParallelStages] = 
  useState<Set<string>>(new Set())
```

### Components Modified
- `WorkflowStageCard`: Added expand/collapse logic and button
- `BranchWorkflowCanvas`: Added expandedParallelStages state management
- Mock data: Updated stage-5 to have 6 installs for testing

## Updated User Instructions

The canvas page description was updated to:
> "Drag the canvas to navigate. Click any stage to view details below. Click step arrows to view logs and details."

## Testing the Features

### To Test Side Panel:
1. Navigate to the canvas page
2. Click on a stage to view details in the bottom pane
3. Look for steps with arrow buttons (→) on the right
4. Click the arrow to open the side panel
5. Verify:
   - Panel slides in smoothly from right
   - Step details are fully visible
   - Logs display with proper formatting
   - Substeps show hierarchy
   - Close button, ESC key, and backdrop click all work

### To Test Expand/Collapse:
1. Navigate to the canvas page
2. Find a stage with >4 parallel installs (e.g., the last "Update Installs" stage with 6 installs)
3. Verify only 4 installs are shown initially
4. Verify "+2 more" button appears at the bottom
5. Click the button to expand
6. Verify all 6 installs are now visible
7. Verify "Collapse" button appears
8. Click collapse to return to 4 visible

## Mock Data Updates

Added 4 additional parallel installs to the second "Update Installs" stage:
- Customer C Install (completed)
- Customer D Install (running)
- Customer E Install (pending)
- Customer F Install (pending)

Total: 6 parallel installs (demonstrating the collapse feature)

## Files Changed

### Modified Files:
- `/services/dashboard-ui/src/app/[org-id]/apps/[app-id]/branches/[branch-id]/canvas/branch-workflow-canvas.tsx`
  - Added StepDetailSidePanel component
  - Updated WorkflowStageCard with expand/collapse
  - Added state management for panel and expansion
  - Updated CollapsibleStepDetailRow with arrow button
  - Updated StageDetailSection to pass panel handler
  - Added 4 more parallel install mocks

- `/services/dashboard-ui/src/app/[org-id]/apps/[app-id]/branches/[branch-id]/canvas/page.tsx`
  - Updated page description text

### Stats:
- 1,565 lines added
- 143 lines removed
- Net: ~1,400 lines of new functionality

## Key Design Decisions

1. **Reused Existing Panel Component:** Leveraged the dashboard's existing Panel component for consistency
2. **Threshold of 4:** Chose 4 as the collapse threshold based on typical CI/CD pipeline displays
3. **Per-Stage Expansion:** Each stage tracks its own expanded state independently
4. **Arrow Button Placement:** Added to step rows only when details exist
5. **Status-Aware Styling:** Panel content uses status colors for visual consistency

## Future Enhancements

Potential improvements for production:
- Add keyboard shortcuts for panel navigation (←/→ for prev/next step)
- Add "View in Temporal" link for steps with workflow IDs
- Add download logs button
- Add search/filter within logs
- Add ability to share deep links to specific steps
- Persist expansion state in URL params or local storage
- Add animation when expanding/collapsing parallel stages
