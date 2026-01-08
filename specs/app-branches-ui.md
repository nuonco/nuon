# App Branch Run Workflow - UI Spec

This spec defines the UI layout for the "Deploying Branch updates" page, which shows the progress of an app branch run workflow.

## Page Overview

The page displays a horizontal 4-stage pipeline showing the progression of an app branch deployment:
1. Fetch repository
2. Build config
3. Build changed components
4. Update installs

## ASCII Layout Reference

```
+--------------------------------------------------------------------------------------------------+
|  HEADER BAR                                                                                      |
|  nuon logo  |  Applications > App_name > Installation (breadcrumb)  |  Search  |  User Profile  |
+--------------------------------------------------------------------------------------------------+
|  SIDEBAR    |  MAIN CONTENT                                                                      |
|             |                                                                                    |
|  Org Select |  PAGE HEADER                                                                       |
|  ┌────────┐ |  ┌────────────────────────────────────────────────────────────────────────────┐   |
|  │Acme Inc│ |  │ Deploying Branch updates                                                   │   |
|  │● Active│ |  │ Watch your install provision here and provide needed approvals.           │   |
|  └────────┘ |  │                                                                            │   |
|             |  │                         ◉ Auto Approve   🔒 Lock   [+ New Function]        │   |
|  Nav Items  |  └────────────────────────────────────────────────────────────────────────────┘   |
|  Dashboard  |                                                                                    |
|  ■ Apps     |  METRICS BAR                                                                       |
|  Installs   |  ┌──────────────┬──────────────────┬─────────────────┬─────────────────────────┐   |
|  Build run  |  │ Total time   │ Pending approval │ Completed steps │ Total steps             │   |
|  Eagle eye  |  │ 2m 21s       │ 7                │ 0               │ 15                      │   |
|             |  └──────────────┴──────────────────┴─────────────────┴─────────────────────────┘   |
|  Settings   |                                                                                    |
|  ├─ Team    |  HORIZONTAL PIPELINE                                                               |
|  └─ Account |  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐  |
|             |  │ ✓ Fetch     │───>│ ◐ Build     │───>│ ✓ Build     │───>│   Update            │  |
|  Resources  |  │   repository│    │   config    │    │   changed   │    │   installs          │  |
|  ├─ Docs    |  │             │    │             │    │   components│    │                     │  |
|  └─ Release |  │ pr:updated- │    │ Building    │    │             │    │ 0/4 completed       │  |
|             |  │ components  │    │ components  │    │ Completed   │    │                     │  |
|             |  │             │    │             │    │             │    │ ┌─────────────────┐ │  |
|             |  │             │    │ Status:     │    │ Changes     │    │ │ Install-01      │ │  |
|             |  │             │    │ 4/10 builds │    │ detected:   │    │ │ View workflow   │ │  |
|             |  │             │    │ completed   │    │ 4 components│    │ ├─────────────────┤ │  |
|             |  │             │    │             │    │             │    │ │ Install-02      │ │  |
|             |  │             │    │             │    │             │    │ │ View workflow   │ │  |
|             |  │             │    │             │    │             │    │ ├─────────────────┤ │  |
|             |  │             │    │             │    │             │    │ │ Install-03      │ │  |
|             |  │             │    │             │    │             │    │ │ View workflow   │ │  |
|             |  │             │    │             │    │             │    │ ├─────────────────┤ │  |
|             |  │             │    │             │    │             │    │ │ Install-04      │ │  |
|             |  │             │    │             │    │             │    │ │ View workflow   │ │  |
|             |  └─────────────┘    └─────────────┘    └─────────────┘    │ └─────────────────┘ │  |
|             |                                                           └─────────────────────┘  |
+--------------------------------------------------------------------------------------------------+
```

## Section Breakdown

### 1. Page Header

**Purpose:** Title, description, and primary actions for the branch run.

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│ Deploying Branch updates                          ◉ Auto Approve  🔒 Lock  [+]   │
│ Watch your install provision here and provide needed approvals.                  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

**Elements:**
- Title: "Deploying Branch updates"
- Subtitle/description text
- Auto Approve toggle (Switch component)
- Lock button (IconButton)
- "+ New Function" button (Primary Button)

### 2. Metrics Bar

**Purpose:** Summary statistics for the current run.

```
┌──────────────┬──────────────────┬─────────────────┬─────────────────────────┐
│ Total time   │ Pending approval │ Completed steps │ Total steps             │
│ 2m 21s       │ 7                │ 0               │ 15                      │
└──────────────┴──────────────────┴─────────────────┴─────────────────────────┘
```

**Metrics:**
| Metric | Description | Format |
|--------|-------------|--------|
| Total time | Elapsed time since run started | Duration (Xm Xs) |
| Pending approval | Steps waiting for user approval | Integer |
| Completed steps | Successfully finished steps | Integer |
| Total steps | Total steps in workflow | Integer |

### 3. Horizontal Pipeline

**Purpose:** Visual representation of the 4-stage deployment workflow.

```
   STAGE 1           STAGE 2           STAGE 3              STAGE 4
┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌─────────────────────┐
│ ✓ Fetch     │──>│ ◐ Build     │──>│ ✓ Build     │──>│   Update            │
│   repository│   │   config    │   │   changed   │   │   installs          │
│             │   │             │   │   components│   │                     │
│ [details]   │   │ [progress]  │   │ [summary]   │   │ [install list]      │
└─────────────┘   └─────────────┘   └─────────────┘   └─────────────────────┘
```

#### Stage 1: Fetch Repository
- **Status:** Complete/In Progress/Pending/Failed
- **Content:** Branch/PR reference (e.g., "pr:updated-components")

#### Stage 2: Build Config
- **Status:** Complete/In Progress/Pending/Failed
- **Content:**
  - Status text (e.g., "Building components...")
  - Progress indicator (e.g., "4/10 builds completed")

#### Stage 3: Build Changed Components
- **Status:** Complete/In Progress/Pending/Failed
- **Content:**
  - Completion status
  - Changes detected count (e.g., "4 components")

#### Stage 4: Update Installs
- **Status:** Complete/In Progress/Pending/Failed
- **Content:**
  - Progress (e.g., "0/4 completed")
  - **Expandable install list:**
    ```
    ┌─────────────────┐
    │ Install-01      │
    │ [View workflow] │
    ├─────────────────┤
    │ Install-02      │
    │ [View workflow] │
    ├─────────────────┤
    │ ...             │
    └─────────────────┘
    ```

## Pipeline Stage States

Each stage can be in one of four states:

| State | Icon | Visual Treatment |
|-------|------|------------------|
| Pending | ○ | Gray/muted, dashed border |
| In Progress | ◐ | Purple/primary, animated |
| Complete | ✓ | Green, solid border |
| Failed | ✗ | Red, error styling |

## Connector Arrows

Stages are connected with directional arrows showing flow:
```
[Stage 1] ───> [Stage 2] ───> [Stage 3] ───> [Stage 4]
```

Arrow styling:
- Solid line for completed transitions
- Dashed line for pending transitions
- Animated for in-progress transitions

## Responsive Behavior

### Desktop (>1024px)
- Full horizontal pipeline with all stages visible
- Install list expanded within Stage 4 card

### Tablet (768px - 1024px)
- Horizontal pipeline, stages may wrap to 2 rows (2x2 grid)
- Install list collapsed by default

### Mobile (<768px)
- Vertical stack of stages
- Stages become expandable accordion items
- Install list in modal/drawer

## Interaction Patterns

### Stage Card Click
- Expands to show detailed step information
- Shows logs/output for that stage

### Install Link Click
- Navigates to install workflow detail page
- Route: `/[org-id]/installs/[install-id]/workflows/[workflow-id]`

### Auto Approve Toggle
- Enables/disables automatic approval for pending steps
- Shows confirmation dialog when enabling

### Lock Button
- Prevents any modifications to the running workflow
- Shows locked state indicator

## Data Requirements

```typescript
interface AppBranchRun {
  id: string;
  appId: string;
  branchId: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  startedAt: Date;
  completedAt?: Date;

  // Metrics
  totalSteps: number;
  completedSteps: number;
  pendingApprovals: number;

  // Stages
  stages: {
    fetchRepository: StageStatus;
    buildConfig: StageStatus;
    buildComponents: StageStatus;
    updateInstalls: StageStatus;
  };

  // Install updates
  installUpdates: InstallUpdate[];
}

interface StageStatus {
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
  startedAt?: Date;
  completedAt?: Date;
  details?: Record<string, any>;
}

interface InstallUpdate {
  installId: string;
  installName: string;
  workflowId?: string;
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
}
```

## Route Structure

```
/[org-id]/apps/[app-id]/branches/[branch-id]/runs/[run-id]
```

## Component Mapping

### Existing Components (Ready to Use)

| UI Section | Component | Location | Props/Usage |
|------------|-----------|----------|-------------|
| **Page Header Title** | `HeadingGroup` | `/src/components/common/HeadingGroup.tsx` | Wraps title + description |
| **Page Header Title** | `Text` | `/src/components/common/Text.tsx` | `variant="h2"`, `weight="strong"` |
| **Page Header Desc** | `Text` | `/src/components/common/Text.tsx` | `variant="body"`, `theme="neutral"` |
| **Lock Button** | `Button` | `/src/components/common/Button.tsx` | `variant="ghost"` with icon |
| **New Function Button** | `Button` | `/src/components/common/Button.tsx` | `variant="primary"` |
| **Metrics Container** | `LabeledValue` | `/src/components/common/LabeledValue.tsx` | Label + value pairs |
| **Duration Display** | `Duration` | `/src/components/common/Duration.tsx` | `nanoseconds={totalTime}` |
| **Stage Status Icons** | `Status` | `/src/components/common/Status.tsx` | `variant="timeline"` for ✓◐○✗ |
| **Stage Cards** | `Card` | `/src/components/common/Card.tsx` | Base card structure |
| **Stage Badges** | `Badge` | `/src/components/common/Badge.tsx` | Status indicators |
| **Install List Divider** | `Divider` | `/src/components/common/Divider.tsx` | Between list items |

### Reference Implementations

| Feature | Reference File | Key Patterns |
|---------|---------------|--------------|
| **Metrics Display** | `/src/components/workflows/workflow-details/WorkflowMetrics.tsx` | Perfect pattern for metrics bar |
| **Horizontal Graph** | `/src/components/actions/ActionStepsGraph.tsx` | ReactFlow + dagre layout |
| **Status Indicators** | `/src/components/common/Status.tsx` | Timeline variant icons |
| **Status Utilities** | `/src/utils/status-utils.ts` | Maps statuses to themes |

### New Components Required

| Component | Priority | Purpose | Implementation Notes |
|-----------|----------|---------|---------------------|
| **Switch/Toggle** | HIGH | Auto Approve toggle | Standard checkbox-based toggle pattern |
| **HorizontalPipeline** | MEDIUM | 4-stage layout container | CSS Grid/Flexbox (simpler than ReactFlow) |
| **PipelineStageCard** | MEDIUM | Specialized stage cards | Extends `Card` component |
| **PipelineConnector** | LOW | Arrow SVG between stages | Simple SVG component |

### Status Mapping

```tsx
// Use Status component with variant="timeline"
<Status status="success" variant="timeline" />  // ✓ Complete
<Status status="running" variant="timeline" />  // ◐ In Progress
<Status status="pending" variant="timeline" />  // ○ Pending
<Status status="error" variant="timeline" />    // ✗ Failed
```

### Example Metrics Implementation

```tsx
// Based on WorkflowMetrics.tsx pattern
<div className="flex gap-6">
  <LabeledValue label="Total time">
    <Duration nanoseconds={run.totalTime} variant="base" />
  </LabeledValue>
  <LabeledValue label="Pending approval">
    <Text variant="base">{run.pendingApprovals}</Text>
  </LabeledValue>
  <LabeledValue label="Completed steps">
    <Text variant="base">{run.completedSteps}</Text>
  </LabeledValue>
  <LabeledValue label="Total steps">
    <Text variant="base">{run.totalSteps}</Text>
  </LabeledValue>
</div>
```

## Implementation Strategy

### Phase 1: Core Layout
1. Create page structure using existing layout patterns
2. Implement metrics bar using `WorkflowMetrics.tsx` pattern
3. Use existing `Status`, `Card`, `Text` components

### Phase 2: New Components
1. Build `Switch/Toggle` component for Auto Approve
2. Create `PipelineStageCard` extending `Card`
3. Add `PipelineConnector` SVG component

### Phase 3: Pipeline Assembly
1. Build horizontal layout container with CSS Grid
2. Integrate stage cards with connectors
3. Add responsive behavior (2x2 grid on tablet, vertical on mobile)

## Related Specs

- [App Branches](./app-branches.md) - Core feature specification
