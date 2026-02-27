import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
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
import { ComponentDependencies } from '@/components/components/ComponentDependencies'
import { ComponentType } from '@/components/components/ComponentType'
import { ComponentTypeFilterDropdown } from '@/components/components/ComponentTypeFilter'
import { ManagementDropdown } from '@/components/components/management/ManagementDropdown'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getComponents } from '@/lib'
import type { TComponent } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

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
                  <Text className="!flex px-2 pb-2" variant="subtext">
                    {toSentenceCase(
                      component?.latest_build?.status_v2
                        ?.status_human_description
                    )}
                  </Text>
                </>
              ) : (
                <Text className="!flex p-2 !text-nowrap" variant="subtext">
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

const LIMIT = 20

export const ComponentsTable = ({
  pollInterval = 20000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const { app } = useApp()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: ['components', org.id, app.id, offset, searchParams.get('q'), searchParams.get('types')],
    queryFn: () => getComponents({
      orgId: org.id,
      appId: app.id,
      offset,
      limit: LIMIT,
      q: searchParams.get('q') || undefined,
      types: searchParams.get('types') || undefined,
    }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  if (isLoading) {
    return <ComponentsTableSkeleton />
  }

  const components = result?.data ?? []

  return (
    <Table<TComponentRow>
      columns={columns}
      data={parseComponentToTableData(components, org.id, app.id)}
      filterActions={
        <div className="flex items-center gap-3">
          <ComponentTypeFilterDropdown />
          <ManagementDropdown />
        </div>
      }
      emptyStateProps={{
        emptyMessage: 'No components found or configured.',
        emptyTitle: 'No components',
      }}
      pagination={{ hasNext: result?.pagination?.hasNext ?? false, offset, limit: LIMIT }}
      searchPlaceholder="Search component name..."
    />
  )
}

export const ComponentsTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
