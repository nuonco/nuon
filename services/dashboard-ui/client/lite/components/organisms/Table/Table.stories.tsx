import type { ColumnDef } from '@tanstack/react-table'
import { useLayoutEffect, useState } from 'react'
import { useLocation } from 'react-router'
import { useListQueryState } from '../../../hooks/use-list-query-state'
import { useTableView, type TTableView } from '../../../hooks/use-table-view'
import { commaSetQueryParameter } from '../../../lib/list-query'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { Brand } from '../../atoms/Brand'
import { Button } from '../../atoms/Button'
import { Card } from '../../atoms/Card'
import { Code } from '../../atoms/Code'
import { Icon } from '../../atoms/Icon'
import { Link } from '../../atoms/Link'
import { Status } from '../../atoms/Status'
import { Text } from '../../atoms/Text'
import {
  FilterDropdown,
  type IFilterMenuOption,
} from '../../molecules/FilterMenu'
import { ID } from '../../molecules/ID'
import { ListSearch } from '../../molecules/ListSearch'
import { Pagination } from '../../molecules/Pagination'
import { Table } from './Table'

export default {
  title: 'lite/organisms/Table',
}

type TInstallRow = {
  id: string
  name: string
  app: string
  status: string
  platform: 'AWS' | 'Azure' | 'GCP'
  region: string
  activity: string
}

const INSTALLS: TInstallRow[] = [
  {
    id: 'ins_acme_production',
    name: 'Production',
    app: 'Payments API',
    status: 'active',
    platform: 'AWS',
    region: 'us-west-2',
    activity: '2 minutes ago',
  },
  {
    id: 'ins_acme_staging',
    name: 'Staging',
    app: 'Payments API',
    status: 'deploying',
    platform: 'AWS',
    region: 'us-east-1',
    activity: '8 minutes ago',
  },
  {
    id: 'ins_acme_preview',
    name: 'Preview',
    app: 'Dashboard',
    status: 'failed',
    platform: 'GCP',
    region: 'us-central1',
    activity: '1 hour ago',
  },
  {
    id: 'ins_acme_eu',
    name: 'EU production',
    app: 'Dashboard',
    status: 'active',
    platform: 'Azure',
    region: 'westeurope',
    activity: 'Yesterday',
  },
]

const InstallActions = ({ name }: { name: string }) => (
  <Button
    size="sm"
    variant="ghost"
    iconOnly
    aria-label={`${name} actions`}
    tooltip="Install actions"
  >
    <Icon variant="DotsThreeIcon" size={16} aria-hidden />
  </Button>
)

const columns: ColumnDef<TInstallRow>[] = [
  {
    accessorKey: 'name',
    header: 'Install',
    size: 220,
    cell: ({ row }) => (
      <span className="flex min-w-0 flex-col gap-0.5">
        <Link
          href={`/org-example/installs/${row.original.id}`}
          variant="body"
          className="w-fit"
        >
          {row.original.name}
        </Link>
        <ID value={row.original.id} label="Copy install ID" truncate />
      </span>
    ),
  },
  { accessorKey: 'app', header: 'App', size: 180 },
  {
    accessorKey: 'status',
    header: 'Status',
    size: 150,
    cell: ({ row }) => <Status status={row.original.status} />,
  },
  {
    accessorKey: 'platform',
    header: 'Platform',
    size: 120,
    cell: ({ row }) => (
      <span className="flex items-center gap-2">
        <Brand variant={row.original.platform} size={18} />
        <Text>{row.original.platform}</Text>
      </span>
    ),
  },
  { accessorKey: 'region', header: 'Region', size: 160 },
  { accessorKey: 'activity', header: 'Activity', size: 160 },
  {
    id: 'actions',
    header: '',
    size: 64,
    cell: ({ row }) => <InstallActions name={row.original.name} />,
  },
]

const InstallCard = ({ install }: { install: TInstallRow }) => (
  <Card className="flex h-full flex-col gap-4">
    <div className="flex items-start justify-between gap-3">
      <span className="flex min-w-0 flex-col gap-1">
        <Link
          href={`/org-example/installs/${install.id}`}
          variant="heading"
          className="w-fit"
        >
          {install.name}
        </Link>
        <ID value={install.id} label="Copy install ID" />
      </span>
      <InstallActions name={install.name} />
    </div>
    <Status status={install.status} className="self-start" />
    <dl className="grid grid-cols-2 gap-x-4 gap-y-3">
      {[
        ['App', install.app],
        ['Region', install.region],
        [
          'Platform',
          <span key="platform" className="flex items-center gap-2">
            <Brand variant={install.platform} size={16} />
            {install.platform}
          </span>,
        ],
        ['Activity', install.activity],
      ].map(([label, value]) => (
        <div key={label as string} className="min-w-0">
          <Text as="dt" variant="label" color="tertiary">
            {label}
          </Text>
          <Text as="dd" variant="body" color="secondary" className="mt-0.5">
            {value}
          </Text>
        </div>
      ))}
    </dl>
  </Card>
)

const InstallsTable = ({
  data = INSTALLS,
  loading,
  toolbar,
}: {
  data?: TInstallRow[]
  loading?: boolean
  toolbar?: React.ReactNode
}) => (
  <Table
    data={data}
    columns={columns}
    getRowId={(row) => row.id}
    loading={loading}
    emptyState="No installs yet"
    toolbar={toolbar}
    renderCard={({ row }) => <InstallCard install={row.original} />}
  />
)

const PreferredView = ({
  view,
  children,
}: {
  view: TTableView
  children: React.ReactNode
}) => {
  const { setView } = useTableView()

  useLayoutEffect(() => setView(view), [setView, view])
  return children
}

export const Overview = () => (
  <ComponentDocs
    name="Table"
    tier="organism"
    summary="One TanStack row model rendered as a semantic table or resource-defined card grid."
    use={[
      'Define columns and renderCard together in each resource table.',
      'Pass ListSearch and resource filters through toolbar; keep Pagination outside Table.',
      'Set column sizes so Table can switch before horizontal overflow.',
    ]}
    avoid={[
      'Do not infer card hierarchy from column headers.',
      'Do not add client-side sorting to server-paginated rows.',
      'Do not make entire rows or cards clickable.',
    ]}
    rules={[
      'Cards are forced when the declared table width no longer fits.',
      'The view preference lasts for the browser session and returns when space allows.',
      'Forced card mode hides the view toggle without changing the preference.',
      'Toolbar controls stay left of the view toggle and wrap when space is constrained.',
    ]}
    props={[
      {
        name: 'columns',
        type: 'ColumnDef<T>[]',
        description: 'Resource-owned TanStack column and cell definitions.',
      },
      {
        name: 'data',
        type: 'T[]',
        description: 'Resolved rows for the current API page.',
      },
      {
        name: 'renderCard',
        type: '({ row: Row<T> }) => ReactNode',
        description: 'Required resource-owned card presentation.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Preserves the active presentation while data is loading.',
      },
      {
        name: 'loadingRows',
        type: 'number',
        default: '5',
        description: 'Skeleton row or card count rendered while loading.',
      },
      {
        name: 'emptyState',
        type: 'ReactNode',
        default: "'No results'",
        description: 'Content shown when the resolved page has no rows.',
      },
      {
        name: 'toolbar',
        type: 'ReactNode',
        description: 'Search and filter controls aligned left of the view toggle.',
      },
    ]}
  />
)

export const TableView = () => (
  <PreferredView view="table">
    <div className="p-8">
      <InstallsTable />
    </div>
  </PreferredView>
)

export const CardGridView = () => (
  <PreferredView view="cards">
    <div className="p-8">
      <InstallsTable />
    </div>
  </PreferredView>
)

export const InteractiveViewSwitching = () => (
  <PreferredView view="table">
    <div className="p-8">
      <InstallsTable />
    </div>
  </PreferredView>
)

export const NarrowContainer = () => (
  <PreferredView view="table">
    <div className="max-w-lg p-4">
      <InstallsTable />
    </div>
  </PreferredView>
)

export const ResponsiveContainer = () => {
  const [wide, setWide] = useState(false)

  return (
    <PreferredView view="table">
      <div className="flex flex-col gap-4 p-8">
        <Button
          className="self-start"
          onClick={() => setWide((value) => !value)}
        >
          {wide ? 'Make container narrow' : 'Make container wide'}
        </Button>
        <div className={wide ? 'w-full' : 'max-w-lg'}>
          <InstallsTable />
        </div>
      </div>
    </PreferredView>
  )
}

export const Loading = () => (
  <PreferredView view="table">
    <div className="p-8">
      <InstallsTable loading />
    </div>
  </PreferredView>
)

export const LoadingCards = () => (
  <PreferredView view="cards">
    <div className="p-8">
      <InstallsTable loading />
    </div>
  </PreferredView>
)

export const Empty = () => (
  <div className="p-8">
    <InstallsTable data={[]} />
  </div>
)

const FILTER_OPTIONS = [
  { value: 'AWS', label: 'AWS' },
  { value: 'Azure', label: 'Azure' },
  { value: 'GCP', label: 'GCP' },
] as const satisfies readonly IFilterMenuOption<string>[]
const FILTER_VALUES = FILTER_OPTIONS.map(({ value }) => value)
const FILTERS = {
  platforms: commaSetQueryParameter<(typeof FILTER_VALUES)[number]>(
    'platforms',
    { defaultValue: FILTER_VALUES }
  ),
}

export const UrlBackedListQuery = () => {
  const location = useLocation()
  const list = useListQueryState({ pageSize: 20, filters: FILTERS })
  const selected = list.filters.platforms
  const setSelected = (next: Set<(typeof FILTER_VALUES)[number]>) =>
    list.setFilter('platforms', next)
  const mappedRequest = {
    q: list.search || undefined,
    offset: list.offset,
    limit: list.pageSize,
    platforms: [...selected].sort().join(','),
  }

  return (
    <div className="flex flex-col gap-4 p-8">
      <InstallsTable
        toolbar={
          <>
            <ListSearch
              value={list.search}
              onValueChange={list.setSearch}
              placeholder="Search installs"
              aria-label="Search installs"
              className="w-full max-w-80"
            />
            <FilterDropdown
              label="Platform"
              options={FILTER_OPTIONS}
              selected={selected}
              onToggle={(value) => {
                const next = new Set(selected)
                if (next.has(value)) next.delete(value)
                else next.add(value)
                setSelected(next)
              }}
              onIsolate={(value) =>
                setSelected(
                  selected.size === 1 && selected.has(value)
                    ? new Set(FILTER_VALUES)
                    : new Set([value])
                )
              }
              onReset={() => setSelected(new Set(FILTER_VALUES))}
            />
          </>
        }
      />
      <Pagination
        offset={list.offset}
        pageSize={list.pageSize}
        hasNext={list.offset < 80}
        onOffsetChange={list.setOffset}
      />
      <Card className="flex flex-col gap-2">
        <Text variant="label">Current URL and mapped request</Text>
        <Code className="block whitespace-pre-wrap">
          {location.search || '(empty)'}
          {'\n'}
          {JSON.stringify(mappedRequest, null, 2)}
        </Code>
      </Card>
    </div>
  )
}
