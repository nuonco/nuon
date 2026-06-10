import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import { ComponentType } from '@/components/components/ComponentType'
import { InstallComponentDependencies } from '@/components/install-components/InstallComponentDependencies'
import { QuickComponentManagementDropdown } from '@/components/install-components/management/QuickComponentManagementDropdown'
import type { TInstallComponent } from '@/types'
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
  deployStatus: ReactNode
  driftStatus: ReactNode
  href: string
  action: ReactNode
  dependencies: ReactNode
  labels: ReactNode
}

export function parseInstallComponentSummaryToTableData(
  components: TInstallComponent[],
  deps: TComponentDeps[],
  orgId: string,
  installId: string
): InstallComponentRow[] {
  return components.map((component) => {
    const depIndex = deps?.findIndex((dep) => dep?.id === component?.id)

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
      dependencies: (
        <InstallComponentDependencies deps={deps?.at(depIndex)?.dependencies} />
      ),
      labels: (() => {
        const lbls = component.component?.labels
        if (!lbls || Object.keys(lbls).length === 0) return null
        return (
          <span className="flex flex-wrap gap-1">
            {Object.keys(lbls)
              .sort()
              .map((k) => (
                <LabelBadge key={k} labelKey={k} labelValue={lbls[k]} size="sm" />
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
          />
        </div>
      ),
    }
  })
}

const columns: ColumnDef<InstallComponentRow>[] = [
  {
    accessorKey: 'componentName',
    header: 'Component name',
    cell: (info) => (
      <span>
        <Text variant="body">
          <Link href={info.row.original.href}>{info.getValue() as string}</Link>
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
    accessorKey: 'action',
    id: 'action',
    header: '',
    cell: (info) => info.getValue<ReactNode>(),
  },
]

interface IInstallComponentsTable {
  data: InstallComponentRow[]
  filterActions: ReactNode
  pagination: {
    hasNext: boolean
    offset: number
    limit: number
  }
  isLoading: boolean
}

export const InstallComponentsTable = ({
  data,
  filterActions,
  pagination,
  isLoading,
}: IInstallComponentsTable) => {
  return (
    <Table<InstallComponentRow>
      columns={columns}
      data={data}
      isLoading={isLoading}
      filterActions={filterActions}
      emptyMessage="No components found"
      pagination={pagination}
      searchPlaceholder="Search by name or ID..."
    />
  )
}

export const InstallComponentsTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
