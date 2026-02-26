import { useSearchParams } from 'react-router-dom'
import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { useQuery } from '@tanstack/react-query'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { type IPagination } from '@/components/common/Pagination'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useQueryParams } from '@/hooks/use-query-params'
import { getActions } from '@/lib'
import type { TActionConfigTriggerType, TAction } from '@/types'
import { ActionTriggerType } from './ActionTriggerType'

export type TActionRow = {
  actionId: string
  actionName: string
  actionTriggers: ReactNode
  actionSteps: ReactNode
  href: string
}

function parseActionsToTableData(
  actions: TAction[],
  orgId: string,
  appId: string
): TActionRow[] {
  return actions.map((action) => {
    const basePath = `/${orgId}/apps/${appId}`

    return {
      actionId: action?.id,
      actionName: action?.name,
      actionSteps: (
        <ol className="flex flex-col gap-1 list-decimal">
          {action?.configs
            ?.at(-1)
            ?.steps?.sort((a, b) => b?.idx - a?.idx)
            ?.reverse()
            ?.map((s) => (
              <li key={s?.id} className="text-sm">
                <Text variant="subtext">{s?.name}</Text>
              </li>
            ))}
        </ol>
      ),
      actionTriggers: (
        <div className="flex flex-wrap gap-2">
          {action?.configs?.at(-1)?.triggers?.map((trigger) => (
            <ActionTriggerType
              key={trigger?.id}
              componentName={trigger?.component?.name}
              componentPath={`${basePath}/components/${trigger?.component?.id}`}
              triggerType={trigger?.type as TActionConfigTriggerType}
            />
          ))}
        </div>
      ),
      href: `${basePath}/actions/${action.id}`,
    }
  })
}

const columns: ColumnDef<TActionRow>[] = [
  {
    accessorKey: 'actionName',
    header: 'Action',
    cell: (info) => (
      <span>
        <Text variant="body">
          <Link href={info.row.original.href}>{info.getValue() as string}</Link>
        </Text>
        <ID>{info.row.original.actionId as string}</ID>
      </span>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'actionTriggers',
    header: 'Triggers',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    accessorKey: 'actionSteps',
    header: 'Steps',
    cell: (info) => info.getValue() as ReactNode,
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

export const ActionsTable = ({
  pagination,
  pollInterval = 20000,
  shouldPoll,
}: {
  pagination: IPagination
  pollInterval?: number
  shouldPoll?: boolean
}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const { app } = useApp()
  const queryParams = useQueryParams({
    offset: pagination?.offset,
    limit: pagination?.limit,
    q: searchParams.get('q'),
  })
  const { data: actions = [] } = useQuery({
    queryKey: ['actions', org?.id, app?.id, queryParams],
    queryFn: () =>
      getActions({
        orgId: org.id,
        appId: app.id,
        limit: pagination?.limit,
        offset: pagination?.offset,
        q: searchParams.get('q') || undefined,
      }).then((r) => r.data ?? []),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!app?.id,
  })

  return (
    <Table<TActionRow>
      columns={columns}
      data={parseActionsToTableData(actions, org.id, app.id)}
      emptyStateProps={{
        emptyMessage:
          'Save time by configuring your actions. Check out our resources.',
        emptyTitle: 'No actions yet',
        action: (
          <Link href="https://docs.nuon.co/concepts/actions" isExternal>
            Learn more <Icon size="14" variant="ArrowSquareOutIcon" />
          </Link>
        ),
      }}
      pagination={pagination}
      searchPlaceholder="Search component name..."
    />
  )
}

export const ActionsTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
