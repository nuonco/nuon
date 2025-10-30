/* eslint-disable react/no-unescaped-entities */
import { ColumnDef } from '@tanstack/react-table'
import { Table } from './Table'
import { Link } from './Link'
import { Text } from './Text'
import { Badge } from './Badge'
import { Status } from './Status'
import { Button } from './Button'
import { Icon } from './Icon'

// Sample data types
type SampleUser = {
  id: string
  name: string
  email: string
  role: string
  status: 'active' | 'inactive' | 'pending'
  joinDate: string
  profileHref: string
}

type SampleApp = {
  id: string
  name: string
  platform: string
  version: string
  status: 'success' | 'failed' | 'in-progress'
  deployedAt: string
  nameHref: string
  actionHref: string
}

// Sample users data
const sampleUsers: SampleUser[] = [
  {
    id: 'user-1',
    name: 'John Doe',
    email: 'john@example.com',
    role: 'Admin',
    status: 'active',
    joinDate: '2024-01-15',
    profileHref: '/users/user-1',
  },
  {
    id: 'user-2',
    name: 'Jane Smith',
    email: 'jane@example.com',
    role: 'Developer',
    status: 'active',
    joinDate: '2024-02-20',
    profileHref: '/users/user-2',
  },
  {
    id: 'user-3',
    name: 'Bob Johnson',
    email: 'bob@example.com',
    role: 'Designer',
    status: 'inactive',
    joinDate: '2024-03-10',
    profileHref: '/users/user-3',
  },
  {
    id: 'user-4',
    name: 'Alice Brown',
    email: 'alice@example.com',
    role: 'Manager',
    status: 'pending',
    joinDate: '2024-04-05',
    profileHref: '/users/user-4',
  },
]

// Sample apps data
const sampleApps: SampleApp[] = [
  {
    id: 'app-1',
    name: 'Web Dashboard',
    platform: 'AWS',
    version: 'v1.2.3',
    status: 'success',
    deployedAt: '2024-07-15T10:00:00Z',
    nameHref: '/apps/app-1',
    actionHref: '/apps/app-1/details',
  },
  {
    id: 'app-2',
    name: 'Mobile API',
    platform: 'Azure',
    version: 'v2.0.1',
    status: 'failed',
    deployedAt: '2024-07-14T15:30:00Z',
    nameHref: '/apps/app-2',
    actionHref: '/apps/app-2/details',
  },
  {
    id: 'app-3',
    name: 'Analytics Service',
    platform: 'GCP',
    version: 'v1.5.0',
    status: 'in-progress',
    deployedAt: '2024-07-13T08:15:00Z',
    nameHref: '/apps/app-3',
    actionHref: '/apps/app-3/details',
  },
]

// User columns definition
const userColumns: ColumnDef<SampleUser>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
    cell: (info) => (
      <Link href={info.row.original.profileHref}>
        {info.getValue() as string}
      </Link>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'email',
    header: 'Email',
    cell: (info) => (
      <Text family="mono" theme="neutral">
        {info.getValue() as string}
      </Text>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'role',
    header: 'Role',
    cell: (info) => <Badge theme="info">{info.getValue() as string}</Badge>,
    enableSorting: true,
  },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: (info) => {
      const status = info.getValue() as string
      return (
        <Status
          status={
            status === 'active'
              ? 'success'
              : status === 'inactive'
                ? 'error'
                : 'warn'
          }
        />
      )
    },
    enableSorting: true,
  },
  {
    accessorKey: 'joinDate',
    header: 'Join Date',
    cell: (info) => (
      <Text theme="neutral">
        {new Date(info.getValue() as string).toLocaleDateString()}
      </Text>
    ),
    enableSorting: true,
  },
]

// App columns definition
const appColumns: ColumnDef<SampleApp>[] = [
  {
    accessorKey: 'name',
    header: 'App Name',
    cell: (info) => (
      <Link href={info.row.original.nameHref}>{info.getValue() as string}</Link>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'id',
    header: 'App ID',
    cell: (info) => (
      <Text family="mono" theme="neutral">
        {info.getValue() as string}
      </Text>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'platform',
    header: 'Platform',
    cell: (info) => <Badge theme="info">{info.getValue() as string}</Badge>,
    enableSorting: true,
  },
  {
    accessorKey: 'version',
    header: 'Version',
    cell: (info) => <Text family="mono">{info.getValue() as string}</Text>,
    enableSorting: true,
  },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: (info) => {
      const status = info.getValue() as string
      return <Status status={status as any} />
    },
    enableSorting: true,
  },
  {
    accessorKey: 'actionHref',
    header: 'Action',
    cell: (info) => <Link href={info.getValue() as string}>View Details</Link>,
    enableSorting: false,
  },
]

// Basic pagination
const basicPagination = {
  limit: 10,
  offset: 0,
}

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Table Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The Table component provides a powerful data grid with sorting,
        searching, and pagination capabilities. Built on TanStack Table,
        it supports custom cell rendering, loading states, and responsive
        layouts for displaying structured data.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Simple User Table</h4>
      <Table
        data={sampleUsers}
        columns={userColumns}
        pagination={basicPagination}
      />
      <Text variant="subtext" theme="neutral">
        Click column headers to sort by different fields
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Built on TanStack Table for robust data handling</li>
        <li>Sortable columns with visual indicators</li>
        <li>Pagination with configurable page sizes</li>
        <li>Loading states with skeleton placeholders</li>
        <li>Search functionality with highlighting</li>
        <li>Custom cell rendering for complex content</li>
      </ul>
    </div>
  </div>
)

export const SearchAndFiltering = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Search and Filtering</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Tables support real-time search across all columns and custom
        filter actions. Search automatically highlights matching text
        and provides instant results as you type.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Table with Search</h4>
      <Table
        data={sampleUsers}
        columns={userColumns}
        pagination={basicPagination}
        searchPlaceholder="Search users..."
      />
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Table with Filter Actions</h4>
      <Table
        data={sampleUsers}
        columns={userColumns}
        pagination={basicPagination}
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
    </div>
  </div>
)

export const LoadingAndEmptyStates = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Loading and Empty States</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Tables provide skeleton loading states while data is fetching
        and customizable empty states when no data is available.
        Empty states can include custom messages and actions.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Loading State</h4>
      <Table
        data={[]}
        columns={userColumns}
        pagination={basicPagination}
        isLoading={true}
        skeletonRows={4}
        searchPlaceholder="Search users..."
      />
      <Text variant="subtext" theme="neutral">
        Skeleton rows match the table structure during loading
      </Text>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Custom Empty State</h4>
      <Table
        data={[]}
        columns={userColumns}
        pagination={basicPagination}
        emptyMessage="No users found. Invite team members to get started."
        searchPlaceholder="Search users..."
      />
      <Text variant="subtext" theme="neutral">
        Custom empty messages provide context and guidance
      </Text>
    </div>
  </div>
)

export const SortingConfiguration = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Sorting Configuration</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Table sorting can be enabled or disabled globally and per-column.
        Sort indicators show current sort direction and allow users to
        cycle through ascending, descending, and unsorted states.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Sorting Enabled (Default)</h4>
      <Table
        data={sampleUsers}
        columns={userColumns}
        pagination={basicPagination}
      />
      <Text variant="subtext" theme="neutral">
        Click headers to sort. Notice the sort indicators on hoverable columns.
      </Text>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Sorting Disabled</h4>
      <Table
        data={sampleUsers}
        columns={userColumns}
        pagination={basicPagination}
        enableSorting={false}
      />
      <Text variant="subtext" theme="neutral">
        When sorting is disabled, headers are not clickable
      </Text>
    </div>
  </div>
)

export const CustomCellRendering = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Custom Cell Rendering</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Table supports rich cell content including links, badges, status
        indicators, and custom components. This example shows applications
        with various data types and interactive elements.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Applications Table</h4>
      <Table
        data={sampleApps}
        columns={appColumns}
        pagination={basicPagination}
        searchPlaceholder="Search applications..."
        filterActions={
          <div className="flex gap-2">
            <Button variant="ghost" size="sm">
              <Icon variant="Sliders" size="16" />
              Platform
            </Button>
            <Button variant="ghost" size="sm">
              <Icon variant="Sliders" size="16" />
              Status
            </Button>
            <Button variant="primary" size="sm">
              <Icon variant="Plus" size="16" />
              New App
            </Button>
          </div>
        }
      />
      <Text variant="subtext" theme="neutral">
        Notice the Links, Badges, Status indicators, and action buttons
      </Text>
    </div>
  </div>
)

export const PaginationAndLargeDatasets = () => {
  const manyUsers = Array.from({ length: 50 }, (_, i) => ({
    id: `user-${i + 1}`,
    name: `User ${i + 1}`,
    email: `user${i + 1}@example.com`,
    role: ['Admin', 'Developer', 'Designer', 'Manager'][i % 4],
    status: ['active', 'inactive', 'pending'][i % 3] as any,
    joinDate: new Date(2024, 0, i + 1).toISOString().split('T')[0],
    profileHref: `/users/user-${i + 1}`,
  }))

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Pagination and Large Datasets</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Tables handle large datasets efficiently with pagination controls.
          Page size is configurable and pagination shows current page
          information. Search and sorting work across all data, not just
          the current page.
        </p>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">50 Users with Pagination</h4>
        <Table
          data={manyUsers}
          columns={userColumns}
          pagination={{
            limit: 10,
            offset: 0,
          }}
          searchPlaceholder="Search from 50 users..."
        />
        <Text variant="subtext" theme="neutral">
          Navigate pages using the pagination controls at the bottom
        </Text>
      </div>
    </div>
  )
}

export const ResponsiveAndStyling = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Responsive Design and Styling</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Tables are responsive by default and support custom styling
        through className props. They adapt to different screen sizes
        while maintaining functionality and readability.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Table with Custom Styling</h4>
      <Table
        data={sampleUsers}
        columns={userColumns}
        pagination={basicPagination}
        className="shadow-lg border-2 border-gray-200 dark:border-gray-700"
        searchPlaceholder="Search users..."
      />
      <Text variant="subtext" theme="neutral">
        Custom shadows, borders, and spacing can be applied via className
      </Text>
    </div>
  </div>
)

export const MinimalConfiguration = () => {
  const minimalColumns: ColumnDef<SampleUser>[] = [
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

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Minimal Configuration</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Tables can be configured with minimal column definitions for
          simple use cases. This example shows the bare minimum setup
          with just data, columns, and pagination.
        </p>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Basic Two-Column Table</h4>
        <Table
          data={sampleUsers}
          columns={minimalColumns}
          pagination={basicPagination}
        />
        <Text variant="subtext" theme="neutral">
          Minimal setup with just name and email columns
        </Text>
      </div>
    </div>
  )
}

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Tables are commonly used for data management interfaces,
        dashboards, admin panels, and anywhere structured data
        needs to be displayed with interaction capabilities.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Complete Feature Example</h4>
      <Table
        data={sampleApps}
        columns={appColumns}
        pagination={basicPagination}
        searchPlaceholder="Search applications..."
        filterActions={
          <div className="flex gap-2">
            <Button variant="ghost" size="sm">
              <Icon variant="CloudArrowDown" size="16" />
              All Platforms
            </Button>
            <Button variant="ghost" size="sm">
              <Icon variant="Pulse" size="16" />
              Status Filter
            </Button>
            <Button variant="primary" size="sm">
              <Icon variant="Plus" size="16" />
              Deploy New
            </Button>
          </div>
        }
        className="shadow-sm"
      />
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Define column widths appropriately for content types</li>
        <li>Use consistent cell components (Link, Badge, Status) for similar data</li>
        <li>Provide meaningful search placeholders and empty states</li>
        <li>Consider mobile responsiveness for wide tables</li>
        <li>Enable sorting on columns where it makes sense</li>
        <li>Use skeleton loading states for better perceived performance</li>
        <li>Group related filter actions logically</li>
      </ul>
    </div>
  </div>
)
