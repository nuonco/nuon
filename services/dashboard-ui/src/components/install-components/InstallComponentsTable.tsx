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
import { Tooltip } from '@/components/common/Tooltip'
import { type IPagination } from '@/components/common/Pagination'
import { InstallComponentDependencies } from '@/components/install-components/InstallComponentDependencies'
import { ComponentType } from '@/components/components/ComponentType'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useQueryParams } from '@/hooks/use-query-params'
import type { TInstallComponentSummary } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

export type InstallComponentRow = {
  buildStatus: ReactNode
  componentId: string
  componentName: string
  componentType: ReactNode
  deployStatus: ReactNode
  href: string
  dependencies: ReactNode
}

function parseInstallComponentSummaryToTableData(
  components: TInstallComponentSummary[],
  orgId: string,
  installId: string,
  appId: string
): InstallComponentRow[] {
  return components.map((component) => {
    //    console.log("build installcomponet deps", component);
    return {
      buildStatus: (
        <Tooltip
          position="top"
          tipContent={
            <div className="flex flex-col w-max">
              <Text>
                {toSentenceCase(component.build_status_description) ||
                  'Status unknown'}
              </Text>
              <Link
                href={`/${orgId}/apps/${appId}/components/${component.id}/builds`}
              >
                View details <Icon variant="CaretRight" />
              </Link>
            </div>
          }
        >
          <Status variant="badge" status={component.build_status} />
        </Tooltip>
      ),
      componentId: component.component_id,
      componentName: component.component_name,
      componentType: (
        <ComponentType
          type={component?.component_config?.type}
          variant="subtext"
        />
      ),
      deployStatus: (
        <Tooltip
          position="top"
          tipContent={
            toSentenceCase(component.deploy_status_description) ||
            'Status unknown'
          }
        >
          <Status variant="badge" status={component.deploy_status} />
        </Tooltip>
      ),
      dependencies: (
        <InstallComponentDependencies deps={component.dependencies} />
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
        <Link href={info.row.original.href}>{info.getValue() as string}</Link>
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
    enableSorting: false,
    accessorKey: 'deployStatus',
    header: 'Latest deploy',
    cell: (info) => (
      <Text className="flex">{info.getValue() as ReactNode}</Text>
    ),
  },
  {
    enableSorting: true,
    accessorKey: 'dependencies',
    header: 'Dependencies',
    cell: (info) => (
      <Text className="flex">{info.getValue() as ReactNode}</Text>
    ),
  },
  {
    accessorKey: 'buildStatus',
    header: 'Latest build',
    cell: (info) => (
      <Text className="flex items-center gap-1">
        {info.getValue() as ReactNode}
      </Text>
    ),
    enableSorting: true,
  },
]

export const InstallComponentsTable = ({
  components: initComponents,
  pagination,
  pollInterval = 20000,
  shouldPoll,
}: {
  components: TInstallComponentSummary[]
  pagination: IPagination
} & IPollingProps) => {
  const searchParams = useSearchParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const queryParams = useQueryParams({
    offset: pagination?.offset,
    limit: pagination?.limit,
    q: searchParams.get('q'),
  })
  const { data: components } = usePolling({
    initData: initComponents,
    path: `/api/orgs/${org.id}/installs/${install.id}/components/summary${queryParams}`,
    pollInterval,
    shouldPoll,
  })
  return (
    <Table<InstallComponentRow>
      columns={columns}
      data={parseInstallComponentSummaryToTableData(
        components,
        org.id,
        install.id,
        install.app_id
      )}
      emptyMessage="No components found"
      pagination={pagination}
      searchPlaceholder="Search component name..."
    />
  )
}

export const InstallComponentsTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
