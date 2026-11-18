export default {
  title: 'Playground / Table v2',
}

import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { TableV2 } from './TableV2'

type TMockInstall = {
  app: string
  branch: string
  id: string
  labels: Record<string, string>
  name: string
  platform: 'aws' | 'azure' | 'gcp'
  region: string
  status: 'active' | 'error' | 'provisioning' | 'deprovisioned'
  updatedAt: string
}

const APPS = ['payments', 'analytics', 'search']
const PLATFORMS: TMockInstall['platform'][] = ['aws', 'azure', 'gcp']
const REGIONS: Record<TMockInstall['platform'], string[]> = {
  aws: ['us-east-1', 'eu-west-1'],
  azure: ['eastus', 'westeurope'],
  gcp: ['us-central1', 'europe-west1'],
}
const STATUSES: TMockInstall['status'][] = [
  'active',
  'active',
  'provisioning',
  'error',
  'deprovisioned',
]

const mockInstalls: TMockInstall[] = Array.from({ length: 24 }, (_, i) => {
  const platform = PLATFORMS[i % PLATFORMS.length]
  return {
    app: APPS[i % APPS.length],
    branch: i % 4 === 0 ? 'release' : 'main',
    id: `inst${String(i + 1).padStart(3, '0')}example`,
    labels:
      i % 3 === 0
        ? { tier: i % 2 ? 'enterprise' : 'standard', env: 'prod' }
        : {},
    name: `acme-${APPS[i % APPS.length]}-${i + 1}`,
    platform,
    region: REGIONS[platform][i % 2],
    status: STATUSES[i % STATUSES.length],
    updatedAt: new Date(Date.UTC(2026, 10, 18 - i, 12)).toISOString(),
  }
})

const columns: ColumnDef<TMockInstall>[] = [
  {
    accessorKey: 'name',
    header: 'Install name',
    enableHiding: false,
    size: 220,
    cell: (info) => (
      <span className="flex min-w-0 flex-col">
        <Text variant="body" className="truncate">
          <Link href={`/installs/${info.row.original.id}`} variant="inline">
            {info.getValue() as string}
          </Link>
        </Text>
        <ID>{info.row.original.id}</ID>
      </span>
    ),
  },
  {
    accessorKey: 'status',
    header: 'Status',
    size: 140,
    meta: { card: 'status' },
    cell: (info) => <Status status={info.getValue() as string} />,
  },
  {
    accessorKey: 'app',
    header: 'App',
    size: 140,
    cell: (info) => (
      <Link href={`/apps/${info.getValue()}`}>{info.getValue() as string}</Link>
    ),
  },
  {
    accessorKey: 'platform',
    header: 'Platform',
    size: 110,
    cell: (info) => (
      <Text variant="subtext" className="uppercase">
        {info.getValue() as string}
      </Text>
    ),
  },
  {
    accessorKey: 'region',
    header: 'Region',
    size: 130,
    cell: (info) => (
      <Text variant="subtext" family="mono">
        {info.getValue() as string}
      </Text>
    ),
  },
  {
    accessorKey: 'branch',
    header: 'Branch',
    size: 120,
    enableSorting: false,
    cell: (info) => (
      <span className="flex items-center gap-1.5">
        <Icon variant="GitBranchIcon" size={14} />
        <Text variant="subtext">{info.getValue() as string}</Text>
      </span>
    ),
  },
  {
    accessorKey: 'labels',
    header: 'Labels',
    size: 200,
    enableSorting: false,
    meta: { defaultHidden: true },
    cell: (info) => {
      const labels = info.getValue() as TMockInstall['labels']
      const keys = Object.keys(labels).sort()
      if (!keys.length) return <Icon variant="MinusIcon" />
      return (
        <span className="flex flex-wrap gap-1">
          {keys.map((key) => (
            <Badge key={key} size="sm" variant="code">
              {key}: {labels[key]}
            </Badge>
          ))}
        </span>
      )
    },
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
      <Button aria-label="Manage install" size="sm" variant="ghost">
        <Icon variant="DotsThreeVerticalIcon" />
      </Button>
    ),
  },
]

const PAGE_SIZE = 10

const useMockPage = () => {
  const [searchParams] = useSearchParams()
  const query = searchParams.get('q')?.toLowerCase() ?? ''
  const offset = Number(searchParams.get('offset') ?? 0)

  return useMemo(() => {
    const matches = mockInstalls.filter(
      (install) =>
        !query ||
        install.name.includes(query) ||
        install.id.includes(query) ||
        install.app.includes(query)
    )
    return {
      data: matches.slice(offset, offset + PAGE_SIZE),
      pagination: {
        limit: PAGE_SIZE,
        offset,
        hasNext: offset + PAGE_SIZE < matches.length,
      },
    }
  }, [query, offset])
}

const Note = ({ children }: { children: string }) => (
  <Text variant="subtext" theme="neutral">
    {children}
  </Text>
)

export const Default = () => {
  const { data, pagination } = useMockPage()
  return (
    <div className="flex flex-col gap-4 p-6">
      <Note>
        Pick columns from the settings button and switch between table and
        grid. Both choices are saved to local storage under
        nuon-table:playground-installs and come back on reload. Narrow the
        window and the table switches to the grid once the visible columns no
        longer fit. Hiding columns lets the table fit in less space.
      </Note>
      <TableV2
        tableId="playground-installs"
        columns={columns}
        data={data}
        pagination={pagination}
        searchPlaceholder="Search by name, app or ID..."
        emptyStateProps={{
          emptyTitle: 'No installs found',
          emptyMessage: 'No installs match your search.',
        }}
      />
    </div>
  )
}

export const NarrowContainer = () => {
  const { data, pagination } = useMockPage()
  return (
    <div className="flex flex-col gap-4 p-6">
      <Note>
        The same table in a 28rem container. The visible columns do not fit, so
        it renders as a grid and hides the layout toggle. Column choices still
        apply.
      </Note>
      <div className="max-w-md">
        <TableV2
          tableId="playground-installs"
          columns={columns}
          data={data}
          pagination={pagination}
          searchPlaceholder="Search by name, app or ID..."
        />
      </div>
    </div>
  )
}

export const Loading = () => (
  <div className="flex flex-col gap-4 p-6">
    <Note>
      Loading in whichever layout is active. Uses its own table ID so it does
      not share settings with the other stories.
    </Note>
    <TableV2
      tableId="playground-installs-loading"
      columns={columns}
      data={[]}
      isLoading
      loadingRows={6}
    />
  </div>
)

export const Empty = () => (
  <div className="flex flex-col gap-4 p-6">
    <TableV2
      tableId="playground-installs-empty"
      columns={columns}
      data={[]}
      emptyStateProps={{
        emptyTitle: 'No installs created',
        emptyMessage:
          'An install is an instance of an application running in a customer cloud account.',
      }}
    />
  </div>
)
