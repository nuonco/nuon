'use client'

import { useEffect, useState, type ReactNode } from 'react'
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  flexRender,
  ColumnDef,
  SortingState,
} from '@tanstack/react-table'
import { PaginationProvider } from '@/providers/pagination-provider'
import { usePagination } from '@/hooks/use-pagination'
import { cn } from '@/utils/classnames'
import { DebouncedSearchInput } from './DeboundedSearch'
import { EmptyState, type IEmptyState } from './EmptyState'
import { Icon } from './Icon'
import { Pagination, type IPagination } from './Pagination'
import { Skeleton } from './Skeleton'

// Skeleton cell for loading state
function SkeletonCell() {
  return <Skeleton height="24px" width="100%" />
}

export interface ITable<TData extends object> {
  className?: string
  columns: ColumnDef<TData, any>[]
  data: TData[]
  emptyMessage?: string
  emptyStateProps?: Omit<IEmptyState, 'variant'>
  enableSorting?: boolean
  filterActions?: ReactNode
  isLoading?: boolean
  pagination: Omit<IPagination, 'position'>
  searchPlaceholder?: string
  skeletonRows?: number
}

export function TableBase<TData extends object>({
  className = '',
  columns,
  data,
  emptyMessage = 'No data found',
  emptyStateProps = { emptyMessage: 'No data found' },
  enableSorting = true,
  filterActions,
  isLoading = false, // default isLoading to false
  pagination,
  searchPlaceholder,
  skeletonRows = 5, // default skeleton row count
}: ITable<TData>) {
  const { isPaginating, setIsPaginating } = usePagination()
  const [sorting, setSorting] = useState<SortingState>([])

  const table = useReactTable({
    data,
    columns,
    state: { sorting },
    onSortingChange: enableSorting ? setSorting : undefined,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: enableSorting ? getSortedRowModel() : undefined,
  })

  useEffect(() => {
    setIsPaginating(false)
  }, [data])

  const renderSkeletonRows = () =>
    Array.from({ length: pagination?.limit || skeletonRows }).map((_, i) => (
      <tr key={`skeleton-row-${i}`}>
        {columns.map((_, j) => (
          <td key={`skeleton-cell-${j}`} className="py-3 px-4 border-t">
            <SkeletonCell />
          </td>
        ))}
      </tr>
    ))

  return (
    <div className="flex flex-col gap-4 md:gap-6 w-full">
      <div className="flex flex-row flex-wrap justify-between">
        <DebouncedSearchInput placeholder={searchPlaceholder} />
        {filterActions ? (
          <div className="flex gap-4 md:gap-6">{filterActions}</div>
        ) : null}
      </div>
      <div className={`rounded-lg border ${className}`}>
        <table className="min-w-full text-sm">
          <thead className="rounded-lg">
            {table.getHeaderGroups().map((headerGroup) => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map((header, idx) => (
                  <th
                    key={header.id}
                    className={cn(
                      'py-3 px-4 text-left bg-cool-grey-100 dark:bg-dark-grey-700',
                      {
                        'cursor-pointer select-none':
                          header.column.getCanSort(),
                        'rounded-tl-lg': idx === 0,
                        'rounded-tr-lg': idx === headerGroup.headers.length - 1,
                      }
                    )}
                    onClick={
                      header.column.getCanSort()
                        ? header.column.getToggleSortingHandler()
                        : undefined
                    }
                  >
                    <span className="flex gap-1 items-center">
                      <span className="font-normal font-sans">
                        {flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                      </span>
                      {header.column.getCanSort() && (
                        <span>
                          {header.column.getIsSorted() === 'asc' ? (
                            <Icon variant="SortDescending" />
                          ) : header.column.getIsSorted() === 'desc' ? (
                            <Icon variant="SortAscending" />
                          ) : (
                            ''
                          )}
                        </span>
                      )}
                    </span>
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody>
            {isLoading || isPaginating ? (
              renderSkeletonRows()
            ) : table.getRowModel().rows.length === 0 ? (
              <tr>
                <td colSpan={columns.length} className="text-center py-8">
                  <EmptyState variant="table" {...emptyStateProps} />
                </td>
              </tr>
            ) : (
              table.getRowModel().rows.map((row) => (
                <tr key={row.id}>
                  {row.getVisibleCells().map((cell) => (
                    <td key={cell.id} className="py-3 px-4 border-t">
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
        {pagination?.hasNext || pagination?.offset !== 0 ? (
          <div className="flex items-center justify-center w-full border-t py-2">
            <Pagination {...pagination} />
          </div>
        ) : null}
      </div>
    </div>
  )
}

export const Table = <T extends object>(props: ITable<T>) => (
  <PaginationProvider>
    <TableBase<T> {...props} />
  </PaginationProvider>
)
