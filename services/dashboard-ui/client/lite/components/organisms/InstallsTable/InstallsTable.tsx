import type { ColumnDef } from '@tanstack/react-table'
import type { TInstall } from '@/types/ctl-api.types'
import { humanize } from '@/utils/string-utils'
import { Badge } from '../../atoms/Badge'
import { Brand, type TBrandVariant } from '../../atoms/Brand'
import { Card } from '../../atoms/Card'
import type { TIconVariant } from '../../atoms/Icon'
import { Link } from '../../atoms/Link'
import { Status } from '../../atoms/Status'
import { Text } from '../../atoms/Text'
import {
  FilterDropdown,
  type IFilterMenuOption,
} from '../../molecules/FilterMenu'
import { CloudRegion } from '../../molecules/CloudRegion'
import { ID } from '../../molecules/ID'
import { ListSearch } from '../../molecules/ListSearch'
import { Pagination } from '../../molecules/Pagination'
import { Time } from '../../molecules/Time'
import { Table } from '../Table/Table'

export type TLabelColors = Record<string, Record<string, string>>

export interface IInstallFilter {
  label: string
  options: readonly IFilterMenuOption<string>[]
  selected: Set<string>
  onToggle: (value: string) => void
  onIsolate: (value: string) => void
  onReset: () => void
  constrained: boolean
}

export interface IInstallsTable {
  installs: TInstall[]
  orgId: string
  search: string
  onSearchChange: (value: string) => void
  offset: number
  pageSize: number
  hasNext: boolean
  onOffsetChange: (offset: number) => void
  labelFilter?: IInstallFilter
  branchFilter?: IInstallFilter
  labelColors?: TLabelColors
  loading?: boolean
  fetching?: boolean
  error?: unknown
}

const installHref = (orgId: string, install: TInstall) =>
  `/${orgId}/installs/${install?.id ?? ''}`

const appHref = (orgId: string, install: TInstall) =>
  `/${orgId}/apps/${install?.app_id ?? ''}`

const installPlatform = (install: TInstall): TBrandVariant | undefined => {
  const normalized = (install?.cloud_platform ?? '').toLowerCase()
  if (normalized === 'aws') return 'AWS'
  if (normalized === 'azure') return 'Azure'
  if (normalized === 'gcp') return 'GCP'
  return undefined
}

const installRegion = (install: TInstall) =>
  install?.aws_account?.region ??
  install?.gcp_account?.region ??
  install?.azure_account?.location

type TInstallFacet = {
  id: string
  title: string
  icon: TIconVariant
  status: string
  description?: string
}

const STALE_PHASE_STATUSES = new Set([
  'active',
  'pending',
  'executing',
  'queued',
  'planning',
  'syncing',
])

const phaseAdjusted = (status?: string, phase?: string) => {
  if (!status || !STALE_PHASE_STATUSES.has(status)) return status
  if (phase === 'deprovisioned' || phase === 'deprovisioning') return phase
  return status
}

const facetTitle = (label: string, status: string) =>
  `${label} ${humanize(status).toLowerCase()}`

const installFacets = (install: TInstall): TInstallFacet[] => {
  const phase = install?.lifecycle_phase?.phase
  const adjust = (status?: string) => phaseAdjusted(status, phase)

  const runnerStatus = adjust(install?.runner_status) ?? 'unknown'
  const sandboxBase = adjust(install?.sandbox_status)
  const sandboxStatus =
    install?.sandbox_health_status && sandboxBase === 'active'
      ? install.sandbox_health_status
      : (sandboxBase ?? 'unknown')
  const componentStatus =
    adjust(install?.composite_component_status) ?? 'unknown'

  const facets: TInstallFacet[] = [
    {
      id: 'runner',
      title: facetTitle('Runner', runnerStatus),
      icon: 'SneakerMoveIcon',
      status: runnerStatus,
      description: install?.runner_status_description,
    },
    {
      id: 'sandbox',
      title: facetTitle('Sandbox', sandboxStatus),
      icon: 'ShippingContainerIcon',
      status: sandboxStatus,
      description:
        sandboxStatus === sandboxBase
          ? install?.sandbox_status_description
          : (install?.sandbox_health_message ??
            install?.sandbox_status_description),
    },
    {
      id: 'components',
      title: facetTitle('Components', componentStatus),
      icon: 'CardsIcon',
      status: componentStatus,
      description: install?.composite_component_status_description,
    },
  ]

  if (install?.composite_health_status) {
    facets.push({
      id: 'health',
      title: facetTitle('Health', install.composite_health_status),
      icon: 'HeartbeatIcon',
      status: install.composite_health_status,
      description: install?.composite_health_status_description,
    })
  }

  if (install?.drifted_objects) {
    const drifted = install.drifted_objects.length
    facets.push({
      id: 'drift',
      title: drifted ? 'Drift detected' : 'No drift',
      icon: 'FileDashedIcon',
      status: drifted ? 'warn' : 'active',
      description: drifted
        ? `${drifted} resource${drifted === 1 ? '' : 's'} have drifted from the last applied state.`
        : undefined,
    })
  }

  return facets
}

const InstallStatuses = ({ install }: { install: TInstall }) => (
  <span className="flex flex-wrap items-center gap-1.5">
    {installFacets(install).map((facet) => (
      <Status
        key={facet.id}
        status={facet.status}
        icon={facet.icon}
        label={facet.title}
        description={facet.description}
        variant="icon"
      />
    ))}
  </span>
)

const Platform = ({ install }: { install: TInstall }) => {
  const platform = installPlatform(install)
  const region = installRegion(install)
  const cloudPlatform =
    platform === 'AWS'
      ? 'aws'
      : platform === 'Azure'
        ? 'azure'
        : platform === 'GCP'
          ? 'gcp'
          : 'unknown'

  if (!platform && !region) {
    return (
      <Text variant="caption" color="tertiary">
        —
      </Text>
    )
  }

  return (
    <span className="flex min-w-0 items-center gap-2">
      {platform ? <Brand variant={platform} size={18} /> : null}
      {region ? (
        <CloudRegion
          platform={cloudPlatform}
          region={cloudPlatform === 'azure' ? undefined : region}
          location={cloudPlatform === 'azure' ? region : undefined}
          lines={1}
        />
      ) : (
        <Text variant="caption" lines={1}>
          {platform}
        </Text>
      )}
    </span>
  )
}

const InstallLabels = ({
  install,
  labelColors,
}: {
  install: TInstall
  labelColors?: TLabelColors
}) => {
  const labels = Object.entries(install?.labels ?? {}).sort(([a], [b]) =>
    a.localeCompare(b)
  )
  if (!labels.length) {
    return (
      <Text variant="caption" color="tertiary">
        —
      </Text>
    )
  }

  const colors = labelColors?.[install?.app_id ?? '']

  return (
    <span className="flex min-w-0 flex-wrap gap-1">
      {labels.map(([key, value]) => (
        <Badge
          key={key}
          variant="code"
          labelKey={key}
          labelValue={value}
          color={colors?.[key]}
        />
      ))}
    </span>
  )
}

export const columnsFor = (
  orgId: string,
  labelColors?: TLabelColors
): ColumnDef<TInstall>[] => [
  {
    id: 'name',
    header: 'Install',
    size: 200,
    cell: ({ row }) => (
      <span className="flex min-w-0 flex-col gap-0.5">
        <Link
          href={installHref(orgId, row.original)}
          variant="body"
          className="w-fit"
        >
          {row.original?.name ?? 'Unnamed install'}
        </Link>
        {row.original?.id ? (
          <ID value={row.original.id} label="Copy install ID" truncate />
        ) : (
          <Text variant="caption" color="tertiary">
            —
          </Text>
        )}
      </span>
    ),
  },
  {
    id: 'app',
    header: 'App',
    size: 140,
    cell: ({ row }) =>
      row.original?.app_id ? (
        <Link href={appHref(orgId, row.original)} variant="body">
          {row.original?.app?.name ?? row.original.app_id}
        </Link>
      ) : (
        <Text variant="caption" color="tertiary">
          —
        </Text>
      ),
  },
  {
    id: 'statuses',
    header: 'Statuses',
    size: 130,
    cell: ({ row }) => <InstallStatuses install={row.original} />,
  },
  {
    id: 'platform',
    header: 'Cloud',
    size: 165,
    cell: ({ row }) => <Platform install={row.original} />,
  },
  {
    id: 'labels',
    header: 'Labels',
    size: 165,
    cell: ({ row }) => (
      <InstallLabels install={row.original} labelColors={labelColors} />
    ),
  },
  {
    id: 'branch',
    header: 'Branch',
    size: 105,
    cell: ({ row }) => (
      <Text variant="caption" family="mono" color="secondary" lines={1}>
        {row.original?.app_branch?.name ?? '—'}
      </Text>
    ),
  },
  {
    accessorKey: 'updated_at',
    header: 'Updated',
    size: 105,
    cell: ({ row }) => (
      <Time value={row.original?.updated_at} format="relative" />
    ),
  },
]

const InstallCard = ({
  install,
  orgId,
  labelColors,
}: {
  install: TInstall
  orgId: string
  labelColors?: TLabelColors
}) => {
  return (
    <Card className="flex h-full flex-col gap-4">
      <span className="flex min-w-0 flex-col gap-1">
        <Link
          href={installHref(orgId, install)}
          variant="heading"
          className="w-fit"
        >
          {install?.name ?? 'Unnamed install'}
        </Link>
        {install?.id ? (
          <ID value={install.id} label="Copy install ID" />
        ) : (
          <Text variant="caption" color="tertiary">
            —
          </Text>
        )}
      </span>
      <InstallStatuses install={install} />
      <dl className="grid grid-cols-2 gap-x-4 gap-y-3">
        <div className="min-w-0">
          <Text as="dt" variant="label" color="tertiary">
            App
          </Text>
          <Text as="dd" className="mt-0.5">
            {install?.app_id ? (
              <Link href={appHref(orgId, install)} variant="body">
                {install?.app?.name ?? install.app_id}
              </Link>
            ) : (
              '—'
            )}
          </Text>
        </div>
        <div className="min-w-0">
          <Text as="dt" variant="label" color="tertiary">
            Cloud
          </Text>
          <Text as="dd" className="mt-0.5">
            <Platform install={install} />
          </Text>
        </div>
        <div className="min-w-0">
          <Text as="dt" variant="label" color="tertiary">
            Branch
          </Text>
          <Text as="dd" family="mono" color="secondary" lines={1}>
            {install?.app_branch?.name ?? '—'}
          </Text>
        </div>
        <div className="min-w-0">
          <Text as="dt" variant="label" color="tertiary">
            Updated
          </Text>
          <Text as="dd" className="mt-0.5">
            <Time value={install?.updated_at} format="relative" />
          </Text>
        </div>
        <div className="col-span-2 min-w-0">
          <Text as="dt" variant="label" color="tertiary">
            Labels
          </Text>
          <Text as="dd" className="mt-1">
            <InstallLabels install={install} labelColors={labelColors} />
          </Text>
        </div>
      </dl>
    </Card>
  )
}

const filterControl = (filter: IInstallFilter) => (
  <FilterDropdown
    key={filter.label}
    label={filter.label}
    options={filter.options}
    selected={filter.selected}
    onToggle={filter.onToggle}
    onIsolate={filter.onIsolate}
    onReset={filter.onReset}
    isConstrained={filter.constrained}
  />
)

export const InstallsTable = ({
  installs,
  orgId,
  search,
  onSearchChange,
  offset,
  pageSize,
  hasNext,
  onOffsetChange,
  labelFilter,
  branchFilter,
  labelColors,
  loading = false,
  fetching = false,
  error,
}: IInstallsTable) => (
  <div className="flex min-w-0 flex-col gap-4">
    <Table
      data={installs}
      columns={columnsFor(orgId, labelColors)}
      getRowId={(install) => install?.id ?? ''}
      loading={loading}
      loadingLabel="Loading installs"
      emptyState={error ? 'Installs failed to load' : 'No installs yet'}
      toolbar={
        <>
          <ListSearch
            value={search}
            onValueChange={onSearchChange}
            placeholder="Search by name, branch or ID"
            aria-label="Search installs"
            className="w-full max-w-sm"
          />
          {labelFilter ? filterControl(labelFilter) : null}
          {branchFilter ? filterControl(branchFilter) : null}
        </>
      }
      renderCard={({ row }) => (
        <InstallCard
          install={row.original}
          orgId={orgId}
          labelColors={labelColors}
        />
      )}
    />
    <Pagination
      label="Installs pagination"
      offset={offset}
      pageSize={pageSize}
      hasNext={hasNext}
      loading={fetching}
      onOffsetChange={onOffsetChange}
    />
  </div>
)
