import { useMemo, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import type { ColumnDef } from '@tanstack/react-table'
import type { PluggableList } from 'unified'
import { cn } from '@/utils/classnames'
import { Table } from '../Table'
import { SearchInput } from '../SearchInput'
import { Link } from '../Link'
import type { ExtractedTable, TableAlign, TableCell } from './markdown-table'

type Row = { __i: number }

const INLINE_REMARK: PluggableList = [remarkGfm]

const INLINE_COMPONENTS = {
  p: ({ children }: any) => <>{children}</>,
  a: ({ href, children, ...props }: any) => {
    const isExternal = href && !href.startsWith('#') && !href.startsWith('/')
    return (
      <Link href={href} isExternal={isExternal} {...props}>
        {children}
      </Link>
    )
  },
  code: ({ className, children, ...props }: any) => (
    <code
      className={cn(
        'bg-code text-sm text-blue-800 dark:text-blue-500 font-mono px-1 py-0.5 rounded',
        className
      )}
      {...props}
    >
      {children}
    </code>
  ),
}

function CellContent({ cell }: { cell: TableCell }) {
  if (cell.content !== undefined) return <>{cell.content}</>
  if (!cell.markdown) return null
  return (
    <ReactMarkdown remarkPlugins={INLINE_REMARK} components={INLINE_COMPONENTS}>
      {cell.markdown}
    </ReactMarkdown>
  )
}

const EMPTY_CELL: TableCell = { text: '' }

function alignClass(align: TableAlign) {
  if (align === 'center') return 'text-center'
  if (align === 'right') return 'text-right'
  return undefined
}

function searchPlaceholder(headers: TableCell[], indexes: number[] | null): string {
  if (!indexes || indexes.length === 0 || indexes.length === headers.length) {
    return 'Search table…'
  }
  const names = indexes.map((i) => headers[i]?.text).filter(Boolean)
  return names.length ? `Search by ${names.join(', ')}…` : 'Search table…'
}

export function MarkdownTable({ headers, rows, align, search }: ExtractedTable) {
  const [query, setQuery] = useState('')

  const searchColIndexes = useMemo(() => {
    if (!search) return null
    if (!search.columns) return headers.map((_, i) => i)
    const wanted = new Set(search.columns)
    return headers
      .map((h, i) => (wanted.has(h.text.toLowerCase()) ? i : -1))
      .filter((i) => i >= 0)
  }, [search, headers])

  const columns = useMemo<ColumnDef<Row, any>[]>(
    () =>
      headers.map((h, i) => ({
        id: `col-${i}`,
        header: h.text || `Column ${i + 1}`,
        accessorFn: (row: Row) => rows[row.__i]?.[i]?.text ?? '',
        cell: (ctx) => {
          const cls = alignClass(align[i])
          return (
            <div className={cls}>
              <CellContent cell={rows[ctx.row.original.__i]?.[i] ?? EMPTY_CELL} />
            </div>
          )
        },
      })),
    [headers, rows, align]
  )

  const data = useMemo<Row[]>(() => rows.map((_, i) => ({ __i: i })), [rows])

  const filtered = useMemo(() => {
    if (!search || !searchColIndexes || !query.trim()) return data
    const q = query.trim().toLowerCase()
    return data.filter((row) =>
      searchColIndexes.some((ci) =>
        (rows[row.__i]?.[ci]?.text ?? '').toLowerCase().includes(q)
      )
    )
  }, [data, rows, query, search, searchColIndexes])

  return (
    <div className="not-prose my-4 flex flex-col gap-3">
      {search ? (
        <SearchInput
          value={query}
          onChange={setQuery}
          placeholder={search.placeholder || searchPlaceholder(headers, searchColIndexes)}
          className="w-full md:w-80"
          labelClassName="w-full md:w-fit"
        />
      ) : null}
      <Table
        columns={columns}
        data={filtered}
        enableSearch={false}
        enableSorting
        emptyStateProps={{ emptyMessage: query ? 'No matching rows' : 'No data' }}
      />
    </div>
  )
}
