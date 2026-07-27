import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { InstallResourceDetailPanelButton } from '@/components/install-resources/InstallResourceDetailPanel'
import type { TInstallResource } from '@/types'
import { getWorstStatusTheme } from '@/utils/status-utils'

export const HEALTH_FILTER_OPTIONS = [
  'healthy',
  'progressing',
  'degraded',
  'unhealthy',
  'unknown',
]

export type TInstallResourceRow = {
  resource: TInstallResource
  kind: string
  namespace: string
  name: string
  health: string
  message: string
  observedAt?: string
}

export type TInstallResourceGroup = {
  key: string
  heading: string
  rows: TInstallResourceRow[]
  worstHealth: string
}

function toInstallResourceRow(resource: TInstallResource): TInstallResourceRow {
  return {
    resource,
    kind: resource?.kind || '',
    namespace: resource?.namespace || '',
    name: resource?.name || '',
    health: resource?.health || 'unknown',
    message: resource?.message || '',
    observedAt: resource?.observed_at,
  }
}

function buildInstallResourceGroups(
  resources: TInstallResource[],
  keyOf: (resource: TInstallResource) => string,
  headingFor: (key: string) => string
): TInstallResourceGroup[] {
  const groups = new Map<string, TInstallResourceRow[]>()

  resources.forEach((resource) => {
    const key = keyOf(resource) || 'unknown'
    groups.set(key, [...(groups.get(key) ?? []), toInstallResourceRow(resource)])
  })

  return Array.from(groups.entries())
    .map(([key, rows]) => ({
      key,
      heading: headingFor(key),
      rows,
      worstHealth: getWorstStatusTheme(rows.map((row) => row.health)).worstStatus,
    }))
    .sort((a, b) => a.heading.localeCompare(b.heading))
}

export function isSandboxResource(resource: TInstallResource): boolean {
  return resource?.source === 'sandbox'
}

export function groupComponentResources(
  resources: TInstallResource[],
  componentNames: Record<string, string>
): TInstallResourceGroup[] {
  return buildInstallResourceGroups(
    resources.filter((resource) => !isSandboxResource(resource)),
    (resource) => resource.install_component_id || 'unknown',
    (key) => componentNames[key] || key
  )
}

export function groupSandboxResources(
  resources: TInstallResource[]
): TInstallResourceGroup[] {
  return buildInstallResourceGroups(
    resources.filter(isSandboxResource),
    (resource) => resource.owner_name || 'unknown',
    (key) => key
  )
}

const columns: ColumnDef<TInstallResourceRow>[] = [
  {
    accessorKey: 'kind',
    header: 'Kind',
    cell: (info) => <Text>{info.getValue() as string}</Text>,
  },
  {
    accessorKey: 'namespace',
    header: 'Namespace',
    cell: (info) =>
      info.getValue() ? (
        <Text>{info.getValue() as string}</Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
  },
  {
    accessorKey: 'name',
    header: 'Name',
    cell: (info) => (
      <Text variant="body" weight="strong">
        {info.getValue() as string}
      </Text>
    ),
  },
  {
    accessorKey: 'health',
    header: 'Health',
    cell: (info) => <Status variant="badge" status={info.getValue() as string} />,
  },
  {
    accessorKey: 'message',
    header: 'Message',
    cell: (info) =>
      info.getValue() ? (
        <Text variant="subtext" theme="neutral" className="line-clamp-2 max-w-[320px]">
          {info.getValue() as string}
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
    enableSorting: false,
  },
  {
    accessorKey: 'observedAt',
    header: 'Age',
    cell: (info) =>
      info.getValue() ? (
        <Time variant="subtext" time={info.getValue() as string} format="relative" />
      ) : (
        <Icon variant="MinusIcon" />
      ),
  },
  {
    id: 'action',
    enableSorting: false,
    header: '',
    cell: (info) => (
      <InstallResourceDetailPanelButton
        installResource={info.row.original.resource}
      />
    ),
  },
]

interface ISingleSelectFilterDropdown {
  id: string
  label: string
  options: string[]
  value: string
  onChange: (value: string) => void
}

const SingleSelectFilterDropdown = ({
  id,
  label,
  options,
  value,
  onChange,
}: ISingleSelectFilterDropdown) => {
  if (options.length === 0) return null

  return (
    <Dropdown
      alignment="left"
      id={id}
      buttonText={
        <>
          <Icon variant="FunnelIcon" size={14} />
          {label}
          {value ? ` (${value})` : ''}
        </>
      }
    >
      <Menu className="!p-0 !w-64">
        <form onReset={() => onChange('')}>
          <div className="flex flex-col gap-0.5 max-h-64 overflow-y-auto w-full p-2">
            <RadioInput
              checked={value === ''}
              labelProps={{ labelText: 'All' }}
              name={id}
              onChange={() => onChange('')}
              value=""
            />
            {options.map((option) => (
              <RadioInput
                key={option}
                checked={value === option}
                labelProps={{ labelText: option }}
                name={id}
                onChange={() => onChange(option)}
                value={option}
              />
            ))}
          </div>
          <div className="flex flex-col gap-0.5 px-2 pb-2 w-full">
            <hr />
            <Button className="mt-1" isMenuButton type="reset" variant="ghost">
              Clear
              <Icon variant="XIcon" />
            </Button>
          </div>
        </form>
      </Menu>
    </Dropdown>
  )
}

const InstallResourceGroupTable = ({ group }: { group: TInstallResourceGroup }) => (
  <div className="flex flex-col gap-3">
    <div className="flex items-center gap-2">
      <Text variant="base" weight="strong">
        {group.heading}
      </Text>
      <Status variant="badge" status={group.worstHealth} />
    </div>
    <Table<TInstallResourceRow> columns={columns} data={group.rows} enableSearch={false} />
  </div>
)

interface IInstallResourcesTable {
  componentGroups: TInstallResourceGroup[]
  sandboxGroups: TInstallResourceGroup[]
  isLoading: boolean
  kind: string
  namespace: string
  health: string
  kindOptions: string[]
  namespaceOptions: string[]
  onKindChange: (value: string) => void
  onNamespaceChange: (value: string) => void
  onHealthChange: (value: string) => void
}

export const InstallResourcesTable = ({
  componentGroups,
  sandboxGroups,
  isLoading,
  kind,
  namespace,
  health,
  kindOptions,
  namespaceOptions,
  onKindChange,
  onNamespaceChange,
  onHealthChange,
}: IInstallResourcesTable) => {
  if (isLoading) {
    return <TableSkeleton columns={columns} skeletonRows={5} />
  }

  const hasResources = componentGroups.length > 0 || sandboxGroups.length > 0

  return (
    <div className="flex flex-col gap-6 md:gap-8">
      <div className="flex flex-wrap items-center gap-3">
        <SingleSelectFilterDropdown
          id="resource-filter-kind"
          label="Kind"
          options={kindOptions}
          value={kind}
          onChange={onKindChange}
        />
        <SingleSelectFilterDropdown
          id="resource-filter-namespace"
          label="Namespace"
          options={namespaceOptions}
          value={namespace}
          onChange={onNamespaceChange}
        />
        <SingleSelectFilterDropdown
          id="resource-filter-health"
          label="Health"
          options={HEALTH_FILTER_OPTIONS}
          value={health}
          onChange={onHealthChange}
        />
      </div>

      {!hasResources ? (
        <EmptyState
          variant="table"
          emptyTitle="No resources yet"
          emptyMessage="Resources will appear here once the runner reports component health for this install."
        />
      ) : (
        <>
          {componentGroups.length > 0 ? (
            <div className="flex flex-col gap-4">
              <Text variant="h3" weight="strong">
                Components
              </Text>
              <div className="flex flex-col gap-6 md:gap-8">
                {componentGroups.map((group) => (
                  <InstallResourceGroupTable key={group.key} group={group} />
                ))}
              </div>
            </div>
          ) : null}

          {sandboxGroups.length > 0 ? (
            <div className="flex flex-col gap-4">
              <Text variant="h3" weight="strong">
                Sandbox
              </Text>
              <div className="flex flex-col gap-6 md:gap-8">
                {sandboxGroups.map((group) => (
                  <InstallResourceGroupTable key={group.key} group={group} />
                ))}
              </div>
            </div>
          ) : null}
        </>
      )}
    </div>
  )
}
