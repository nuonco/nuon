'use client'

import { useSearchParams } from 'next/navigation'
import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import { type IPagination } from '@/components/common/Pagination'
import { InstallComponentDependencies } from '@/components/install-components/InstallComponentDependencies'
import { ComponentType } from '@/components/components/ComponentType'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useQueryParams } from '@/hooks/use-query-params'
import type { TInstallComponent } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'
import { ManageAllDropdown } from '@/components/install-components/management/ManageAllDropdown'

// NOTE: old stuff
import { ComponentTypeFilterDropdown } from '@/components/old/Components/NewComponentTypeFilter'

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
  dependencies: ReactNode
}

function parseInstallComponentSummaryToTableData(
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
        <ComponentType type={component?.component?.type} variant="subtext" />
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
                  <Text className="!flex px-2 pb-2" variant="subtext">
                    {toSentenceCase(
                      component?.status_v2?.status_human_description
                    )}
                  </Text>
                </>
              ) : (
                <Text className="!flex p-2" variant="subtext">
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
      href: `/${orgId}/installs/${installId}/components/${component.component_id}`,
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
      <Text className="!flex">{info.getValue() as ReactNode}</Text>
    ),
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
      <Text className="!flex">{info.getValue() as ReactNode}</Text>
    ),
  },
  {
    enableSorting: false,
    accessorKey: 'href',
    id: 'action',
    header: '',
    cell: (info) => (
      <Text>
        <Link className="text-left" href={info.getValue() as string}>
          View <Icon variant="CaretRightIcon" />
        </Link>
      </Text>
    ),
  },
]

export const InstallComponentsTable = ({
  components: initComponents,
  deps,
  pagination,
  pollInterval = 20000,
  shouldPoll,
}: {
  components: TInstallComponent[]
  deps: TComponentDeps[]
  pagination: IPagination
} & IPollingProps) => {
  const searchParams = useSearchParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const queryParams = useQueryParams({
    offset: pagination?.offset,
    limit: pagination?.limit,
    q: searchParams.get('q'),
    types: searchParams.get('types'),
  })
  const { data: components } = usePolling({
    initData: initComponents,
    path: `/api/orgs/${org.id}/installs/${install.id}/components${queryParams}`,
    pollInterval,
    shouldPoll,
  })
  return (
    <Table<InstallComponentRow>
      columns={columns}
      data={parseInstallComponentSummaryToTableData(
        components,
        deps,
        org.id,
        install.id
      )}
      filterActions={
        <div className="flex items-center gap-3">
          <ComponentTypeFilterDropdown />
          <ManageAllDropdown />
        </div>
      }
      emptyMessage="No components found"
      pagination={pagination}
      searchPlaceholder="Search component name..."
    />
  )
}

export const InstallComponentsTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
