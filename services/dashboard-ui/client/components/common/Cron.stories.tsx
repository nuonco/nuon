/* eslint-disable react/no-unescaped-entities */
import { Cron } from './Cron'
import { Text } from './Text'
import { Icon } from './Icon'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Cron Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The Cron component displays cron expressions in human-readable formats
        using the cronstrue library. It automatically converts cron syntax to
        plain English, with tooltips showing the raw expression. The component
        handles invalid expressions gracefully with appropriate fallbacks.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Common Cron Expressions</h4>
      <div className="space-y-3 p-4 border rounded">
        <div className="flex items-center justify-between">
          <Cron cron="0 */6 * * *" />
          <Text variant="subtext" theme="neutral">
            Every 6 hours
          </Text>
        </div>
        <div className="flex items-center justify-between">
          <Cron cron="0 0 * * *" />
          <Text variant="subtext" theme="neutral">
            Daily at midnight
          </Text>
        </div>
        <div className="flex items-center justify-between">
          <Cron cron="*/15 * * * *" />
          <Text variant="subtext" theme="neutral">
            Every 15 minutes
          </Text>
        </div>
        <div className="flex items-center justify-between">
          <Cron cron="0 9 * * 1" />
          <Text variant="subtext" theme="neutral">
            Weekly on Monday at 9 AM
          </Text>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Automatic conversion of cron expressions to human-readable text</li>
        <li>Tooltip support showing raw cron expression</li>
        <li>Multiple display formats (human, expression, both)</li>
        <li>Graceful handling of invalid cron expressions</li>
        <li>Consistent styling with other time-related components</li>
      </ul>
    </div>
  </div>
)

export const Formats = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Cron Display Formats</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          format
        </code>{' '}
        prop controls how cron expressions are displayed. Choose between
        human-readable text, raw expression, or both depending on your use case
        and audience technical level.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Format Comparison</h4>
      <div className="grid grid-cols-1 gap-4">
        <div className="p-4 border rounded">
          <Text weight="strong" className="mb-3">
            Human Format (Default)
          </Text>
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Cron cron="0 */6 * * *" format="human" />
              <Text variant="subtext">User-friendly display</Text>
            </div>
            <div className="flex items-center justify-between">
              <Cron cron="*/30 * * * *" format="human" />
              <Text variant="subtext">With tooltip on hover</Text>
            </div>
          </div>
        </div>

        <div className="p-4 border rounded">
          <Text weight="strong" className="mb-3">
            Expression Format
          </Text>
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Cron cron="0 */6 * * *" format="expression" />
              <Text variant="subtext">Raw cron syntax</Text>
            </div>
            <div className="flex items-center justify-between">
              <Cron cron="*/30 * * * *" format="expression" />
              <Text variant="subtext">Monospace font</Text>
            </div>
          </div>
        </div>

        <div className="p-4 border rounded">
          <Text weight="strong" className="mb-3">
            Both Formats
          </Text>
          <div className="space-y-3">
            <Cron cron="0 */6 * * *" format="both" />
            <Cron cron="*/30 * * * *" format="both" />
          </div>
        </div>
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>human:</strong> Human-readable format for end users and
        non-technical stakeholders
      </div>
      <div>
        <strong>expression:</strong> Raw cron syntax for technical users and
        configuration displays
      </div>
      <div>
        <strong>both:</strong> Combined display showing both human text and raw
        expression for documentation
      </div>
    </div>
  </div>
)

export const CommonSchedules = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Schedule Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Cron expressions support a wide variety of scheduling patterns. Here
        are the most commonly used patterns in production systems for tasks
        like backups, reports, and maintenance operations.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Frequent Schedules</h4>
      <div className="space-y-3 p-4 border rounded">
        <div className="flex items-center justify-between">
          <Cron cron="* * * * *" />
          <Text variant="subtext" theme="neutral">
            Every minute
          </Text>
        </div>
        <div className="flex items-center justify-between">
          <Cron cron="*/5 * * * *" />
          <Text variant="subtext" theme="neutral">
            Every 5 minutes
          </Text>
        </div>
        <div className="flex items-center justify-between">
          <Cron cron="*/15 * * * *" />
          <Text variant="subtext" theme="neutral">
            Every 15 minutes
          </Text>
        </div>
        <div className="flex items-center justify-between">
          <Cron cron="0 * * * *" />
          <Text variant="subtext" theme="neutral">
            Every hour
          </Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Daily Schedules</h4>
      <div className="space-y-3 p-4 border rounded">
        <div className="flex items-center justify-between">
          <Cron cron="0 0 * * *" />
          <Text variant="subtext" theme="neutral">
            Midnight daily
          </Text>
        </div>
        <div className="flex items-center justify-between">
          <Cron cron="0 9 * * *" />
          <Text variant="subtext" theme="neutral">
            9 AM daily
          </Text>
        </div>
        <div className="flex items-center justify-between">
          <Cron cron="0 */6 * * *" />
          <Text variant="subtext" theme="neutral">
            Every 6 hours
          </Text>
        </div>
        <div className="flex items-center justify-between">
          <Cron cron="0 0,12 * * *" />
          <Text variant="subtext" theme="neutral">
            Midnight and noon
          </Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Weekly Schedules</h4>
      <div className="space-y-3 p-4 border rounded">
        <div className="flex items-center justify-between">
          <Cron cron="0 9 * * 1" />
          <Text variant="subtext" theme="neutral">
            Monday 9 AM
          </Text>
        </div>
        <div className="flex items-center justify-between">
          <Cron cron="0 0 * * 0" />
          <Text variant="subtext" theme="neutral">
            Sunday midnight
          </Text>
        </div>
        <div className="flex items-center justify-between">
          <Cron cron="0 9 * * 1-5" />
          <Text variant="subtext" theme="neutral">
            Weekdays at 9 AM
          </Text>
        </div>
        <div className="flex items-center justify-between">
          <Cron cron="0 0 * * 6,0" />
          <Text variant="subtext" theme="neutral">
            Weekend midnight
          </Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Monthly Schedules</h4>
      <div className="space-y-3 p-4 border rounded">
        <div className="flex items-center justify-between">
          <Cron cron="0 0 1 * *" />
          <Text variant="subtext" theme="neutral">
            First day of month
          </Text>
        </div>
        <div className="flex items-center justify-between">
          <Cron cron="0 0 15 * *" />
          <Text variant="subtext" theme="neutral">
            15th of every month
          </Text>
        </div>
        <div className="flex items-center justify-between">
          <Cron cron="0 0 1 */3 *" />
          <Text variant="subtext" theme="neutral">
            Quarterly (every 3 months)
          </Text>
        </div>
        <div className="flex items-center justify-between">
          <Cron cron="0 0 1 1 *" />
          <Text variant="subtext" theme="neutral">
            First day of year
          </Text>
        </div>
      </div>
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Real-World Usage Examples</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Cron components are commonly used in drift detection schedules, backup
        configurations, and automated task management. Here are typical
        patterns for different application contexts.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Drift Detection Schedule</h4>
      <div className="space-y-3 border rounded-lg p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Icon variant="ClockClockwise" size="16" className="text-blue-600" />
            <div>
              <Text weight="strong">Infrastructure Drift Check</Text>
              <Text variant="subtext" theme="neutral">
                Compare deployed state with desired configuration
              </Text>
            </div>
          </div>
          <Cron cron="0 */6 * * *" />
        </div>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Icon variant="ClockClockwise" size="16" className="text-blue-600" />
            <div>
              <Text weight="strong">Security Compliance Scan</Text>
              <Text variant="subtext" theme="neutral">
                Validate security policies and access controls
              </Text>
            </div>
          </div>
          <Cron cron="0 2 * * *" />
        </div>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Icon variant="ClockClockwise" size="16" className="text-blue-600" />
            <div>
              <Text weight="strong">Configuration Audit</Text>
              <Text variant="subtext" theme="neutral">
                Weekly review of configuration changes
              </Text>
            </div>
          </div>
          <Cron cron="0 9 * * 1" />
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Maintenance Windows</h4>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="p-4 border rounded">
          <Text weight="strong">Backup Schedules</Text>
          <div className="mt-3 space-y-2">
            <div className="flex justify-between items-center">
              <Text variant="subtext">Database backup:</Text>
              <Cron cron="0 2 * * *" />
            </div>
            <div className="flex justify-between items-center">
              <Text variant="subtext">File system backup:</Text>
              <Cron cron="0 3 * * 0" />
            </div>
            <div className="flex justify-between items-center">
              <Text variant="subtext">Log rotation:</Text>
              <Cron cron="0 0 * * *" />
            </div>
          </div>
        </div>

        <div className="p-4 border rounded">
          <Text weight="strong">System Maintenance</Text>
          <div className="mt-3 space-y-2">
            <div className="flex justify-between items-center">
              <Text variant="subtext">Cache cleanup:</Text>
              <Cron cron="0 */4 * * *" />
            </div>
            <div className="flex justify-between items-center">
              <Text variant="subtext">Index rebuild:</Text>
              <Cron cron="0 1 * * 0" />
            </div>
            <div className="flex justify-between items-center">
              <Text variant="subtext">Health checks:</Text>
              <Cron cron="*/15 * * * *" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Component Configuration</h4>
      <div className="p-4 border rounded-lg bg-gray-50 dark:bg-gray-800">
        <div className="flex items-center justify-between mb-4">
          <Text weight="stronger">Terraform Module</Text>
          <Text variant="subtext" theme="neutral">
            Drift detection enabled
          </Text>
        </div>
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <div>
              <Text variant="label" weight="strong">
                Drift Schedule
              </Text>
              <Text variant="label" theme="neutral">
                Automatic drift detection
              </Text>
            </div>
            <Cron cron="0 */6 * * *" format="both" />
          </div>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use human format for end-user interfaces and dashboards</li>
        <li>Use expression format for technical configuration pages</li>
        <li>
          Use both format for documentation and administrative interfaces
        </li>
        <li>
          Enable tooltips (default) to help users understand the raw expression
        </li>
        <li>Consider timezone implications when displaying schedules</li>
        <li>Validate cron expressions before saving to prevent errors</li>
      </ul>
    </div>
  </div>
)

export const EdgeCases = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Edge Cases & Error Handling</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The Cron component gracefully handles various edge cases including
        invalid expressions, missing values, and complex cron syntax. Here's
        how different scenarios are handled.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Missing or Empty Values</h4>
      <div className="space-y-3 p-4 border rounded">
        <div className="flex items-center justify-between">
          <Text variant="subtext">No cron provided:</Text>
          <Cron />
        </div>
        <div className="flex items-center justify-between">
          <Text variant="subtext">Empty string:</Text>
          <Cron cron="" />
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Invalid Expressions</h4>
      <div className="space-y-3 p-4 border rounded">
        <div className="flex items-center justify-between">
          <Text variant="subtext">Invalid syntax:</Text>
          <Cron cron="invalid cron" />
        </div>
        <div className="flex items-center justify-between">
          <Text variant="subtext">Malformed expression:</Text>
          <Cron cron="* * * *" />
        </div>
      </div>
      <Text variant="subtext" theme="neutral">
        Invalid expressions fall back to displaying the raw string
      </Text>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Complex Expressions</h4>
      <div className="space-y-3 p-4 border rounded">
        <div className="flex items-center justify-between">
          <Cron cron="0 0,6,12,18 * * *" />
          <Text variant="subtext" theme="neutral">
            Multiple specific hours
          </Text>
        </div>
        <div className="flex items-center justify-between">
          <Cron cron="0 9-17 * * 1-5" />
          <Text variant="subtext" theme="neutral">
            Weekday business hours
          </Text>
        </div>
        <div className="flex items-center justify-between">
          <Cron cron="0 0 1,15 * *" />
          <Text variant="subtext" theme="neutral">
            Twice monthly
          </Text>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Error Handling Strategy:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Missing values display as em dash (—)</li>
        <li>Invalid expressions fall back to raw string display</li>
        <li>No error messages shown to end users</li>
        <li>Tooltips work even with invalid expressions</li>
        <li>Component never crashes or throws errors</li>
      </ul>
    </div>
  </div>
)
