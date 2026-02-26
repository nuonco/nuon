/* eslint-disable react/no-unescaped-entities */
import { ColumnDef } from '@tanstack/react-table'
import { TableSkeleton } from './TableSkeleton'
import { Table } from './Table'
import { Button } from './Button'
import { Text } from './Text'
import { Icon } from './Icon'

// Sample data types for skeleton
type SampleUser = {
  id: string
  name: string
  email: string
  role: string
  status: string
  joinDate: string
}

type SampleApp = {
  id: string
  name: string
  platform: string
  version: string
  status: string
  deployedAt: string
}

// Simple columns for skeleton
const userColumns: ColumnDef<SampleUser>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
    enableSorting: true,
  },
  {
    accessorKey: 'email',
    header: 'Email',
    enableSorting: true,
  },
  {
    accessorKey: 'role',
    header: 'Role',
    enableSorting: true,
  },
  {
    accessorKey: 'status',
    header: 'Status',
    enableSorting: true,
  },
  {
    accessorKey: 'joinDate',
    header: 'Join Date',
    enableSorting: true,
  },
]

const appColumns: ColumnDef<SampleApp>[] = [
  {
    accessorKey: 'name',
    header: 'App Name',
    enableSorting: true,
  },
  {
    accessorKey: 'id',
    header: 'App ID',
    enableSorting: true,
  },
  {
    accessorKey: 'platform',
    header: 'Platform',
    enableSorting: true,
  },
  {
    accessorKey: 'version',
    header: 'Version',
    enableSorting: true,
  },
  {
    accessorKey: 'status',
    header: 'Status',
    enableSorting: true,
  },
]

// Two column layout for comparison
const twoColumns: ColumnDef<SampleUser>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
    enableSorting: true,
  },
  {
    accessorKey: 'email',
    header: 'Email',
    enableSorting: true,
  },
]

// Many columns layout
const manyColumns: ColumnDef<SampleUser>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
    enableSorting: true,
  },
  {
    accessorKey: 'email',
    header: 'Email',
    enableSorting: true,
  },
  {
    accessorKey: 'role',
    header: 'Role',
    enableSorting: true,
  },
  {
    accessorKey: 'status',
    header: 'Status',
    enableSorting: true,
  },
  {
    accessorKey: 'joinDate',
    header: 'Join Date',
    enableSorting: true,
  },
  {
    accessorKey: 'id',
    header: 'ID',
    enableSorting: true,
  },
]

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic TableSkeleton Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        TableSkeleton provides loading placeholders that match the structure of
        your actual table data. It uses the same column definitions to create
        skeleton rows with animated shimmer effects, maintaining visual
        consistency during data loading states.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Default Skeleton (5 rows)</h4>
      <TableSkeleton columns={userColumns} />
      <Text variant="subtext" theme="neutral">
        Skeleton automatically matches your table's column structure
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Matches actual table structure using column definitions</li>
        <li>Animated shimmer effects for realistic loading appearance</li>
        <li>Configurable number of skeleton rows</li>
        <li>Consistent spacing and layout with real table data</li>
        <li>Dark mode support with appropriate contrast</li>
      </ul>
    </div>
  </div>
)

export const CustomRowCounts = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Custom Row Counts</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The number of skeleton rows can be customized based on your expected
        data size and loading context. Use fewer rows for quick operations and
        more rows for larger datasets or longer loading times.
      </p>
    </div>

    <div className="space-y-6">
      <div className="space-y-3">
        <h4 className="text-sm font-medium">Compact Loading (3 rows)</h4>
        <TableSkeleton columns={userColumns} skeletonRows={3} />
        <Text variant="subtext" theme="neutral">
          Good for quick loading scenarios or small datasets
        </Text>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium">Extended Loading (8 rows)</h4>
        <TableSkeleton columns={userColumns} skeletonRows={8} />
        <Text variant="subtext" theme="neutral">
          Suitable for larger datasets or longer loading operations
        </Text>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium">Single Row Loading</h4>
        <TableSkeleton columns={userColumns} skeletonRows={1} />
        <Text variant="subtext" theme="neutral">
          Minimal loading state for single item scenarios
        </Text>
      </div>
    </div>
  </div>
)

export const DifferentColumnLayouts = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Different Column Layouts</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        TableSkeleton adapts to any column configuration, creating appropriate
        skeleton placeholders for narrow or wide tables. The skeleton
        automatically adjusts column widths and spacing to match your actual
        table layout.
      </p>
    </div>

    <div className="space-y-6">
      <div className="space-y-3">
        <h4 className="text-sm font-medium">Narrow Table (2 columns)</h4>
        <TableSkeleton columns={twoColumns} skeletonRows={3} />
        <Text variant="subtext" theme="neutral">
          Compact layout suitable for simple data displays
        </Text>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium">Standard Table (5 columns)</h4>
        <TableSkeleton columns={userColumns} skeletonRows={3} />
        <Text variant="subtext" theme="neutral">
          Typical table width for most data management interfaces
        </Text>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium">Wide Table (6 columns)</h4>
        <TableSkeleton columns={manyColumns} skeletonRows={3} />
        <Text variant="subtext" theme="neutral">
          Extended layout for comprehensive data views
        </Text>
      </div>
    </div>
  </div>
)

export const SpecificUseCases = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Specific Use Cases</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Different table contexts may require different skeleton configurations.
        Here are examples tailored for specific data types and loading
        scenarios.
      </p>
    </div>

    <div className="space-y-6">
      <div className="space-y-3">
        <h4 className="text-sm font-medium">Application Management Table</h4>
        <TableSkeleton columns={appColumns} skeletonRows={4} />
        <Text variant="subtext" theme="neutral">
          Optimized for application deployment and management interfaces
        </Text>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium">
          Extended Loading (Large Dataset)
        </h4>
        <TableSkeleton columns={userColumns} skeletonRows={10} />
        <Text variant="subtext" theme="neutral">
          Suitable for large datasets with longer loading times
        </Text>
      </div>
    </div>
  </div>
)

export const IntegratedTableLoading = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Integrated Table Loading</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        When using TableSkeleton within the Table component's loading state, the
        skeleton integrates seamlessly with search bars, filter actions, and
        pagination controls. This provides a complete loading experience.
      </p>
    </div>

    <div className="space-y-6">
      <div className="space-y-3">
        <h4 className="text-sm font-medium">Full Featured Loading State</h4>
        <Table
          data={[]}
          columns={userColumns}
          pagination={{ limit: 5, offset: 0 }}
          isLoading={true}
          skeletonRows={5}
          searchPlaceholder="Search users..."
          filterActions={
            <div className="flex gap-2">
              <Button variant="ghost" size="sm">
                <Icon variant="Sliders" size="16" />
                Role
              </Button>
              <Button variant="ghost" size="sm">
                <Icon variant="Sliders" size="16" />
                Status
              </Button>
              <Button variant="primary" size="sm">
                <Icon variant="Plus" size="16" />
                Add User
              </Button>
            </div>
          }
        />
        <Text variant="subtext" theme="neutral">
          Search and filter controls remain visible during loading
        </Text>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium">Search-Only Loading State</h4>
        <Table
          data={[]}
          columns={appColumns}
          pagination={{ limit: 3, offset: 0 }}
          isLoading={true}
          skeletonRows={3}
          searchPlaceholder="Search applications..."
        />
        <Text variant="subtext" theme="neutral">
          Simpler loading state with search functionality
        </Text>
      </div>
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        TableSkeleton is most commonly used within the Table component's loading
        state, but can also be used independently for custom loading scenarios.
        Choose row counts based on expected data size.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Implementation Example</h4>
      <div className="p-4 border rounded-lg bg-gray-50 dark:bg-gray-800">
        <Text variant="label" theme="neutral" className="mb-2">
          TypeScript Example:
        </Text>
        <div className="font-mono text-sm space-y-1">
          <div>{'<Table'}</div>
          <div>{'  data={isLoading ? [] : actualData}'}</div>
          <div>{'  columns={tableColumns}'}</div>
          <div>{'  pagination={pagination}'}</div>
          <div>{'  isLoading={isLoading}'}</div>
          <div>{'  skeletonRows={expectedRowCount}'}</div>
          <div>{'  searchPlaceholder="Search..."'}</div>
          <div>{'/>'}</div>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>
          Match skeleton row count to expected data size for realistic loading
        </li>
        <li>
          Use the same column definitions for skeleton as the actual table
        </li>
        <li>Consider loading time duration when choosing row count</li>
        <li>Test skeleton appearance in both light and dark modes</li>
        <li>Ensure skeleton maintains table responsive behavior</li>
        <li>Use consistent skeleton patterns across your application</li>
      </ul>
    </div>
  </div>
)
