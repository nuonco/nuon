'use client'

import { useSearchParams } from 'next/navigation'
import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { ComponentsGraph } from '@/components/apps/ConfigGraph'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { type IPagination } from '@/components/common/Pagination'
import { ComponentDependencies } from '@/components/components/ComponentDependencies'
import { ComponentType } from '@/components/components/ComponentType'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useQueryParams } from '@/hooks/use-query-params'
import type { TComponent } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

// NOTE: old stuff
import { ComponentTypeFilterDropdown } from '@/components/old/Components/NewComponentTypeFilter'

export type TComponentRow = {
  buildStatus: ReactNode
  componentId: string
  componentName: string
  componentType: ReactNode
  href: string
  dependencies: ReactNode
}

function parseComponentToTableData(
  components: TComponent[],
  orgId: string,
  appId: string
): TComponentRow[] {
  return components.map((component) => {
    return {
      componentId: component.id,
      componentName: component.name,
      componentType: <ComponentType type={component.type} variant="subtext" />,
      buildStatus: (
        <Tooltip
          position="top"
          tipContent={
            toSentenceCase(component.status_description) || 'Status unknown'
          }
        >
          <Status variant="badge" status={component.status} />
        </Tooltip>
      ),

      dependencies: component?.dependencies?.length ? (
        <ComponentDependencies deps={component?.dependencies} />
      ) : (
        <Icon variant="MinusIcon" />
      ),
      href: `/${orgId}/apps/${appId}/components/${component.id}`,
    }
  })
}

const columns: ColumnDef<TComponentRow>[] = [
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
    accessorKey: 'buildStatus',
    header: 'Latest build',
    cell: (info) => (
      <Text className="flex">{info.getValue() as ReactNode}</Text>
    ),
  },
  {
    enableSorting: false,
    accessorKey: 'href',
    id: 'href',
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

export const ComponentsTable = ({
  components: initComponents,
  pagination,
  pollInterval = 20000,
  shouldPoll,
}: {
  components: TComponent[]
  pagination: IPagination
} & IPollingProps) => {
  const searchParams = useSearchParams()
  const { org } = useOrg()
  const { app } = useApp()
  const queryParams = useQueryParams({
    offset: pagination?.offset,
    limit: pagination?.limit,
    q: searchParams.get('q'),
    types: searchParams.get('types'),
  })
  const { data: components } = usePolling({
    dependencies: [queryParams],
    initData: initComponents,
    path: `/api/orgs/${org.id}/apps/${app.id}/components${queryParams}`,
    pollInterval,
    shouldPoll,
  })

  return (
    <Table<TComponentRow>
      columns={columns}
      data={parseComponentToTableData(components, org.id, app.id)}
      filterActions={
        <div className="flex items-center gap-3">
          <ComponentsGraph
            appId={app?.id}
            configId={app?.app_configs?.at(-1)?.id}
          />
          <ComponentTypeFilterDropdown />
        </div>
      }
      emptyMessage="No components found"
      pagination={pagination}
      searchPlaceholder="Search component name..."
    />
  )
}

export const ComponentsTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
