import type { ColumnDef } from '@tanstack/react-table'
import type { TApp } from '@/types/ctl-api.types'
import { Brand, type TBrandVariant } from '../../atoms/Brand'
import { Card } from '../../atoms/Card'
import { Link } from '../../atoms/Link'
import { Status } from '../../atoms/Status'
import { Text } from '../../atoms/Text'
import { AppSource, appSourceFromApp } from '../../molecules/AppSource'
import { ID } from '../../molecules/ID'
import { ListSearch } from '../../molecules/ListSearch'
import { Pagination } from '../../molecules/Pagination'
import { Time } from '../../molecules/Time'
import { appSetupHref } from '../../../utils/hrefs'
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
  incompleteIds?: ReadonlySet<string>
  loading?: boolean
  fetching?: boolean
  error?: unknown
}

const appHref = (orgId: string, app: TApp, incomplete: boolean) =>
  incomplete && app?.id
    ? appSetupHref(orgId, app.id)
    : `/${orgId}/apps/${app?.id ?? ''}`

const AppStatus = ({
  app,
  incomplete,
  className,
}: {
  app: TApp
  incomplete: boolean
  className?: string
}) =>
  incomplete ? (
    <Status
      status="pending"
      theme="warn"
      label="Setup incomplete"
      className={className}
    />
  ) : (
    <Status
      status={app?.status_v2 ?? app?.status ?? 'unknown'}
      className={className}
    />
  )

const appPlatform = (app: TApp): TBrandVariant | undefined => {
  const platform =
    app?.runner_config?.cloud_platform ?? app?.cloud_platform ?? ''
  const normalized = platform.toLowerCase()
  if (normalized === 'aws') return 'AWS'
  if (normalized === 'azure') return 'Azure'
  if (normalized === 'gcp') return 'GCP'
  return undefined
}

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

const columnsFor = (
  orgId: string,
  incompleteIds?: ReadonlySet<string>
): ColumnDef<TApp>[] => [
  {
    id: 'name',
    header: 'App',
    size: 260,
    cell: ({ row }) => (
      <span className="flex min-w-0 flex-col gap-0.5">
        <Link
          href={appHref(
            orgId,
            row.original,
            Boolean(row.original?.id && incompleteIds?.has(row.original.id))
          )}
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
      <AppStatus
        app={row.original}
        incomplete={Boolean(
          row.original?.id && incompleteIds?.has(row.original.id)
        )}
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
      <AppSource
        source={appSourceFromApp(row.original)}
        className="max-w-full"
      />
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

const AppCard = ({
  app,
  orgId,
  incomplete,
}: {
  app: TApp
  orgId: string
  incomplete: boolean
}) => (
  <Card className="flex h-full flex-col gap-4">
    <span className="flex min-w-0 flex-col gap-1">
      <Link
        href={appHref(orgId, app, incomplete)}
        variant="heading"
        className="w-fit"
      >
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
    <AppStatus app={app} incomplete={incomplete} className="self-start" />
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
        <AppSource
          source={appSourceFromApp(app)}
          className="mt-0.5 max-w-full"
        />
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
  incompleteIds,
  loading = false,
  fetching = false,
  error,
}: IAppsTable) => (
  <div className="flex min-w-0 flex-col gap-4">
    <Table
      data={apps}
      columns={columnsFor(orgId, incompleteIds)}
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
      renderCard={({ row }) => (
        <AppCard
          app={row.original}
          orgId={orgId}
          incomplete={Boolean(
            row.original?.id && incompleteIds?.has(row.original.id)
          )}
        />
      )}
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
