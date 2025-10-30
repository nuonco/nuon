/* eslint-disable react/no-unescaped-entities */
import { Time } from './Time'
import { Text } from './Text'
import { Badge } from './Badge'
import { Button } from './Button'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Time Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The Time component provides consistent timestamp formatting with
        multiple display formats. It automatically uses the current time
        when no specific time is provided and renders as a semantic HTML
        time element with proper datetime attributes.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Current Time (Default)</h4>
      <div className="p-3 border rounded">
        <Time />
      </div>
      <Text variant="subtext" theme="neutral">
        Shows current timestamp using default format
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Multiple format options for different contexts</li>
        <li>Supports Unix timestamps and ISO date strings</li>
        <li>Semantic HTML time element with datetime attributes</li>
        <li>Consistent timezone handling and formatting</li>
        <li>Accessible timestamp display for screen readers</li>
      </ul>
    </div>
  </div>
)

export const TimeFormats = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Time Format Options</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Different format options serve different use cases - from compact
        displays to detailed timestamps. Each format is optimized for
        specific interface contexts and user needs.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Available Formats</h4>
      <div className="space-y-3">
        <div className="flex items-center justify-between p-3 border rounded">
          <Text variant="subtext" theme="neutral">Short DateTime:</Text>
          <Time format="short-datetime" />
        </div>
        <div className="flex items-center justify-between p-3 border rounded">
          <Text variant="subtext" theme="neutral">Long DateTime:</Text>
          <Time format="long-datetime" />
        </div>
        <div className="flex items-center justify-between p-3 border rounded">
          <Text variant="subtext" theme="neutral">Relative Time:</Text>
          <Time format="relative" />
        </div>
        <div className="flex items-center justify-between p-3 border rounded">
          <Text variant="subtext" theme="neutral">Time Only:</Text>
          <Time format="time-only" />
        </div>
        <div className="flex items-center justify-between p-3 border rounded">
          <Text variant="subtext" theme="neutral">Log DateTime:</Text>
          <Time format="log-datetime" />
        </div>
      </div>
    </div>

    <div className="space-y-3">
      <h4 className="text-sm font-medium">Format Use Cases</h4>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
        <div>
          <Text weight="strong">short-datetime:</Text> Compact format for tables and lists
        </div>
        <div>
          <Text weight="strong">long-datetime:</Text> Detailed format for full context
        </div>
        <div>
          <Text weight="strong">relative:</Text> Human-readable time differences
        </div>
        <div>
          <Text weight="strong">time-only:</Text> Just the time portion for schedules
        </div>
        <div>
          <Text weight="strong">log-datetime:</Text> Precise format for debugging and logs
        </div>
      </div>
    </div>
  </div>
)

export const UnixTimestamps = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Unix Timestamp Support</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The Time component accepts Unix timestamps (seconds since epoch)
        for displaying specific moments in time. This is especially useful
        for API data and historical timestamps.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Timestamp Examples</h4>
      <div className="space-y-3">
        <div className="p-3 border rounded">
          <div className="flex justify-between items-center">
            <Text variant="subtext" theme="neutral">Historical (Jan 1, 2022):</Text>
            <Time seconds={1640995200} />
          </div>
          <Text variant="label" theme="neutral">Unix timestamp: 1640995200</Text>
        </div>
        
        <div className="p-3 border rounded">
          <div className="flex justify-between items-center">
            <Text variant="subtext" theme="neutral">Recent Past (1 hour ago):</Text>
            <Time seconds={Math.floor(Date.now() / 1000) - 3600} />
          </div>
          <Text variant="label" theme="neutral">Unix timestamp: {Math.floor(Date.now() / 1000) - 3600}</Text>
        </div>
        
        <div className="p-3 border rounded">
          <div className="flex justify-between items-center">
            <Text variant="subtext" theme="neutral">Future (in 2 hours):</Text>
            <Time seconds={Math.floor(Date.now() / 1000) + 7200} />
          </div>
          <Text variant="label" theme="neutral">Unix timestamp: {Math.floor(Date.now() / 1000) + 7200}</Text>
        </div>
      </div>
    </div>
  </div>
)

export const TimestampFormats = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Timestamp with Different Formats</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The same Unix timestamp can be displayed in various formats
        depending on the interface context. This example shows how
        the same moment (Jan 1, 2022) appears in each format option.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Same Time, Different Formats</h4>
      <div className="space-y-2">
        <div className="flex items-center justify-between p-3 border rounded">
          <Text variant="subtext" theme="neutral">Short DateTime:</Text>
          <Time seconds={1640995200} format="short-datetime" />
        </div>
        <div className="flex items-center justify-between p-3 border rounded">
          <Text variant="subtext" theme="neutral">Long DateTime:</Text>
          <Time seconds={1640995200} format="long-datetime" />
        </div>
        <div className="flex items-center justify-between p-3 border rounded">
          <Text variant="subtext" theme="neutral">Relative Time:</Text>
          <Time seconds={1640995200} format="relative" />
        </div>
        <div className="flex items-center justify-between p-3 border rounded">
          <Text variant="subtext" theme="neutral">Time Only:</Text>
          <Time seconds={1640995200} format="time-only" />
        </div>
        <div className="flex items-center justify-between p-3 border rounded">
          <Text variant="subtext" theme="neutral">Log DateTime:</Text>
          <Time seconds={1640995200} format="log-datetime" />
        </div>
      </div>
      <Text variant="label" theme="neutral">
        All formats show Unix timestamp 1640995200 (January 1, 2022, 00:00:00 UTC)
      </Text>
    </div>
  </div>
)

export const InputTypes = () => {
  const timestamp = 1640995200
  const isoString = '2022-01-01T00:00:00.000Z'

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Input Type Comparison</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          The Time component accepts both Unix timestamps (seconds prop)
          and ISO date strings (time prop). Both inputs represent the same
          moment in time but come from different data sources.
        </p>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Same Time, Different Input Methods</h4>
        <div className="space-y-3">
          <div className="p-4 border rounded">
            <div className="flex justify-between items-center mb-2">
              <Text weight="strong">From Unix Timestamp:</Text>
              <Time seconds={timestamp} />
            </div>
            <Text variant="label" theme="neutral" family="mono">
              seconds={timestamp}
            </Text>
          </div>
          
          <div className="p-4 border rounded">
            <div className="flex justify-between items-center mb-2">
              <Text weight="strong">From ISO String:</Text>
              <Time time={isoString} />
            </div>
            <Text variant="label" theme="neutral" family="mono">
              time="{isoString}"
            </Text>
          </div>
        </div>
        
        <div className="p-3 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded">
          <Text variant="subtext" theme="info">
            Both inputs represent exactly the same moment: January 1, 2022, 00:00:00 UTC
          </Text>
        </div>
      </div>
    </div>
  )
}

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Time components are commonly used in tables, activity feeds,
        logs, and anywhere timestamps need to be displayed. Choose
        formats based on context and available space.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Activity Feed</h4>
      <div className="p-4 border rounded space-y-3">
        <div className="flex justify-between items-center">
          <div>
            <Text weight="strong">John deployed v2.1.0</Text>
            <Text variant="subtext" theme="neutral">Production environment</Text>
          </div>
          <Time seconds={Math.floor(Date.now() / 1000) - 1200} format="relative" />
        </div>
        
        <div className="flex justify-between items-center">
          <div>
            <Text weight="strong">Sarah created new API key</Text>
            <Text variant="subtext" theme="neutral">Development environment</Text>
          </div>
          <Time seconds={Math.floor(Date.now() / 1000) - 3600} format="relative" />
        </div>
        
        <div className="flex justify-between items-center">
          <div>
            <Text weight="strong">Mike updated documentation</Text>
            <Text variant="subtext" theme="neutral">README.md</Text>
          </div>
          <Time seconds={Math.floor(Date.now() / 1000) - 7200} format="relative" />
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">System Logs</h4>
      <div className="p-4 border rounded space-y-2 font-mono text-sm">
        <div className="flex gap-4">
          <Time seconds={Math.floor(Date.now() / 1000) - 300} format="log-datetime" />
          <Badge theme="info" size="sm">INFO</Badge>
          <Text>Application started successfully</Text>
        </div>
        <div className="flex gap-4">
          <Time seconds={Math.floor(Date.now() / 1000) - 240} format="log-datetime" />
          <Badge theme="warn" size="sm">WARN</Badge>
          <Text>High memory usage detected: 85%</Text>
        </div>
        <div className="flex gap-4">
          <Time seconds={Math.floor(Date.now() / 1000) - 180} format="log-datetime" />
          <Badge theme="error" size="sm">ERROR</Badge>
          <Text>Failed to connect to external API</Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Scheduled Events</h4>
      <div className="p-4 border rounded space-y-3">
        <div className="flex justify-between items-center">
          <div>
            <Text weight="strong">Daily Backup</Text>
            <Text variant="subtext" theme="neutral">Automated database backup</Text>
          </div>
          <div className="text-right">
            <Text variant="subtext">Next run:</Text>
            <Time seconds={Math.floor(Date.now() / 1000) + 3600} format="time-only" />
          </div>
        </div>
        
        <div className="flex justify-between items-center">
          <div>
            <Text weight="strong">Security Scan</Text>
            <Text variant="subtext" theme="neutral">Weekly vulnerability check</Text>
          </div>
          <div className="text-right">
            <Text variant="subtext">Last run:</Text>
            <Time seconds={Math.floor(Date.now() / 1000) - 86400} format="short-datetime" />
          </div>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use relative format for recent activities and timelines</li>
        <li>Use short-datetime for tables and compact displays</li>
        <li>Use long-datetime when full context is needed</li>
        <li>Use time-only for schedules and recurring events</li>
        <li>Use log-datetime for debugging and technical logs</li>
        <li>Consider timezone implications for user-facing timestamps</li>
      </ul>
    </div>
  </div>
)
