import { Button } from './Button'
import { KeyValueList, KeyValueListSkeleton } from './KeyValueList'

export const Default = () => (
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
)

export const WithLongValues = () => (
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
)

export const WithEmptyValues = () => (
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
)

export const ServerConfiguration = () => (
  <div className="max-w-2xl">
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
)

export const Loading = () => (
  <div className="max-w-2xl">
    <KeyValueListSkeleton count={6} />
  </div>
)

export const LoadingFewItems = () => (
  <div className="max-w-2xl">
    <KeyValueListSkeleton count={3} />
  </div>
)

export const EmptyState = () => (
  <div className="max-w-2xl">
    <KeyValueList values={[]} />
  </div>
)

export const EmptyStateWithCustomMessage = () => (
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
)

export const EmptyStateWithActions = () => (
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
)

export const EmptyStateComparison = () => (
  <div className="space-y-8">
    <div>
      <h3 className="text-lg font-semibold mb-4">With Data</h3>
      <div className="max-w-2xl">
        <KeyValueList
          values={[
            { key: 'NODE_ENV', value: 'production' },
            { key: 'PORT', value: '3000' },
          ]}
        />
      </div>
    </div>

    <div>
      <h3 className="text-lg font-semibold mb-4">Empty State (Default)</h3>
      <div className="max-w-2xl">
        <KeyValueList values={[]} />
      </div>
    </div>

    <div>
      <h3 className="text-lg font-semibold mb-4">Empty State (Custom)</h3>
      <div className="max-w-2xl">
        <KeyValueList
          values={[]}
          emptyStateProps={{
            variant: 'table',
            size: 'sm',
            emptyTitle: 'No data available',
            emptyMessage:
              'This shows how the empty state looks with custom messaging',
          }}
        />
      </div>
    </div>
  </div>
)
