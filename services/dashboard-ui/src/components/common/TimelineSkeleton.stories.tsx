/* eslint-disable react/no-unescaped-entities */
import { TimelineSkeleton } from './TimelineSkeleton'
import { Text } from './Text'
import { Button } from './Button'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic TimelineSkeleton Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        TimelineSkeleton provides loading placeholders for timeline components.
        It creates realistic skeleton events with animated shimmer effects
        that match the structure and spacing of actual TimelineEvent components.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Default Loading State (5 events)</h4>
      <div className="border rounded-lg p-4">
        <TimelineSkeleton />
      </div>
      <Text variant="subtext" theme="neutral">
        Default skeleton shows 5 placeholder events with animated loading effects
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Matches TimelineEvent structure and visual layout</li>
        <li>Animated shimmer effects for realistic loading appearance</li>
        <li>Configurable number of skeleton events</li>
        <li>Consistent spacing and timeline connector lines</li>
        <li>Dark mode support with appropriate contrast</li>
      </ul>
    </div>
  </div>
)

export const CustomEventCounts = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Custom Event Counts</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The number of skeleton events can be customized based on expected
        timeline length and loading context. Use fewer events for quick
        operations and more for extensive activity histories.
      </p>
    </div>

    <div className="space-y-6">
      <div className="space-y-3">
        <h4 className="text-sm font-medium">Short Timeline (3 events)</h4>
        <div className="border rounded-lg p-4">
          <TimelineSkeleton eventCount={3} />
        </div>
        <Text variant="subtext" theme="neutral">
          Compact loading state for shorter timelines
        </Text>
      </div>
      
      <div className="space-y-3">
        <h4 className="text-sm font-medium">Extended Timeline (8 events)</h4>
        <div className="border rounded-lg p-4">
          <TimelineSkeleton eventCount={8} />
        </div>
        <Text variant="subtext" theme="neutral">
          Extended loading state for comprehensive activity histories
        </Text>
      </div>
      
      <div className="space-y-3">
        <h4 className="text-sm font-medium">Single Event (1 event)</h4>
        <div className="border rounded-lg p-4">
          <TimelineSkeleton eventCount={1} />
        </div>
        <Text variant="subtext" theme="neutral">
          Minimal loading state for single event scenarios
        </Text>
      </div>
    </div>
  </div>
)

export const CustomStyling = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Custom Styling</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        TimelineSkeleton supports custom styling through className props.
        This allows for consistent integration with your application's
        design system and layout requirements.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Styled Container</h4>
      <TimelineSkeleton 
        className="border border-gray-300 dark:border-gray-600 rounded-lg p-6 shadow-sm" 
        eventCount={4}
      />
      <Text variant="subtext" theme="neutral">
        Custom borders, padding, and shadows applied to skeleton container
      </Text>
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        TimelineSkeleton is commonly used in activity feeds, deployment
        histories, and audit logs while actual timeline data is loading.
        It provides visual continuity and better perceived performance.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Activity Feed Loading</h4>
      <div className="p-4 border rounded-lg space-y-4">
        <div className="flex justify-between items-center">
          <Text variant="h3" weight="stronger">Recent Activity</Text>
          <Button variant="ghost" size="sm" disabled>Refresh</Button>
        </div>
        <TimelineSkeleton eventCount={4} />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Deployment History Loading</h4>
      <div className="p-4 border rounded-lg space-y-4">
        <div className="flex justify-between items-center">
          <Text variant="h3" weight="stronger">Deployment History</Text>
          <div className="flex gap-2">
            <Button variant="ghost" size="sm" disabled>Filter</Button>
            <Button variant="primary" size="sm" disabled>New Deployment</Button>
          </div>
        </div>
        <TimelineSkeleton eventCount={6} />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Implementation Example</h4>
      <div className="p-4 border rounded-lg bg-gray-50 dark:bg-gray-800">
        <Text variant="label" theme="neutral" className="mb-2">Conditional Rendering Pattern:</Text>
        <div className="font-mono text-sm space-y-1">
          <div>{'{'}</div>
          <div>{'  isLoading ? ('}</div>
          <div>{'    <TimelineSkeleton eventCount={expectedCount} />'}</div>
          <div>{'  ) : ('}</div>
          <div>{'    <Timeline events={actualEvents} ... />'}</div>
          <div>{'  )'}</div>
          <div>{'}'}</div>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Match skeleton event count to expected timeline length</li>
        <li>Use consistent skeleton patterns across your application</li>
        <li>Consider loading time duration when choosing event count</li>
        <li>Test skeleton appearance in both light and dark modes</li>
        <li>Ensure skeleton maintains proper spacing with surrounding content</li>
        <li>Disable interactive elements during skeleton loading state</li>
      </ul>
    </div>
  </div>
)
