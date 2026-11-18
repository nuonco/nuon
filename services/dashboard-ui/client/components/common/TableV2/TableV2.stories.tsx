export default {
  title: 'UI / Table v2',
}

import { useMemo, useState, type ReactNode } from 'react'
import { useSearchParams } from 'react-router'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { CheckboxFilterDropdown } from '@/components/common/CheckboxFilterDropdown'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { ColumnSettings, TableV2, type IColumnSettingsOption } from './index'

type TComponent = {
  id: string
  name: string
  owner: string
  status: string
  type: string
  updatedAt: string
  version: string
}

const NAMES = [
  'api',
  'worker',
  'database',
  'cache',
  'ingress',
  'scheduler',
  'search-index',
  'object-storage',
  'queue',
  'metrics',
  'web',
  'migrations',
]
const TYPES = [
  'helm chart',
  'terraform module',
  'docker build',
  'kubernetes manifest',
]
const STATUSES = ['active', 'active', 'building', 'error', 'queued']

const mockComponents: TComponent[] = NAMES.map((name, i) => ({
  id: `cmp${String(i + 1).padStart(3, '0')}example`,
  name,
  owner: i % 3 === 0 ? 'platform' : i % 3 === 1 ? 'data' : 'web',
  status: STATUSES[i % STATUSES.length],
  type: TYPES[i % TYPES.length],
  updatedAt: new Date(Date.UTC(2026, 10, 18, 12 - i * 5)).toISOString(),
  version: `v1.${i}.${(i * 7) % 10}`,
}))

const columns: ColumnDef<TComponent>[] = [
  {
    accessorKey: 'name',
    header: 'Component',
    enableHiding: false,
    size: 200,
    cell: (info) => (
      <Text variant="body">
        <Link href={`/components/${info.row.original.id}`} variant="inline">
          {info.getValue() as string}
        </Link>
      </Text>
    ),
  },
  {
    accessorKey: 'status',
    header: 'Status',
    size: 130,
    meta: { card: 'status' },
    cell: (info) => <Status status={info.getValue() as string} />,
  },
  {
    accessorKey: 'type',
    header: 'Type',
    size: 170,
    cell: (info) => <Badge theme="info">{info.getValue() as string}</Badge>,
  },
  {
    accessorKey: 'version',
    header: 'Version',
    size: 110,
    cell: (info) => <Text family="mono">{info.getValue() as string}</Text>,
  },
  {
    accessorKey: 'owner',
    header: 'Owner',
    size: 120,
    meta: { defaultHidden: true },
    cell: (info) => <Text>{info.getValue() as string}</Text>,
  },
  {
    accessorKey: 'updatedAt',
    header: 'Updated',
    size: 130,
    cell: (info) => (
      <Time
        time={info.getValue() as string}
        variant="subtext"
        format="relative"
      />
    ),
  },
  {
    id: 'action',
    header: '',
    size: 56,
    enableSorting: false,
    enableHiding: false,
    cell: () => (
      <Button aria-label="Component actions" size="sm" variant="ghost">
        <Icon variant="DotsThreeVerticalIcon" />
      </Button>
    ),
  },
]

const gridColumns: ColumnDef<TComponent>[] = columns.map((column) =>
  (column as { accessorKey?: string }).accessorKey === 'version'
    ? { ...column, meta: { ...column.meta, card: 'hidden' } }
    : column
)

const PAGE_SIZE = 5

const usePagedComponents = (rows: TComponent[]) => {
  const [searchParams] = useSearchParams()
  const query = searchParams.get('q')?.toLowerCase() ?? ''
  const offset = Number(searchParams.get('offset') ?? 0)
  return useMemo(() => {
    const matches = rows.filter(
      (row) => !query || row.name.includes(query) || row.id.includes(query)
    )
    return {
      data: matches.slice(offset, offset + PAGE_SIZE),
      pagination: {
        limit: PAGE_SIZE,
        offset,
        hasNext: offset + PAGE_SIZE < matches.length,
      },
    }
  }, [rows, query, offset])
}

const Frame = ({
  children,
  description,
}: {
  children: ReactNode
  description: string
}) => (
  <div className="flex flex-col gap-4">
    <Text variant="subtext" theme="neutral">
      {description}
    </Text>
    {children}
  </div>
)

export const Default = () => {
  const { data, pagination } = usePagedComponents(mockComponents)
  return (
    <Frame description="Search, sort, paginate, switch between table and grid, and pick columns. Layout and column choices are saved per tableId in local storage.">
      <TableV2
        tableId="ui-table-v2-default"
        columns={columns}
        data={data}
        pagination={pagination}
        searchPlaceholder="Search by name or ID..."
      />
    </Frame>
  )
}

export const ColumnVisibility = () => (
  <Frame description="Component is locked with enableHiding: false. Owner starts hidden with meta.defaultHidden. The actions column has no header, so it is left out of the picker and always shown.">
    <ColumnSettingsMenu />
    <TableV2
      tableId="ui-table-v2-column-visibility"
      columns={columns}
      data={mockComponents.slice(0, 4)}
      enableSearch={false}
    />
  </Frame>
)

export const GridCards = () => (
  <Frame description="In a narrow container the table switches to the grid and hides the layout toggle. The first column is the card title, meta.card: 'status' and header-less columns sit in the top right, and every other visible column becomes a labelled value. meta.card: 'hidden' keeps a column in the table only.">
    <div className="max-w-xl">
      <TableV2
        tableId="ui-table-v2-grid"
        columns={gridColumns}
        data={mockComponents.slice(0, 4)}
        enableSearch={false}
      />
    </div>
  </Frame>
)

export const WithFilterActions = () => {
  const [types, setTypes] = useState<Set<string>>(new Set())
  const rows = types.size
    ? mockComponents.filter((row) => types.has(row.type))
    : mockComponents
  const { data, pagination } = usePagedComponents(rows)
  return (
    <Frame description="filterActions sit between the search box and the layout controls.">
      <TableV2
        tableId="ui-table-v2-filters"
        columns={columns}
        data={data}
        pagination={pagination}
        searchPlaceholder="Search by name or ID..."
        filterActions={
          <CheckboxFilterDropdown
            id="ui-table-v2-type-filter"
            label="Type"
            options={TYPES.map((type) => ({ label: type, value: type }))}
            selected={types}
            onChange={setTypes}
          />
        }
      />
    </Frame>
  )
}

export const Loading = () => (
  <Frame description="isLoading draws the active layout empty: text placeholders in table rows, or placeholder cards in the grid.">
    <TableV2
      tableId="ui-table-v2-default"
      columns={columns}
      data={[]}
      isLoading
      loadingRows={4}
      enableSearch={false}
    />
  </Frame>
)

export const Empty = () => (
  <Frame description="emptyStateProps are passed to EmptyState in both layouts.">
    <TableV2
      tableId="ui-table-v2-default"
      columns={columns}
      data={[]}
      enableSearch={false}
      emptyStateProps={{
        emptyTitle: 'No components yet',
        emptyMessage: 'Components are the pieces of an app that get deployed.',
      }}
    />
  </Frame>
)

const ColumnSettingsMenu = () => {
  const [options, setOptions] = useState<IColumnSettingsOption[]>([
    { id: 'name', label: 'Component', locked: true, visible: true },
    { id: 'status', label: 'Status', locked: false, visible: true },
    { id: 'type', label: 'Type', locked: false, visible: true },
    { id: 'owner', label: 'Owner', locked: false, visible: false },
  ])
  const isCustomized = options.some(
    (option) => option.id === 'owner' ? option.visible : !option.visible
  )
  return (
    <div className="flex items-center gap-3">
      <Text variant="subtext" theme="neutral">
        The settings menu on its own:
      </Text>
      <ColumnSettings
        id="ui-table-v2-settings-menu"
        canReset={isCustomized}
        onReset={() =>
          setOptions((current) =>
            current.map((option) => ({
              ...option,
              visible: option.id !== 'owner',
            }))
          )
        }
        onToggle={(id, visible) =>
          setOptions((current) =>
            current.map((option) =>
              option.id === id ? { ...option, visible } : option
            )
          )
        }
        options={options}
      />
    </div>
  )
}
