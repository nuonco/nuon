'use client'

import { ColumnDef } from '@tanstack/react-table'
import { Table } from './Table'

export function TableSkeleton<TData extends object>({
  columns,
  skeletonRows = 5,
}: {
  columns: ColumnDef<TData, any>[]
  skeletonRows?: number
}) {
  return (
    <Table<TData>
      data={[]} // Empty, no data
      columns={columns}
      isLoading={true}
      skeletonRows={skeletonRows}
      pagination={{ limit: skeletonRows, offset: 0 }}
    />
  )
}
