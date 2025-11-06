/* eslint-disable react/no-unescaped-entities */
import { Button } from './Button'
import { KeyValueList, KeyValueListSkeleton } from './KeyValueList'
import { Text } from './Text'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic KeyValueList Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        KeyValueList components display structured data in a clean, scannable
        format. They automatically handle key formatting, empty value display,
        and provide loading states and empty state management. Perfect for
        configuration settings, user profiles, and system information displays.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">User Profile Example</h4>
      <div className="max-w-2xl">
        <KeyValueList
          values={[
            { key: 'name', value: 'John Doe' },
            { key: 'email', value: 'john.doe@example.com' },
            { key: 'role', value: 'Administrator' },
            { key: 'department', value: 'Engineering' },
            { key: 'location', value: 'San Francisco, CA' },
          ]}
        />
      </div>
      <Text variant="subtext" theme="neutral">
        Keys are automatically formatted from camelCase/snake_case to readable
        labels
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Automatic key formatting and capitalization</li>
        <li>Empty value handling with fallback display</li>
        <li>Responsive layout with proper spacing</li>
        <li>Loading states with skeleton components</li>
        <li>Customizable empty states with actions</li>
      </ul>
    </div>
  </div>
)

export const LongValues = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Handling Long Values</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        KeyValueList gracefully handles long values including IDs, descriptions,
        and technical strings. Values automatically wrap or can be truncated
        based on the container width and content type.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Technical Data with Long Values</h4>
      <div className="max-w-2xl">
        <KeyValueList
          values={[
            { key: 'id', value: 'usr_1234567890abcdef1234567890abcdef' },
            {
              key: 'description',
              value:
                'This is a very long description that demonstrates how the component handles text that spans multiple lines and may need to wrap or be truncated depending on the container width.',
            },
            {
              key: 'api_key',
              value:
                'sk-proj-abcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ',
            },
            { key: 'created_at', value: '2024-01-15T10:30:45.123Z' },
          ]}
        />
      </div>
      <Text variant="subtext" theme="neutral">
        Long values wrap naturally while maintaining readability
      </Text>
    </div>
  </div>
)

export const EmptyValues = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Empty Value Handling</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        KeyValueList automatically handles empty or undefined values with
        appropriate fallback display. Empty strings show a dash, while null or
        undefined values are handled gracefully.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Mixed Data with Empty Values</h4>
      <div className="max-w-2xl">
        <KeyValueList
          values={[
            { key: 'username', value: 'johndoe' },
            { key: 'middle_name', value: '' },
            { key: 'phone', value: '+1 (555) 123-4567' },
            { key: 'fax', value: '' },
            { key: 'website', value: 'https://johndoe.dev' },
          ]}
        />
      </div>
      <Text variant="subtext" theme="neutral">
        Empty values show as dashes to maintain visual consistency
      </Text>
    </div>
  </div>
)

export const LoadingStates = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Loading States</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        KeyValueListSkeleton provides loading placeholders with customizable
        item counts. The skeleton matches the structure of the actual component
        with animated placeholders for keys and values.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Standard Loading (6 items)</h4>
      <div className="max-w-2xl">
        <KeyValueListSkeleton count={6} />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Compact Loading (3 items)</h4>
      <div className="max-w-2xl">
        <KeyValueListSkeleton count={3} />
      </div>
    </div>
  </div>
)

export const EmptyStates = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Empty State Management</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        KeyValueList provides customizable empty states when no data is
        available. Empty states can include custom messages, actions, and
        different sizing options to match your interface needs.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Default Empty State</h4>
      <div className="max-w-2xl">
        <KeyValueList values={[]} />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Custom Empty State</h4>
      <div className="max-w-2xl">
        <KeyValueList
          values={[]}
          emptyStateProps={{
            variant: 'table',
            size: 'sm',
            emptyTitle: 'No configuration found',
            emptyMessage: 'Add some key-value pairs to get started',
          }}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Empty State with Actions</h4>
      <div className="max-w-2xl">
        <KeyValueList
          values={[]}
          emptyStateProps={{
            variant: 'table',
            size: 'sm',
            emptyTitle: 'No environment variables',
            emptyMessage: 'Set up environment variables for your application',
            action: (
              <div className="flex items-center gap-4">
                <Button key="add">Add Variable</Button>
                <Button key="import">Import from File</Button>
              </div>
            ),
          }}
        />
      </div>
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        KeyValueList is commonly used for configuration settings, system
        information, user profiles, and technical specifications. Here are
        typical patterns for different contexts.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Server Configuration</h4>
      <div className="max-w-2xl p-4 border rounded-lg">
        <Text weight="stronger" className="mb-3">
          Production Server Details
        </Text>
        <KeyValueList
          values={[
            { key: 'hostname', value: 'api.example.com' },
            { key: 'port', value: '443' },
            { key: 'protocol', value: 'https' },
            { key: 'environment', value: 'production' },
            { key: 'region', value: 'us-west-2' },
            { key: 'instance_type', value: 't3.medium' },
            { key: 'disk_size', value: '20GB' },
            { key: 'memory', value: '4GB' },
          ]}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Application Settings</h4>
      <div className="max-w-2xl p-4 border rounded-lg">
        <Text weight="stronger" className="mb-3">
          Environment Variables
        </Text>
        <KeyValueList
          values={[
            { key: 'NODE_ENV', value: 'production' },
            { key: 'PORT', value: '3000' },
            {
              key: 'DATABASE_URL',
              value: 'postgresql://user:pass@localhost:5432/db',
            },
            { key: 'REDIS_URL', value: 'redis://localhost:6379' },
            { key: 'JWT_SECRET', value: '••••••••••••••••' },
            { key: 'LOG_LEVEL', value: 'info' },
          ]}
        />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">User Account Information</h4>
      <div className="max-w-2xl p-4 border rounded-lg">
        <Text weight="stronger" className="mb-3">
          Account Details
        </Text>
        <KeyValueList
          values={[
            { key: 'user_id', value: 'usr_1234567890' },
            { key: 'username', value: 'john.doe' },
            { key: 'email', value: 'john.doe@company.com' },
            { key: 'role', value: 'Administrator' },
            { key: 'department', value: 'Engineering' },
            { key: 'last_login', value: '2024-03-15 14:30:22 UTC' },
            { key: 'account_status', value: 'Active' },
            { key: 'two_factor_enabled', value: 'Yes' },
          ]}
        />
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use descriptive keys that clearly identify the data</li>
        <li>Group related information logically</li>
        <li>Provide helpful empty states with actionable guidance</li>
        <li>Use consistent formatting for similar data types</li>
        <li>Consider security when displaying sensitive information</li>
        <li>Include loading states for async data fetching</li>
      </ul>
    </div>
  </div>
)
