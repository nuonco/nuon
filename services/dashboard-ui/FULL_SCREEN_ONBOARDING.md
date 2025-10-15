# Full-Screen Onboarding Implementation

This document describes the implementation of the full-screen onboarding experience as an alternative to the existing modal-based onboarding flow.

## Overview

The full-screen onboarding provides a focused, distraction-free experience for new users by displaying onboarding content in a clean full-viewport layout instead of a modal overlay.

## Implementation Details

### Core Components

#### 1. `FullScreenOnboarding.tsx`
- **Location**: `/src/components/Apps/FullScreenOnboarding.tsx`
- **Purpose**: Main full-screen onboarding component
- **Key Features**:
  - Full viewport overlay with clean background
  - Top navigation bar with Skip button and manual navigation arrows
  - Progress indicator showing step completion
  - Reuses all existing step content components
  - Maintains all existing journey progression logic

#### 2. `OnboardingNavigation` (Sub-component)
- **Embedded in**: `FullScreenOnboarding.tsx`
- **Purpose**: Top navigation bar for the full-screen experience
- **Features**:
  - Left/Right arrow navigation
  - Step progress display (hidden on mobile if space is limited)
  - Skip button and Close button
  - Responsive design with hover effects and transitions

#### 3. `StepProgressIndicator` (Sub-component)
- **Embedded in**: `FullScreenOnboarding.tsx`
- **Purpose**: Visual progress indicator
- **Features**:
  - Circular step indicators with completion checkmarks
  - Connecting lines between steps
  - Responsive sizing for mobile and desktop
  - Smooth animations for step transitions

### Feature Flag System

#### Environment Variable
- **Name**: `NEXT_PUBLIC_ENABLE_FULL_SCREEN_ONBOARDING`
- **Default**: `false` (modal experience)
- **Location**: Added to both `env` and `docker_env` sections in `service.yml`
- **Usage**: Set to `true` to enable full-screen onboarding

#### Configuration Files
- **Service Config**: `services/dashboard-ui/service.yml` - Contains default values
- **Local Override**: `.env.local.example` - Template for local development testing

### Integration Points

#### UserJourneyProvider Updates
- **File**: `/src/providers/user-journey-provider.tsx`
- **Changes**: Added conditional rendering based on feature flag
- **Behavior**:
  - Checks `process.env.NEXT_PUBLIC_ENABLE_FULL_SCREEN_ONBOARDING`
  - Renders `FullScreenOnboarding` when `true`, `EvaluationUserJourneyModal` when `false`
  - Preserves all existing modal logic as fallback
  - Uses same portal rendering mechanism

### Design Features

#### Layout Structure
```
┌─────────────────────────────────────────┐
│  [←] [→]     Step X of Y    [Skip] [×]  │  <- Top navigation
├─────────────────────────────────────────┤
│                                         │
│    Get started with Nuon               │  <- Header
│    Follow these steps to set up...     │
│                                         │
│          ● ── ● ── ○ ── ○ ── ○           │  <- Progress indicator
│                                         │
│         ┌─────────────────────┐         │
│         │   STEP CONTENT      │         │  <- Expandable step content
│         │   (Existing components)      │
│         └─────────────────────┘         │
│                                         │
└─────────────────────────────────────────┘
```

#### Responsive Design
- **Mobile**: Smaller progress indicators, reduced padding, hidden step counter on very small screens
- **Desktop**: Larger elements, more generous spacing, full navigation visible
- **Transitions**: Smooth animations on hover and step progression
- **Accessibility**: Proper disabled states and keyboard navigation support

### Key Benefits

#### User Experience
- ✅ **Focused Experience**: No distracting background content
- ✅ **More Space**: Generous room for educational content
- ✅ **Manual Navigation**: Users can navigate back/forth between steps
- ✅ **Visual Progress**: Clear indication of completion status

#### Technical
- ✅ **Safe Rollout**: Feature flag allows gradual deployment
- ✅ **Backward Compatible**: Original modal preserved as fallback
- ✅ **Code Reuse**: All existing step content components preserved
- ✅ **Consistent Logic**: Journey progression and completion logic unchanged

#### Development
- ✅ **Low Risk**: No modifications to existing functionality
- ✅ **Easy Testing**: Local environment variable for development
- ✅ **Maintainable**: Clean separation of concerns
- ✅ **Rollback Ready**: Simple flag flip to revert if needed

### Usage Instructions

#### For Developers

**Enable for Local Testing**:
1. Create `.env.local` file in dashboard-ui directory
2. Add: `NEXT_PUBLIC_ENABLE_FULL_SCREEN_ONBOARDING=true`
3. Restart development server

**Enable via nuonctl**:
1. Add to `~/.nuonctl-env.yml`:
   ```yaml
   dashboard-ui:
     NEXT_PUBLIC_ENABLE_FULL_SCREEN_ONBOARDING: true
   ```
2. Restart services with `./run-nuonctl.sh services dev --dev dashboard-ui`

#### For Production Deployment

**Gradual Rollout**:
1. Deploy code with feature flag `false` (default)
2. Test in staging environment with flag enabled
3. Enable in production by updating environment configuration
4. Monitor metrics and user feedback
5. Keep flag for easy rollback if needed

## Single-Step Display Enhancement

### Focused Step-by-Step Experience

The full-screen onboarding now displays one step at a time instead of showing all steps in accordions, providing a more focused and guided experience.

#### Single-Step Display Features
- **One Step at a Time**: Only the current step is visible, eliminating distractions
- **Rich Step Headers**: Each step shows a large status indicator, title, description, and progress counter
- **Manual Navigation**: Users can navigate between steps using arrow buttons in the top navigation
- **Smooth Transitions**: Steps fade and slide when transitioning between them
- **Navigation Context**: Visual indicators show when users are viewing previous or future steps

#### Navigation Behavior
- **Automatic Progression**: Steps advance automatically when completed (existing behavior)
- **Manual Browsing**: Users can navigate backward to review completed steps or forward through the journey
- **Current Step Reset**: "Go to current step" button returns users to their actual progress when manually browsing
- **Visual Feedback**: Navigation arrows are disabled at first/last steps, with clear visual states

#### Layout Structure
```
┌─────────────────────────────────────────┐
│  [←] [→]     Step X of Y    [Skip] [×]  │  <- Top navigation
├─────────────────────────────────────────┤
│          ● ── ● ── ○ ── ○ ── ○           │  <- Progress indicator
│                                         │
│              [Step Icon]                │  <- Large step status circle
│           Step 2: Install CLI           │  <- Step title
│         Install the Nuon CLI tool       │  <- Step description
│              Step 2 of 6                │  <- Progress text
│                                         │
│         ┌─────────────────────┐         │
│         │   STEP CONTENT      │         │  <- Full step content
│         │   (Existing components)      │
│         └─────────────────────┘         │
│                                         │
└─────────────────────────────────────────┘
```

### File Changes Summary

#### New Files
- `/src/components/Apps/FullScreenOnboarding.tsx` - Main component with single-step display
- `/src/components/Apps/OnboardingStepHeader.tsx` - Step header with title/description
- `/src/hooks/use-journey-polling-interval.ts` - Journey-based dynamic polling
- `/services/dashboard-ui/.env.local.example` - Environment template
- `/services/dashboard-ui/FULL_SCREEN_ONBOARDING.md` - This documentation

#### Modified Files
- `/services/dashboard-ui/service.yml` - Added feature flag environment variable
- `/src/providers/user-journey-provider.tsx` - Added conditional rendering logic
- `/src/providers/account-provider.tsx` - Added dynamic polling support

#### Preserved Files (No Changes)
- All existing step content components (`CreateAppStepContent.tsx`, etc.)
- All journey progression and completion logic
- `EvaluationUserJourneyModal.tsx` - Original modal preserved
- `ChecklistItem.tsx` - Original accordion component preserved

### Technical Considerations

#### Performance
- Full-screen overlay uses `fixed` positioning for optimal rendering
- Transitions and animations are GPU-accelerated with `transform` properties
- Progress indicator uses `flex-shrink-0` to prevent layout issues on mobile

#### Accessibility
- Proper ARIA labels and keyboard navigation support maintained
- Focus management preserved from original modal implementation
- High contrast indicators for step completion status

#### Browser Compatibility
- Uses modern CSS features (`backdrop-blur`, `transition-all`) with fallbacks
- Responsive design uses established Tailwind breakpoints
- No JavaScript features beyond what original modal used

## Dynamic Polling Enhancement

### Responsive Data Refresh

To improve the onboarding experience, the system now includes dynamic polling that adjusts refresh intervals based on user journey state:

#### Polling Behavior
- **During Onboarding**: 5-second polling interval for responsive step updates
- **After Completion**: 20-second polling interval for normal operation
- **No Account Data**: 5-second polling for faster initial load

#### Implementation Details

**Journey-Based Interval Logic**:
- Monitors evaluation journey completion status
- Automatically switches between fast/slow polling
- Reduces wait time from potential 20s to maximum 5s during onboarding

**Files Modified for Polling**:
- `src/hooks/use-journey-polling-interval.ts` - Journey-based polling logic
- `src/providers/account-provider.tsx` - Dynamic polling integration

**Debug Information**:
In development mode, console logs show current polling intervals:
- `📊 Account polling: 5s (onboarding active - incomplete: app_created, install_created)`
- `📊 Account polling: 20s (onboarding complete)`

#### Benefits
- ✅ **Faster Response**: Step completion detected within 5 seconds instead of up to 20 seconds
- ✅ **Automatic Optimization**: Polling slows down automatically after onboarding completion
- ✅ **Resource Efficient**: Only uses fast polling when needed
- ✅ **Development Friendly**: Clear logging shows polling behavior in dev mode

This enhancement ensures that users don't have to wait awkwardly for the dashboard to refresh during the critical onboarding flow, while maintaining efficient resource usage during normal operation.

This implementation provides a significant UX improvement while maintaining system stability and allowing for safe, controlled rollout.