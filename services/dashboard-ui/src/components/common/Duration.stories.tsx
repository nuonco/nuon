/* eslint-disable react/no-unescaped-entities */
import { Duration } from './Duration'
import { Text } from './Text'
import { Icon } from './Icon'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Duration Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Duration components display time spans in human-readable formats using
        Luxon library. They accept either nanoseconds or ISO date strings for
        begin/end times. The component automatically handles formatting, zero
        durations, and invalid time ranges with appropriate fallbacks.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">From Nanoseconds</h4>
      <div className="space-y-3 p-4 border rounded">
        <div className="flex items-center gap-3">
          <Duration nanoseconds={1000000000} />
          <Text variant="subtext" theme="neutral">1 second (1,000,000,000 nanoseconds)</Text>
        </div>
        <div className="flex items-center gap-3">
          <Duration nanoseconds={0} />
          <Text variant="subtext" theme="neutral">Zero duration shows dash icon</Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">From ISO Timestamps</h4>
      <div className="space-y-3 p-4 border rounded">
        <div className="flex items-center gap-3">
          <Duration beginTime="2022-01-01T00:00:00Z" endTime="2022-01-01T00:00:01Z" />
          <Text variant="subtext" theme="neutral">1 second between timestamps</Text>
        </div>
        <div className="flex items-center gap-3">
          <Duration beginTime="2022-01-01T00:00:00Z" endTime="2022-01-01T01:30:45Z" />
          <Text variant="subtext" theme="neutral">1 hour, 30 minutes, 45 seconds</Text>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Automatic conversion from nanoseconds to milliseconds</li>
        <li>Human-readable formatting with appropriate units</li>
        <li>Zero duration handling with dash icon fallback</li>
        <li>Invalid duration detection with icon fallback</li>
      </ul>
    </div>
  </div>
)

export const Formats = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Duration Formats</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          format
        </code>{' '}
        prop controls how durations are displayed. Choose between human-readable
        text format and precise timer format depending on the context and
        precision requirements.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Format Comparison</h4>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="p-4 border rounded">
          <Text weight="strong" className="mb-3">Default Format</Text>
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Duration nanoseconds={1000000000} format="default" />
              <Text variant="subtext">1 second</Text>
            </div>
            <div className="flex items-center justify-between">
              <Duration nanoseconds={65000000000} format="default" />
              <Text variant="subtext">65 seconds</Text>
            </div>
            <div className="flex items-center justify-between">
              <Duration nanoseconds={3661000000000} format="default" />
              <Text variant="subtext">1 hour, 1 minute, 1 second</Text>
            </div>
          </div>
        </div>
        
        <div className="p-4 border rounded">
          <Text weight="strong" className="mb-3">Timer Format</Text>
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Duration nanoseconds={1000000000} format="timer" />
              <Text variant="subtext">Timer format</Text>
            </div>
            <div className="flex items-center justify-between">
              <Duration nanoseconds={65000000000} format="timer" />
              <Text variant="subtext">Timer format</Text>
            </div>
            <div className="flex items-center justify-between">
              <Duration nanoseconds={3661000000000} format="timer" />
              <Text variant="subtext">Timer format</Text>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>default:</strong> Human-readable format like "1h 30m 45s"
        for general use and user interfaces
      </div>
      <div>
        <strong>timer:</strong> Precise timer format like "T-01:30:45:00"
        for technical displays and monitoring
      </div>
    </div>
  </div>
)

export const TimeRanges = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Various Time Ranges</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Duration components handle a wide range of time spans from milliseconds
        to hours and beyond. Sub-second durations show higher precision, while
        longer durations are rounded appropriately for readability.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Short Durations</h4>
      <div className="space-y-3 p-4 border rounded">
        <div className="flex items-center justify-between">
          <Duration nanoseconds={500000} />
          <Text variant="subtext" theme="neutral">0.5 milliseconds</Text>
        </div>
        <div className="flex items-center justify-between">
          <Duration nanoseconds={1000000} />
          <Text variant="subtext" theme="neutral">1 millisecond</Text>
        </div>
        <div className="flex items-center justify-between">
          <Duration nanoseconds={100000000} />
          <Text variant="subtext" theme="neutral">100 milliseconds</Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Medium Durations</h4>
      <div className="space-y-3 p-4 border rounded">
        <div className="flex items-center justify-between">
          <Duration nanoseconds={30000000000} />
          <Text variant="subtext" theme="neutral">30 seconds</Text>
        </div>
        <div className="flex items-center justify-between">
          <Duration nanoseconds={300000000000} />
          <Text variant="subtext" theme="neutral">5 minutes</Text>
        </div>
        <div className="flex items-center justify-between">
          <Duration nanoseconds={1800000000000} />
          <Text variant="subtext" theme="neutral">30 minutes</Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Long Durations</h4>
      <div className="space-y-3 p-4 border rounded">
        <div className="flex items-center justify-between">
          <Duration nanoseconds={7200000000000} />
          <Text variant="subtext" theme="neutral">2 hours</Text>
        </div>
        <div className="flex items-center justify-between">
          <Duration nanoseconds={86400000000000} />
          <Text variant="subtext" theme="neutral">24 hours (1 day)</Text>
        </div>
        <div className="flex items-center justify-between">
          <Duration nanoseconds={604800000000000} />
          <Text variant="subtext" theme="neutral">7 days (1 week)</Text>
        </div>
      </div>
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Duration components are commonly used in deployment logs, performance
        metrics, and time tracking interfaces. Here are typical patterns for
        different application contexts.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Deployment Timeline</h4>
      <div className="space-y-3 border rounded-lg p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Icon variant="CheckCircle" size="16" className="text-green-600" />
            <div>
              <Text weight="strong">Build Completed</Text>
              <Text variant="subtext" theme="neutral">Docker image created successfully</Text>
            </div>
          </div>
          <Duration nanoseconds={125000000000} />
        </div>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Icon variant="CheckCircle" size="16" className="text-green-600" />
            <div>
              <Text weight="strong">Tests Passed</Text>
              <Text variant="subtext" theme="neutral">All unit and integration tests</Text>
            </div>
          </div>
          <Duration nanoseconds={45000000000} />
        </div>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Icon variant="CheckCircle" size="16" className="text-green-600" />
            <div>
              <Text weight="strong">Deployment</Text>
              <Text variant="subtext" theme="neutral">Rolling update to production</Text>
            </div>
          </div>
          <Duration nanoseconds={180000000000} />
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Performance Metrics</h4>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="p-4 border rounded">
          <Text weight="strong">API Response Times</Text>
          <div className="mt-3 space-y-2">
            <div className="flex justify-between">
              <Text variant="subtext">Average:</Text>
              <Duration nanoseconds={250000000} />
            </div>
            <div className="flex justify-between">
              <Text variant="subtext">95th percentile:</Text>
              <Duration nanoseconds={500000000} />
            </div>
            <div className="flex justify-between">
              <Text variant="subtext">99th percentile:</Text>
              <Duration nanoseconds={1200000000} />
            </div>
          </div>
        </div>
        
        <div className="p-4 border rounded">
          <Text weight="strong">Database Queries</Text>
          <div className="mt-3 space-y-2">
            <div className="flex justify-between">
              <Text variant="subtext">Select queries:</Text>
              <Duration nanoseconds={50000000} />
            </div>
            <div className="flex justify-between">
              <Text variant="subtext">Insert queries:</Text>
              <Duration nanoseconds={150000000} />
            </div>
            <div className="flex justify-between">
              <Text variant="subtext">Complex joins:</Text>
              <Duration nanoseconds={800000000} />
            </div>
          </div>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Process Monitoring</h4>
      <div className="p-4 border rounded-lg bg-gray-50 dark:bg-gray-800">
        <div className="flex items-center justify-between mb-4">
          <Text weight="stronger">Background Jobs</Text>
          <Text variant="subtext" theme="neutral">Duration tracking</Text>
        </div>
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <div>
              <Text variant="subtext" weight="strong">Data Backup</Text>
              <Text variant="label" theme="neutral">Completed 2 minutes ago</Text>
            </div>
            <Duration nanoseconds={1800000000000} format="timer" />
          </div>
          <div className="flex items-center justify-between">
            <div>
              <Text variant="subtext" weight="strong">Log Rotation</Text>
              <Text variant="label" theme="neutral">In progress</Text>
            </div>
            <Duration nanoseconds={45000000000} />
          </div>
          <div className="flex items-center justify-between">
            <div>
              <Text variant="subtext" weight="strong">Cache Warming</Text>
              <Text variant="label" theme="neutral">Scheduled</Text>
            </div>
            <Duration nanoseconds={0} />
          </div>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use default format for user-facing duration displays</li>
        <li>Use timer format for technical monitoring and logging</li>
        <li>Handle zero and invalid durations gracefully with icon fallbacks</li>
        <li>Consider the precision needed - sub-second vs. rounded durations</li>
        <li>Combine with appropriate text styling for different contexts</li>
      </ul>
    </div>
  </div>
)
