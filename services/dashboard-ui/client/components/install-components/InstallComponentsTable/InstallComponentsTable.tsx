import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import { ComponentType } from '@/components/components/ComponentType'
import { InstallComponentDependencies } from '@/components/install-components/InstallComponentDependencies'
import { QuickComponentManagementDropdown } from '@/components/install-components/management/QuickComponentManagementDropdown'
import { RemovedFromAppConfigBadge } from '@/components/installs/RemovedFromAppConfig'
import type { TComponentConfig, TInstallComponent } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

type TComponentDeps = {
  id: string
  component_id: string
  dependencies: string[]
}

export type InstallComponentRow = {
  componentId: string
  componentName: string
  componentType: ReactNode
  toggleStatus: ReactNode
  overrideStatus: ReactNode
  deployStatus: ReactNode
  driftStatus: ReactNode
  health?: string
  healthMessage?: string
  href: string
  action: ReactNode
  dependencies: ReactNode
  labels: ReactNode
  removed?: boolean
}

export function parseInstallComponentSummaryToTableData(
  components: TInstallComponent[],
  deps: TComponentDeps[],
  orgId: string,
  installId: string,
  configConnections?: TComponentConfig[],
  componentToggles?: { [key: string]: boolean },
  labelColors?: Record<string, string>,
  overriddenComponentNames?: Set<string>,
  removed = false
): InstallComponentRow[] {
  return components.map((component) => {
    const depIndex = deps?.findIndex((dep) => dep?.id === component?.id)

    const configConnection = configConnections?.find(
      (c) => c?.component_id === component?.component_id
    )
    const isToggleable = configConnection?.toggleable === true

    let toggleStatusNode: ReactNode = <Icon variant="MinusIcon" />
    if (isToggleable) {
      const componentId = component?.component_id
      let isDisabled = false
      if (componentToggles && componentId && componentId in componentToggles) {
        isDisabled = !componentToggles[componentId]
      } else {
        isDisabled = !configConnection?.default_enabled
      }
      toggleStatusNode = (
        <Badge size="sm" theme={isDisabled ? 'neutral' : 'success'}>
          {isDisabled ? 'Disabled' : 'Enabled'}
        </Badge>
      )
    }

    return {
      componentId: component.component_id,
      componentName: component.component?.name,
      componentType: (
        <ComponentType
          type={component?.component?.type}
          variant="subtext"
          colorVariant="color"
        />
      ),
      toggleStatus: toggleStatusNode,
      overrideStatus: overriddenComponentNames?.has(
        component.component?.name ?? ''
      ) ? (
        <Badge size="sm" theme="info">
          Customized
        </Badge>
      ) : (
        <Icon variant="MinusIcon" />
      ),
      deployStatus: (
        <Tooltip
          position="top"
          tipContentClassName="!p-0"
          tipContent={
            <div className="w-fit max-w-64">
              {component?.status_v2?.status ? (
                <>
                  <Time
                    className="!text-nowrap px-2 py-1"
                    variant="subtext"
                    seconds={component?.status_v2?.created_at_ts}
                    weight="strong"
                  />
                  <hr className="my-1" />
                  <Text as="div" className="flex px-2 pb-2" variant="subtext">
                    {toSentenceCase(
                      component?.status_v2?.status_human_description
                    )}
                  </Text>
                </>
              ) : (
                <Text flex nowrap className="p-2" variant="subtext">
                  Status unknown
                </Text>
              )}
            </div>
          }
        >
          <Status variant="badge" status={component.status_v2?.status} />
        </Tooltip>
      ),
      driftStatus: component?.drifted_object ? (
        <Status variant="badge" status="drifted" />
      ) : (
        <Icon variant="MinusIcon" />
      ),
      health:
        component?.health_status === 'not-applicable'
          ? undefined
          : component?.health_status,
      healthMessage: component?.health_status_description,
      dependencies: (
        <InstallComponentDependencies deps={deps?.at(depIndex)?.dependencies} />
      ),
      labels: (() => {
        const lbls = component.component?.labels
        if (!lbls || Object.keys(lbls).length === 0) return <Icon variant="MinusIcon" />
        return (
          <span className="flex flex-wrap gap-1">
            {Object.keys(lbls)
              .sort()
              .map((k) => (
                <LabelBadge key={k} labelKey={k} labelValue={lbls[k]} size="sm" customColor={labelColors?.[k]} />
              ))}
          </span>
        )
      })(),
      href: `/${orgId}/installs/${installId}/components/${component.component_id}`,
      action: (
        <div className="hidden md:block">
          <QuickComponentManagementDropdown
            installComponent={component}
            orgId={orgId}
            installId={installId}
            removed={removed}
          />
        </div>
      ),
      removed,
    }
  })
}

const healthColumn: ColumnDef<InstallComponentRow> = {
  enableSorting: false,
  accessorKey: 'health',
  header: 'Health',
  cell: (info) => {
    const health = info.getValue() as string | undefined
    if (!health) return <Icon variant="MinusIcon" />
    const badge = <Status variant="badge" status={health} />
    const message = info.row.original.healthMessage
    if (!message) return badge
    return (
      <Tooltip
        position="top"
        tipContentClassName="!p-0"
        tipContent={
          <Text as="div" className="flex w-fit max-w-96 p-2" variant="subtext">
            {message}
          </Text>
        }
      >
        {badge}
      </Tooltip>
    )
  },
}

const baseColumns: ColumnDef<InstallComponentRow>[] = [
  {
    accessorKey: 'componentName',
    header: 'Component name',
    cell: (info) => (
      <span>
        <Text variant="body" flex className="items-center gap-2">
          <Link href={info.row.original.href}>{info.getValue() as string}</Link>
          {info.row.original.removed ? (
            <RemovedFromAppConfigBadge kind="component" />
          ) : null}
        </Text>
        <ID>{info.row.original.componentId as string}</ID>
      </span>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'componentType',
    header: 'Type',
    cell: (info) => <Text>{info.getValue() as string}</Text>,
  },
  {
    enableSorting: true,
    accessorKey: 'dependencies',
    header: 'Dependencies',
    cell: (info) => (
      <Text as="div" className="flex">{info.getValue() as ReactNode}</Text>
    ),
  },
  {
    enableSorting: false,
    accessorKey: 'labels',
    header: 'Labels',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    enableSorting: false,
    accessorKey: 'deployStatus',
    header: 'Latest deploy',
    cell: (info) => (
      <Text className="flex">{info.getValue() as ReactNode}</Text>
    ),
  },
  {
    enableSorting: false,
    accessorKey: 'driftStatus',
    header: 'Drifted',
    cell: (info) => (
      <Text as="div" className="flex">{info.getValue() as ReactNode}</Text>
    ),
  },
  {
    enableSorting: false,
    accessorKey: 'toggleStatus',
    header: 'Toggle',
    cell: (info) => (
      <Text as="div" className="flex">{info.getValue() as ReactNode}</Text>
    ),
  },
  {
    enableSorting: false,
    accessorKey: 'overrideStatus',
    header: 'Overrides',
    cell: (info) => (
      <Text as="div" className="flex">{info.getValue() as ReactNode}</Text>
    ),
  },
  {
    enableSorting: false,
    accessorKey: 'action',
    id: 'action',
    header: '',
    cell: (info) => info.getValue<ReactNode>(),
  },
]

function buildColumns(showHealth: boolean): ColumnDef<InstallComponentRow>[] {
  if (!showHealth) return baseColumns
  const cols = [...baseColumns]
  cols.splice(1, 0, healthColumn)
  return cols
}

interface IInstallComponentsTable {
  data: InstallComponentRow[]
  filterActions: ReactNode
  pagination: {
    hasNext: boolean
    offset: number
    limit: number
  }
  isLoading: boolean
  showHealth?: boolean
}

export const InstallComponentsTable = ({
  data,
  filterActions,
  pagination,
  isLoading,
  showHealth = false,
}: IInstallComponentsTable) => {
  return (
    <Table<InstallComponentRow>
      columns={buildColumns(showHealth)}
      data={data}
      isLoading={isLoading}
      filterActions={filterActions}
      emptyMessage="No components found"
      pagination={pagination}
      searchPlaceholder="Search by name or ID..."
    />
  )
}
