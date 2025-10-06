import { ColumnDef } from '@tanstack/react-table'
import { TableSkeleton } from './TableSkeleton'
import { Table } from './Table'
import { Button } from './Button'

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

export const Default = () => <TableSkeleton columns={userColumns} />

export const WithCustomRowCount = () => (
  <div className="flex flex-col gap-8">
    <div>
      <h3 className="mb-4 text-lg font-semibold">3 Rows</h3>
      <TableSkeleton columns={userColumns} skeletonRows={3} />
    </div>
    <div>
      <h3 className="mb-4 text-lg font-semibold">8 Rows</h3>
      <TableSkeleton columns={userColumns} skeletonRows={8} />
    </div>
    <div>
      <h3 className="mb-4 text-lg font-semibold">1 Row</h3>
      <TableSkeleton columns={userColumns} skeletonRows={1} />
    </div>
  </div>
)

export const DifferentColumnLayouts = () => (
  <div className="flex flex-col gap-8">
    <div>
      <h3 className="mb-4 text-lg font-semibold">Two Columns</h3>
      <TableSkeleton columns={twoColumns} skeletonRows={3} />
    </div>
    <div>
      <h3 className="mb-4 text-lg font-semibold">Five Columns</h3>
      <TableSkeleton columns={userColumns} skeletonRows={3} />
    </div>
    <div>
      <h3 className="mb-4 text-lg font-semibold">Six Columns</h3>
      <TableSkeleton columns={manyColumns} skeletonRows={3} />
    </div>
  </div>
)

export const AppTableSkeleton = () => (
  <TableSkeleton columns={appColumns} skeletonRows={4} />
)

export const LongLoadingSkeleton = () => (
  <TableSkeleton columns={userColumns} skeletonRows={10} />
)

export const SkeletonWithSearchAndFilters = () => (
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
          Filter by Role
        </Button>
        <Button variant="ghost" size="sm">
          Filter by Status
        </Button>
      </div>
    }
  />
)

export const SkeletonWithSearchOnly = () => (
  <Table
    data={[]}
    columns={appColumns}
    pagination={{ limit: 3, offset: 0 }}
    isLoading={true}
    skeletonRows={3}
    searchPlaceholder="Search applications..."
  />
)
