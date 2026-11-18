import {
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import {
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
  type Cell,
  type Column,
  type ColumnDef,
  type Row,
  type SortingState,
  type VisibilityState,
} from '@tanstack/react-table'
import { Card } from '@/components/common/Card'
import { DebouncedSearchInput } from '@/components/common/DeboundedSearch'
import { EmptyState, type IEmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Pagination, type IPagination } from '@/components/common/Pagination'
import { Text } from '@/components/common/Text'
import { ToggleButton } from '@/components/common/ToggleButton'
import { usePagination } from '@/hooks/use-pagination'
import { PaginationProvider } from '@/providers/pagination-provider'
import { cn } from '@/utils/classnames'
import { ColumnSettings } from './ColumnSettings'
import { useTableSettings, type TTableV2View } from './use-table-settings'

export type TTableV2CardSlot = 'title' | 'status' | 'field' | 'actions' | 'hidden'

export interface ITableV2ColumnMeta {
  card?: TTableV2CardSlot
  className?: string
  defaultHidden?: boolean
  label?: string
}

export interface ITableV2<TData extends object> {
  className?: string
  columns: ColumnDef<TData, any>[]
  data: TData[]
  emptyStateProps?: IEmptyState
  enableSearch?: boolean
  enableSorting?: boolean
  filterActions?: ReactNode
  initialSorting?: SortingState
  isLoading?: boolean
  loadingRows?: number
  pagination?: Omit<IPagination, 'position'>
  searchPlaceholder?: string
  tableId: string
}

const columnMeta = (column: { columnDef: { meta?: unknown } }) =>
  (column.columnDef.meta ?? {}) as ITableV2ColumnMeta

const columnLabel = <TData,>(column: Column<TData, unknown>) => {
  const meta = columnMeta(column)
  if (meta.label) return meta.label
  const header = column.columnDef.header
  return typeof header === 'string' && header.trim() ? header : undefined
}

const cardSlot = <TData,>(
  column: Column<TData, unknown>,
  index: number
): TTableV2CardSlot => {
  const meta = columnMeta(column)
  if (meta.card) return meta.card
  if (index === 0) return 'title'
  return columnLabel(column) ? 'field' : 'actions'
}

const isActionsColumn = <TData,>(column: Column<TData, unknown>) => {
  const meta = columnMeta(column)
  if (meta.card) return meta.card === 'actions'
  return !columnLabel(column)
}

const useContainerWidth = () => {
  const ref = useRef<HTMLDivElement>(null)
  const [width, setWidth] = useState<number | null>(null)

  useLayoutEffect(() => {
    const el = ref.current
    if (!el) return
    const update = (next: number) => {
      if (next > 0) setWidth(next)
    }
    update(el.getBoundingClientRect().width)
    if (typeof ResizeObserver === 'undefined') return
    const observer = new ResizeObserver(([entry]) =>
      update(entry.contentRect.width)
    )
    observer.observe(el)
    return () => observer.disconnect()
  }, [])

  return [ref, width] as const
}

const renderCell = <TData,>(cell: Cell<TData, unknown>) =>
  flexRender(cell.column.columnDef.cell, cell.getContext())

const GridCard = <TData,>({ row }: { row: Row<TData> }) => {
  const slots = new Map<TTableV2CardSlot, Cell<TData, unknown>[]>()
  row.getAllCells().forEach((cell, index) => {
    if (!cell.column.getIsVisible()) return
    const slot = cardSlot(cell.column, index)
    slots.set(slot, [...(slots.get(slot) ?? []), cell])
  })

  const title = slots.get('title') ?? []
  const status = slots.get('status') ?? []
  const fields = slots.get('field') ?? []
  const actions = slots.get('actions') ?? []

  return (
    <Card className="@container h-full gap-4 p-4">
      <div className="flex items-start justify-between gap-4">
        <div className="flex min-w-0 flex-col gap-1">
          {title.map((cell) => (
            <div key={cell.id} className="min-w-0">
              {renderCell(cell)}
            </div>
          ))}
        </div>
        {status.length || actions.length ? (
          <div className="flex shrink-0 items-center gap-2">
            {status.map((cell) => (
              <div key={cell.id}>{renderCell(cell)}</div>
            ))}
            {actions.map((cell) => (
              <div key={cell.id}>{renderCell(cell)}</div>
            ))}
          </div>
        ) : null}
      </div>
      {fields.length ? (
        <div className="grid grid-cols-1 gap-4 @xs:grid-cols-2">
          {fields.map((cell) => (
            <LabeledValue
              key={cell.id}
              className="min-w-0"
              label={columnLabel(cell.column) ?? cell.column.id}
            >
              <div className="min-w-0 break-words">{renderCell(cell)}</div>
            </LabeledValue>
          ))}
        </div>
      ) : null}
    </Card>
  )
}

const LoadingCard = ({ fieldLabels }: { fieldLabels: string[] }) => (
  <Card className="@container h-full gap-4 p-4">
    <div className="flex flex-col gap-1">
      <Text loading loadingWidth={18} />
      <Text variant="subtext" loading loadingWidth={12} />
    </div>
    {fieldLabels.length ? (
      <div className="grid grid-cols-1 gap-4 @xs:grid-cols-2">
        {fieldLabels.map((label) => (
          <LabeledValue key={label} label={label} loading loadingWidth={10} />
        ))}
      </div>
    ) : null}
  </Card>
)

const VIEW_OPTIONS = [
  {
    value: 'table' as const,
    label: <Icon variant="TableIcon" size={16} />,
    ariaLabel: 'Table view',
    title: 'Table',
  },
  {
    value: 'grid' as const,
    label: <Icon variant="SquaresFourIcon" size={16} />,
    ariaLabel: 'Grid view',
    title: 'Grid',
  },
]

export function TableV2Base<TData extends object>({
  className,
  columns,
  data,
  emptyStateProps = { emptyMessage: 'No data found' },
  enableSearch = true,
  enableSorting = true,
  filterActions,
  initialSorting = [],
  isLoading = false,
  loadingRows = 5,
  pagination,
  searchPlaceholder,
  tableId,
}: ITableV2<TData>) {
  const { isPaginating, setIsPaginating } = usePagination()
  const [sorting, setSorting] = useState<SortingState>(initialSorting)
  const [containerRef, containerWidth] = useContainerWidth()

  const defaultVisibility = useMemo<VisibilityState>(
    () =>
      Object.fromEntries(
        columns
          .map((column) => {
            const id =
              column.id ??
              (column as { accessorKey?: string }).accessorKey ??
              ''
            return [id, !(column.meta as ITableV2ColumnMeta)?.defaultHidden]
          })
          .filter(([id]) => id)
      ),
    [columns]
  )

  const settings = useTableSettings(tableId, defaultVisibility)

  const table = useReactTable({
    data,
    columns,
    state: { sorting, columnVisibility: settings.columnVisibility },
    onSortingChange: enableSorting ? setSorting : undefined,
    onColumnVisibilityChange: (updater) =>
      settings.setColumnVisibility(
        typeof updater === 'function'
          ? updater(settings.columnVisibility)
          : updater
      ),
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: enableSorting ? getSortedRowModel() : undefined,
  })

  useEffect(() => {
    setIsPaginating(false)
  }, [data])

  const requiredWidth = table.getTotalSize()
  const forcedGrid =
    containerWidth !== null && containerWidth + 1 < requiredWidth
  const view: TTableV2View = forcedGrid ? 'grid' : (settings.view ?? 'table')
  const loading = isLoading || isPaginating
  const rows = table.getRowModel().rows
  const visibleColumns = table.getVisibleLeafColumns()
  const placeholders = Array.from({ length: pagination?.limit || loadingRows })

  const columnOptions = table
    .getAllLeafColumns()
    .map((column) => {
      const label = columnLabel(column)
      if (!label) return null
      return {
        id: column.id,
        label,
        locked: !column.getCanHide(),
        visible: column.getIsVisible(),
      }
    })
    .filter((option) => option !== null)

  const loadingFieldLabels = table
    .getAllLeafColumns()
    .map((column, index) =>
      column.getIsVisible() && cardSlot(column, index) === 'field'
        ? columnLabel(column)
        : undefined
    )
    .filter((label): label is string => Boolean(label))
    .slice(0, 4)

  const empty = !loading && rows.length === 0
  const hasPagination = Boolean(
    pagination && (pagination.hasNext || pagination.offset !== 0)
  )

  return (
    <div
      ref={containerRef}
      data-table-view={view}
      className="flex w-full min-w-0 flex-col gap-4 md:gap-6"
    >
      <div className="flex flex-row flex-wrap items-center justify-between gap-4">
        {enableSearch ? (
          <DebouncedSearchInput
            labelClassName="w-full md:w-fit"
            className="w-full md:w-fit"
            placeholder={searchPlaceholder}
          />
        ) : null}
        <div
          className={cn('flex flex-wrap items-center gap-4', {
            'w-full justify-end': !enableSearch,
            'w-full md:w-fit': enableSearch,
          })}
        >
          {filterActions}
          <div className="ml-auto flex items-center gap-2">
            {forcedGrid ? null : (
              <ToggleButton
                label="Layout"
                options={VIEW_OPTIONS}
                value={view}
                onChange={settings.setView}
              />
            )}
            <ColumnSettings
              id={`${tableId}-column-settings`}
              canReset={settings.hasCustomColumns || Boolean(settings.view)}
              onReset={settings.reset}
              onToggle={(id, visible) =>
                table.getColumn(id)?.toggleVisibility(visible)
              }
              options={columnOptions}
            />
          </div>
        </div>
      </div>

      {view === 'table' ? (
        <div className={cn('overflow-hidden rounded-lg border', className)}>
          <table className="w-full text-sm">
            <thead>
              {table.getHeaderGroups().map((headerGroup) => (
                <tr key={headerGroup.id}>
                  {headerGroup.headers.map((header, idx) => (
                    <th
                      key={header.id}
                      className={cn(
                        'py-3 px-4 text-left bg-cool-grey-100 dark:bg-dark-grey-700',
                        isActionsColumn(header.column) &&
                          'w-px whitespace-nowrap text-right',
                        columnMeta(header.column).className,
                        {
                          'cursor-pointer select-none':
                            header.column.getCanSort(),
                          'rounded-tl-lg': idx === 0,
                          'rounded-tr-lg':
                            idx === headerGroup.headers.length - 1,
                        }
                      )}
                      onClick={
                        header.column.getCanSort()
                          ? header.column.getToggleSortingHandler()
                          : undefined
                      }
                    >
                      <span
                        className={cn('flex gap-1 items-center', {
                          'justify-end': isActionsColumn(header.column),
                        })}
                      >
                        <span className="font-normal font-sans">
                          {flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                        </span>
                        {header.column.getIsSorted() === 'asc' ? (
                          <Icon variant="SortDescendingIcon" />
                        ) : header.column.getIsSorted() === 'desc' ? (
                          <Icon variant="SortAscendingIcon" />
                        ) : null}
                      </span>
                    </th>
                  ))}
                </tr>
              ))}
            </thead>
            <tbody>
              {loading ? (
                placeholders.map((_, i) => (
                  <tr key={`loading-${i}`}>
                    {visibleColumns.map((column) => (
                      <td
                        key={column.id}
                        className={cn(
                          'py-3 px-4 border-t',
                          isActionsColumn(column) &&
                            'w-px whitespace-nowrap text-right',
                          columnMeta(column).className
                        )}
                      >
                        <Text loading loadingWidth={10} />
                      </td>
                    ))}
                  </tr>
                ))
              ) : empty ? (
                <tr>
                  <td
                    colSpan={visibleColumns.length}
                    className="text-center py-8"
                  >
                    <EmptyState variant="table" {...emptyStateProps} />
                  </td>
                </tr>
              ) : (
                rows.map((row) => (
                  <tr key={row.id}>
                    {row.getVisibleCells().map((cell) => (
                      <td
                        key={cell.id}
                        className={cn(
                          'py-3 px-4 border-t',
                          isActionsColumn(cell.column) &&
                            'w-px whitespace-nowrap text-right',
                          columnMeta(cell.column).className
                        )}
                      >
                        {renderCell(cell)}
                      </td>
                    ))}
                  </tr>
                ))
              )}
            </tbody>
          </table>
          {hasPagination ? (
            <div className="flex items-center justify-center w-full border-t py-2">
              <Pagination {...pagination} />
            </div>
          ) : null}
        </div>
      ) : (
        <div className="@container flex flex-col gap-4">
          {empty ? (
            <Card className="items-center py-8">
              <EmptyState variant="table" {...emptyStateProps} />
            </Card>
          ) : (
            <div className="grid grid-cols-1 gap-4 @xl:grid-cols-2 @5xl:grid-cols-3">
              {loading
                ? placeholders.map((_, i) => (
                    <LoadingCard
                      key={`loading-${i}`}
                      fieldLabels={loadingFieldLabels}
                    />
                  ))
                : rows.map((row) => <GridCard key={row.id} row={row} />)}
            </div>
          )}
          {hasPagination ? (
            <div className="flex items-center justify-center w-full">
              <Pagination {...pagination} />
            </div>
          ) : null}
        </div>
      )}
    </div>
  )
}

export const TableV2 = <T extends object>(props: ITableV2<T>) => (
  <PaginationProvider>
    <TableV2Base<T> {...props} />
  </PaginationProvider>
)
