import { ColumnDef } from '@tanstack/react-table'
import { Table } from './Table'
import { Link } from './Link'
import { Text } from './Text'
import { Badge } from './Badge'
import { Status } from './Status'
import { Button } from './Button'

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

export const BasicTable = () => (
  <Table
    data={sampleUsers}
    columns={userColumns}
    pagination={basicPagination}
  />
)

export const TableWithSearch = () => (
  <Table
    data={sampleUsers}
    columns={userColumns}
    pagination={basicPagination}
    searchPlaceholder="Search users..."
  />
)

export const TableWithCustomEmptyMessage = () => (
  <Table
    data={[]}
    columns={userColumns}
    pagination={basicPagination}
    emptyMessage="No users found. Add some users to get started."
  />
)

export const LoadingTable = () => (
  <Table
    data={[]}
    columns={userColumns}
    pagination={basicPagination}
    isLoading={true}
    skeletonRows={3}
  />
)

export const TableWithSortingDisabled = () => (
  <Table
    data={sampleUsers}
    columns={userColumns}
    pagination={basicPagination}
    enableSorting={false}
  />
)

export const AppTable = () => (
  <Table
    data={sampleApps}
    columns={appColumns}
    pagination={basicPagination}
    searchPlaceholder="Search apps..."
  />
)

export const TableWithFilterActions = () => (
  <Table
    data={sampleUsers}
    columns={userColumns}
    pagination={basicPagination}
    searchPlaceholder="Search users..."
    filterActions={
      <div className="flex gap-2">
        <Button variant="ghost" size="sm">
          Filter by Role
        </Button>
        <Button variant="ghost" size="sm">
          Filter by Status
        </Button>
      </div>
    }
  />
)

export const TableWithManyRows = () => {
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
    <Table
      data={manyUsers}
      columns={userColumns}
      pagination={{
        limit: 10,
        offset: 0,
      }}
      searchPlaceholder="Search from 50 users..."
    />
  )
}

export const TableWithCustomClassName = () => (
  <Table
    data={sampleUsers}
    columns={userColumns}
    pagination={basicPagination}
    className="shadow-lg"
    searchPlaceholder="Search users..."
  />
)

export const CompleteTableExample = () => (
  <Table
    data={sampleApps}
    columns={appColumns}
    pagination={basicPagination}
    searchPlaceholder="Search applications..."
    filterActions={
      <div className="flex gap-2">
        <Button variant="ghost" size="sm">
          All Platforms
        </Button>
        <Button variant="ghost" size="sm">
          Status Filter
        </Button>
        <Button variant="primary" size="sm">
          Deploy New
        </Button>
      </div>
    }
    className="shadow-sm"
  />
)

export const MinimalTable = () => {
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
    <Table
      data={sampleUsers}
      columns={minimalColumns}
      pagination={basicPagination}
    />
  )
}
