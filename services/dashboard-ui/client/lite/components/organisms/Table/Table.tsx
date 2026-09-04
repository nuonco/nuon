import {
  flexRender,
  getCoreRowModel,
  useReactTable,
  type ColumnDef,
  type Row,
  type TableOptions,
} from '@tanstack/react-table'
import {
  useLayoutEffect,
  useRef,
  useState,
  type HTMLAttributes,
  type ReactNode,
} from 'react'
import { cn } from '@/utils/classnames'
import { useTableView } from '../../../hooks/use-table-view'
import { Card } from '../../atoms/Card'
import { Text } from '../../atoms/Text'
import { TableViewToggle } from '../../molecules/TableViewToggle'

export interface ITable<TData extends object>
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  data: TData[]
  columns: ColumnDef<TData, any>[]
  renderCard: ({ row }: { row: Row<TData> }) => ReactNode
  getRowId?: TableOptions<TData>['getRowId']
  loading?: boolean
  loadingLabel?: string
  loadingRows?: number
  emptyState?: ReactNode
  toolbar?: ReactNode
}

const CELL_WIDTH_RATIOS = [0.9, 0.6, 0.75, 0.5]

const cellLoadingWidth = (
  columnWidth: number,
  rowIndex: number,
  columnIndex: number
) => {
  const characters = Math.max(4, Math.round((columnWidth - 32) / 8))
  const ratio =
    CELL_WIDTH_RATIOS[(rowIndex + columnIndex) % CELL_WIDTH_RATIOS.length]
  return Math.max(3, Math.round(characters * ratio))
}

export const Table = <TData extends object>({
  data,
  columns,
  renderCard,
  getRowId,
  loading = false,
  loadingLabel = 'Loading results',
  loadingRows = 5,
  emptyState = 'No results',
  toolbar,
  className,
  ...props
}: ITable<TData>) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const [containerWidth, setContainerWidth] = useState<number | null>(null)
  const { view: preferredView, setView } = useTableView()
  const table = useReactTable({
    data,
    columns,
    getRowId,
    getCoreRowModel: getCoreRowModel(),
  })
  const requiredWidth = table.getTotalSize()
  const forcedCards =
    containerWidth !== null && containerWidth + 1 < requiredWidth
  const view = forcedCards ? 'cards' : preferredView
  const rows = table.getRowModel().rows

  useLayoutEffect(() => {
    const container = containerRef.current
    if (!container) return

    const updateWidth = (width: number) => {
      if (width > 0) setContainerWidth(width)
    }

    updateWidth(container.getBoundingClientRect().width)
    if (typeof ResizeObserver === 'undefined') return

    const observer = new ResizeObserver(([entry]) =>
      updateWidth(entry.contentRect.width)
    )
    observer.observe(container)
    return () => observer.disconnect()
  }, [])

  const leafColumns = table.getAllLeafColumns()
  const fieldLabels = leafColumns
    .map(({ columnDef }) => columnDef.header)
    .filter(
      (header): header is string =>
        typeof header === 'string' && header.trim().length > 0
    )
    .slice(1, 5)
  const placeholderRows = Array.from({ length: Math.max(1, loadingRows) })

  const emptyCollection =
    !loading && rows.length === 0 ? (
      <div className="flex min-h-24 items-center justify-center px-4 py-8 text-center">
        {typeof emptyState === 'string' ? (
          <Text color="secondary">{emptyState}</Text>
        ) : (
          emptyState
        )}
      </div>
    ) : null

  return (
    <div
      ref={containerRef}
      data-table-view={view}
      data-table-forced-cards={forcedCards || undefined}
      aria-busy={loading || undefined}
      className={cn('@container flex min-w-0 flex-col gap-3', className)}
      {...props}
    >
      {loading ? (
        <span role="status" aria-label={loadingLabel} className="sr-only">
          {loadingLabel}
        </span>
      ) : null}

      {toolbar || !forcedCards ? (
        <div data-table-toolbar className="flex flex-wrap items-center gap-3">
          {toolbar ? (
            <div className="flex min-w-0 flex-1 flex-wrap items-center gap-2">
              {toolbar}
            </div>
          ) : null}
          {!forcedCards ? (
            <TableViewToggle
              value={preferredView}
              onValueChange={setView}
              className="ml-auto"
            />
          ) : null}
        </div>
      ) : null}

      {view === 'table' ? (
        <div className="overflow-hidden rounded-xl border border-card-border bg-card-bg">
          <table className="w-full table-fixed border-collapse text-left">
            <colgroup>
              {table.getAllLeafColumns().map((column) => (
                <col key={column.id} style={{ width: column.getSize() }} />
              ))}
            </colgroup>
            <thead className="bg-field-bg">
              {table.getHeaderGroups().map((headerGroup) => (
                <tr key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <th
                      key={header.id}
                      scope="col"
                      className="border-b border-divider px-4 py-3 text-label font-medium text-secondary"
                    >
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                    </th>
                  ))}
                </tr>
              ))}
            </thead>
            <tbody>
              {loading ? (
                placeholderRows.map((_, rowIndex) => (
                  <tr
                    key={`loading-${rowIndex}`}
                    data-table-loading-row
                    className="border-b border-divider last:border-b-0"
                  >
                    {leafColumns.map((column, columnIndex) => (
                      <td
                        key={column.id}
                        className="min-w-0 px-4 py-3 align-middle"
                      >
                        <Text
                          loading
                          loadingWidth={cellLoadingWidth(
                            column.getSize(),
                            rowIndex,
                            columnIndex
                          )}
                        />
                      </td>
                    ))}
                  </tr>
                ))
              ) : emptyCollection ? (
                <tr>
                  <td colSpan={leafColumns.length}>{emptyCollection}</td>
                </tr>
              ) : (
                rows.map((row) => (
                  <tr
                    key={row.id}
                    data-row-id={row.id}
                    className="border-b border-divider last:border-b-0"
                  >
                    {row.getVisibleCells().map((cell) => (
                      <td
                        key={cell.id}
                        className="min-w-0 px-4 py-3 align-middle text-body text-primary"
                      >
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
        </div>
      ) : loading ? (
        <div className="grid grid-cols-1 gap-3 @xl:grid-cols-2 @4xl:grid-cols-3">
          {placeholderRows.map((_, cardIndex) => (
            <Card
              key={`loading-${cardIndex}`}
              data-table-loading-card
              className="flex flex-col gap-4"
            >
              <div className="flex flex-col gap-1">
                <Text variant="heading" loading loadingWidth={14} />
                <Text variant="caption" loading loadingWidth={20} />
              </div>
              {fieldLabels.length ? (
                <div className="grid grid-cols-2 gap-3">
                  {fieldLabels.map((label, fieldIndex) => (
                    <div key={label} className="flex flex-col gap-1">
                      <Text variant="label" color="secondary">
                        {label}
                      </Text>
                      <Text
                        loading
                        loadingWidth={cellLoadingWidth(
                          160,
                          cardIndex,
                          fieldIndex
                        )}
                      />
                    </div>
                  ))}
                </div>
              ) : null}
            </Card>
          ))}
        </div>
      ) : emptyCollection ? (
        <Card>{emptyCollection}</Card>
      ) : (
        <div className="grid grid-cols-1 gap-3 @xl:grid-cols-2 @4xl:grid-cols-3">
          {rows.map((row) => (
            <div key={row.id} data-table-card data-row-id={row.id}>
              {renderCard({ row })}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
