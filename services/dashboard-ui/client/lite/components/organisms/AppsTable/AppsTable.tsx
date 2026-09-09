import type { ColumnDef } from '@tanstack/react-table'
import type { TApp } from '@/types/ctl-api.types'
import { Brand, type TBrandVariant } from '../../atoms/Brand'
import { Card } from '../../atoms/Card'
import { Link } from '../../atoms/Link'
import { Status } from '../../atoms/Status'
import { Text } from '../../atoms/Text'
import { ID } from '../../molecules/ID'
import { ListSearch } from '../../molecules/ListSearch'
import { Pagination } from '../../molecules/Pagination'
import { Time } from '../../molecules/Time'
import { Table } from '../Table/Table'

export interface IAppsTable {
  apps: TApp[]
  orgId: string
  search: string
  onSearchChange: (value: string) => void
  offset: number
  pageSize: number
  hasNext: boolean
  onOffsetChange: (offset: number) => void
  loading?: boolean
  fetching?: boolean
  error?: unknown
}

const appHref = (orgId: string, app: TApp) => `/${orgId}/apps/${app?.id ?? ''}`

const appPlatform = (app: TApp): TBrandVariant | undefined => {
  const platform =
    app?.runner_config?.cloud_platform ?? app?.cloud_platform ?? ''
  const normalized = platform.toLowerCase()
  if (normalized === 'aws') return 'AWS'
  if (normalized === 'azure') return 'Azure'
  if (normalized === 'gcp') return 'GCP'
  return undefined
}

const appSource = (app: TApp) =>
  app?.sandbox_config?.public_git_vcs_config?.repo ??
  app?.sandbox_config?.connected_github_vcs_config?.repo ??
  app?.config_repo

const Platform = ({ app }: { app: TApp }) => {
  const platform = appPlatform(app)
  if (!platform) {
    return (
      <Text variant="caption" color="tertiary">
        —
      </Text>
    )
  }

  return (
    <span className="flex items-center gap-2">
      <Brand variant={platform} size={18} />
      <Text variant="caption">{platform}</Text>
    </span>
  )
}

const columnsFor = (orgId: string): ColumnDef<TApp>[] => [
  {
    id: 'name',
    header: 'App',
    size: 260,
    cell: ({ row }) => (
      <span className="flex min-w-0 flex-col gap-0.5">
        <Link
          href={appHref(orgId, row.original)}
          variant="body"
          className="w-fit"
        >
          {row.original?.name ?? 'Unnamed app'}
        </Link>
        {row.original?.id ? (
          <ID value={row.original.id} label="Copy app ID" truncate />
        ) : (
          <Text variant="caption" color="tertiary">
            —
          </Text>
        )}
      </span>
    ),
  },
  {
    id: 'status',
    header: 'Status',
    size: 150,
    cell: ({ row }) => (
      <Status
        status={row.original?.status_v2 ?? row.original?.status ?? 'unknown'}
      />
    ),
  },
  {
    id: 'platform',
    header: 'Platform',
    size: 150,
    cell: ({ row }) => <Platform app={row.original} />,
  },
  {
    id: 'source',
    header: 'Source',
    size: 280,
    cell: ({ row }) => (
      <Text variant="caption" family="mono" color="secondary" lines={1}>
        {appSource(row.original) ?? '—'}
      </Text>
    ),
  },
  {
    accessorKey: 'updated_at',
    header: 'Updated',
    size: 160,
    cell: ({ row }) => (
      <Time value={row.original?.updated_at} format="relative" />
    ),
  },
]

const AppCard = ({ app, orgId }: { app: TApp; orgId: string }) => (
  <Card className="flex h-full flex-col gap-4">
    <span className="flex min-w-0 flex-col gap-1">
      <Link href={appHref(orgId, app)} variant="heading" className="w-fit">
        {app?.name ?? 'Unnamed app'}
      </Link>
      {app?.id ? (
        <ID value={app.id} label="Copy app ID" />
      ) : (
        <Text variant="caption" color="tertiary">
          —
        </Text>
      )}
    </span>
    <Status
      status={app?.status_v2 ?? app?.status ?? 'unknown'}
      className="self-start"
    />
    <dl className="grid grid-cols-2 gap-x-4 gap-y-3">
      <div className="min-w-0">
        <Text as="dt" variant="label" color="tertiary">
          Platform
        </Text>
        <Text as="dd" className="mt-0.5">
          <Platform app={app} />
        </Text>
      </div>
      <div className="min-w-0">
        <Text as="dt" variant="label" color="tertiary">
          Updated
        </Text>
        <Text as="dd" className="mt-0.5">
          <Time value={app?.updated_at} format="relative" />
        </Text>
      </div>
      <div className="col-span-2 min-w-0">
        <Text as="dt" variant="label" color="tertiary">
          Source
        </Text>
        <Text as="dd" family="mono" color="secondary" lines={1}>
          {appSource(app) ?? '—'}
        </Text>
      </div>
    </dl>
  </Card>
)

export const AppsTable = ({
  apps,
  orgId,
  search,
  onSearchChange,
  offset,
  pageSize,
  hasNext,
  onOffsetChange,
  loading = false,
  fetching = false,
  error,
}: IAppsTable) => (
  <div className="flex min-w-0 flex-col gap-4">
    <Table
      data={apps}
      columns={columnsFor(orgId)}
      getRowId={(app) => app?.id ?? ''}
      loading={loading}
      loadingLabel="Loading apps"
      emptyState={error ? 'Apps failed to load' : 'No apps yet'}
      toolbar={
        <ListSearch
          value={search}
          onValueChange={onSearchChange}
          placeholder="Search by name or ID"
          aria-label="Search apps"
          className="w-full max-w-sm"
        />
      }
      renderCard={({ row }) => <AppCard app={row.original} orgId={orgId} />}
    />
    <Pagination
      label="Apps pagination"
      offset={offset}
      pageSize={pageSize}
      hasNext={hasNext}
      loading={fetching}
      onOffsetChange={onOffsetChange}
    />
  </div>
)
