import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import { ComponentDependencies } from '@/components/components/ComponentDependencies'
import { ComponentType } from '@/components/components/ComponentType'
import type { TComponent } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

export type TComponentRow = {
  buildStatus: ReactNode
  componentId: string
  componentName: string
  componentType: ReactNode
  href: string
  dependencies: ReactNode
  labels: ReactNode
}

export function parseComponentToTableData(
  components: TComponent[],
  orgId: string,
  appId: string
): TComponentRow[] {
  return components.map((component) => {
    return {
      componentId: component.id,
      componentName: component.name,
      componentType: (
        <ComponentType
          type={component.type}
          variant="subtext"
          colorVariant="color"
        />
      ),
      buildStatus: (
        <Tooltip
          position="top"
          tipContent={
            <div className="w-fit max-w-64">
              {component?.latest_build?.status_v2?.status ? (
                <>
                  <Time
                    className="!text-nowrap px-2 py-1"
                    variant="subtext"
                    seconds={component?.latest_build?.status_v2?.created_at_ts}
                    weight="strong"
                  />
                  <hr className="my-1" />
                  <Text as="div" className="flex px-2 pb-2" variant="subtext">
                    {toSentenceCase(
                      component?.latest_build?.status_v2
                        ?.status_human_description
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
          <Status
            variant="badge"
            status={component?.latest_build?.status_v2?.status}
          />
        </Tooltip>
      ),
      dependencies: component?.dependencies?.length ? (
        <ComponentDependencies deps={component?.dependencies} />
      ) : (
        <Icon variant="MinusIcon" />
      ),
      labels: (() => {
        const lbls = component.labels
        if (!lbls || Object.keys(lbls).length === 0) return null
        return (
          <span className="flex flex-wrap gap-1">
            {Object.keys(lbls)
              .sort()
              .map((k) => (
                <Badge key={k} variant="code" size="sm" theme="neutral">
                  {k}: {lbls[k]}
                </Badge>
              ))}
          </span>
        )
      })(),
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

interface IComponentsTable {
  data: TComponentRow[]
  isLoading: boolean
  filterActions?: ReactNode
  pagination: { hasNext?: boolean; offset: number; limit: number }
}

export const ComponentsTable = ({
  data,
  isLoading,
  filterActions,
  pagination,
}: IComponentsTable) => {
  return (
    <Table<TComponentRow>
      columns={columns}
      data={data}
      isLoading={isLoading}
      filterActions={filterActions}
      emptyStateProps={{
        emptyMessage: 'No components found or configured.',
        emptyTitle: 'No components',
      }}
      pagination={pagination}
      searchPlaceholder="Search by name or ID..."
    />
  )
}

export const ComponentsTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
