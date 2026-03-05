import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
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

const LIMIT = 20

export const ActionsTable = ({
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
    queryKey: ['actions', org?.id, app?.id, offset, searchParams.get('q')],
    queryFn: () =>
      getActions({
        orgId: org.id,
        appId: app.id,
        offset,
        limit: LIMIT,
        q: searchParams.get('q') || undefined,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!app?.id,
  })

  if (isLoading) {
    return <ActionsTableSkeleton />
  }

  const actions = result?.data ?? []

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
      pagination={{ hasNext: result?.pagination?.hasNext ?? false, offset, limit: LIMIT }}
      searchPlaceholder="Search action name..."
    />
  )
}

export const ActionsTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
