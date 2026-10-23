import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { EmptyState } from '@/components/common/EmptyState'
import { RadioFilterDropdown } from '@/components/common/RadioFilterDropdown'
import { Toggle } from '@/components/common/form/Toggle'
import { Icon } from '@/components/common/Icon'
import { Loading } from '@/components/common/Loading'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { InstallResourceBoard } from '@/components/install-resources/InstallResourceBoard/InstallResourceBoard'
import { InstallResourceDetailPanelButton } from '@/components/install-resources/InstallResourceDetailPanel'
import { resourceStatusFields } from '@/components/install-resources/InstallResourceDetailPanel/resource-details'
import { ResourceKind } from '@/components/install-resources/ResourceKind'
import { RemovedFromAppConfigBadge } from '@/components/installs/RemovedFromAppConfig'
import { panelTriggerClass } from '@/components/surfaces/panel-trigger'
import type { TInstallResource } from '@/types'
import { isStaleObservation, latestTimestamp } from '@/utils/time-utils'
import {
  isHealthCheckResource,
  resourceIdentity,
  resourceOwner,
} from '@/components/install-resources/resource-utils'

export interface IInstallHealthResources {
  resources?: TInstallResource[]
  componentNames?: Record<string, string>
  isLoading?: boolean
  error?: string
  clusterAccessError?: string
  caption?: string
  defaultComponent?: string
}

const CheckName = ({ row }: { row: { original: TInstallResource } }) => (
  <InstallResourceDetailPanelButton
    key={resourceIdentity(row.original)}
    installResource={row.original}
    className={panelTriggerClass}
    aria-label={`Inspect check ${row.original.name}`}
  >
    <Icon variant="HeartbeatIcon" theme="brand" />
    <Text family="mono">{row.original.name}</Text>
  </InstallResourceDetailPanelButton>
)

export const InstallHealthResources = ({
  resources = [],
  componentNames = {},
  isLoading,
  error,
  clusterAccessError,
  caption,
  defaultComponent = '',
}: IInstallHealthResources) => {
  const [params, setParams] = useSearchParams()
  const component = params.get('component') ?? defaultComponent
  const namespace = params.get('namespace') || ''
  const includeSandbox = params.get('sandbox') !== 'false'
  const [, tick] = useState(0)
  useEffect(() => {
    const interval = setInterval(() => tick((value) => value + 1), 30_000)
    return () => clearInterval(interval)
  }, [])

  const setFilter = (key: string, value: string) =>
    setParams(
      (previous) => {
        previous.set(key, value)
        if (key === 'component' || key === 'sandbox')
          previous.delete('namespace')
        return previous
      },
      { replace: true }
    )

  const managed = resources.filter(
    (resource) => !isHealthCheckResource(resource)
  )
  const checks = resources.filter(isHealthCheckResource)
  const owners = new Map(
    managed
      .filter((resource) => resource.source !== 'sandbox')
      .map((resource) => {
        const owner = resourceOwner(resource, componentNames)
        return [owner.key, owner.label]
      })
  )
  const scoped = managed.filter((resource) =>
    resource.source === 'sandbox'
      ? includeSandbox
      : !component || resourceOwner(resource, componentNames).key === component
  )
  const namespaces = [
    ...new Set(
      scoped
        .map((resource) => resource.namespace)
        .filter((value): value is string => !!value)
    ),
  ].sort()
  const items = scoped
    .filter((resource) => !namespace || resource.namespace === namespace)
    .map((resource) => ({
      id: resourceIdentity(resource),
      resource,
      summary: [
        resource.namespace,
        resourceOwner(resource, componentNames).label,
      ]
        .filter(Boolean)
        .join(' · '),
    }))
  const observedAt = latestTimestamp(
    scoped.map((resource) => resource.observed_at)
  )
  const columns: ColumnDef<TInstallResource>[] = [
    {
      id: 'name',
      header: 'Check',
      cell: CheckName,
    },
    {
      id: 'component',
      header: 'Component',
      cell: ({ row }) => (
        <Text family="mono" variant="subtext">
          {resourceOwner(row.original, componentNames).label}
        </Text>
      ),
    },
    {
      id: 'kind',
      header: 'Type',
      cell: ({ row }) => <ResourceKind resource={row.original} />,
    },
    {
      id: 'result',
      header: 'Result',
      cell: ({ row }) => (
        <div className="flex flex-col gap-1">
          {row.original.removed_from_config ? (
            <RemovedFromAppConfigBadge kind="probe" />
          ) : null}
          <Text variant="subtext" family="mono">
            {resourceStatusFields(row.original).find(
              ([label]) => label === 'Result'
            )?.[1] || 'Result not reported'}
          </Text>
        </div>
      ),
    },
    {
      id: 'latency',
      header: 'Latency',
      cell: ({ row }) => (
        <Text family="mono" variant="subtext">
          {resourceStatusFields(row.original).find(
            ([label]) => label === 'Latency'
          )?.[1] || 'Not reported'}
        </Text>
      ),
    },
    {
      id: 'observed_at',
      header: 'Last observed',
      cell: ({ row }) => (
        <span className="flex flex-wrap items-center gap-2">
          {isStaleObservation(
            row.original.observed_at,
            row.original.stale_after_seconds || undefined
          ) ? (
            <Badge theme="warn" size="sm">
              Stale
            </Badge>
          ) : null}
          {row.original.observed_at ? (
            <Time
              variant="subtext"
              time={row.original.observed_at}
              format="relative"
              shouldTick
            />
          ) : (
            <Text variant="subtext" theme="neutral">
              Not reported
            </Text>
          )}
        </span>
      ),
    },
  ]

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center gap-2">
        <Icon variant="GraphIcon" theme="brand" />
        <Text weight="strong">Resource canvas</Text>
        <Text variant="subtext" theme="neutral">
          {caption || 'Latest reported observations · Refresh ~15s'}
          {observedAt ? (
            <>
              {' · Last observation '}
              <Time
                time={observedAt}
                format="relative"
                variant="subtext"
                shouldTick
              />
            </>
          ) : null}
        </Text>
      </div>
      {error ? (
        <Banner theme="error">
          <div className="flex flex-col gap-1">
            <Text weight="strong">{error}</Text>
            {resources.length ? (
              <Text variant="subtext">
                Showing the last received observations.
              </Text>
            ) : null}
          </div>
        </Banner>
      ) : null}
      {isLoading && !resources.length ? (
        <div className="flex items-center justify-center gap-3 h-64">
          <Loading />
          <Text theme="neutral">Loading the latest snapshot…</Text>
        </div>
      ) : !resources.length ? (
        <EmptyState
          emptyTitle={
            error
              ? 'Resources failed to load'
              : clusterAccessError
                ? 'Resource collection failed'
                : 'No observations yet'
          }
          emptyMessage={
            error
              ? 'Try refreshing the page to load the latest observations.'
              : clusterAccessError
                ? 'Resources will appear after the runner can read the cluster again.'
                : 'Resources and checks will appear after the runner sends its first report.'
          }
        />
      ) : (
        <>
          {managed.length ? (
            <InstallResourceBoard
              key={`${component}/${namespace}/${includeSandbox}`}
              items={items}
              filters={
                <>
                  <RadioFilterDropdown
                    id="health-component"
                    label="Component"
                    options={[...owners.entries()]
                      .sort((a, b) => a[1].localeCompare(b[1]))
                      .map(([value, label]) => ({ label, value }))}
                    selected={component || undefined}
                    onChange={(value) => setFilter('component', value || '')}
                  />
                  <RadioFilterDropdown
                    id="health-namespace"
                    label="Namespace"
                    options={namespaces.map((value) => ({
                      label: value,
                      value,
                    }))}
                    selected={namespace || undefined}
                    onChange={(value) => setFilter('namespace', value || '')}
                  />
                </>
              }
              toggles={
                <Toggle
                  checked={includeSandbox}
                  onChange={(checked) => setFilter('sandbox', String(checked))}
                  label="Include sandbox"
                  className="px-2"
                />
              }
            />
          ) : null}
          {checks.length ? (
            <div className="flex flex-col gap-3">
              <Text weight="strong">Probes and custom checks</Text>
              <Text variant="subtext" theme="neutral">
                Latest results across the install, independent of resource
                filters. Freshness is shown separately from the reported result.
              </Text>
              <Table columns={columns} data={checks} enableSearch={false} />
            </div>
          ) : null}
        </>
      )}
    </div>
  )
}
