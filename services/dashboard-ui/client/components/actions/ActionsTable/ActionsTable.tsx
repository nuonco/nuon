import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import type { TActionConfigTriggerType, TAction } from '@/types'
import { ActionTriggerType } from '../ActionTriggerType'

export type TActionRow = {
  actionId: string
  actionName: string
  labels: ReactNode
  actionTriggers: ReactNode
  actionSteps: ReactNode
  href: string
}

export function parseActionsToTableData(
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
      labels: (() => {
        const lbls = action.labels
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
      actionTriggers: (
        <div className="flex flex-wrap gap-2">
          {action?.configs?.at(-1)?.triggers?.map((trigger) => (
            <ActionTriggerType
              key={trigger?.id}
              componentName={trigger?.component?.name}
              componentPath={`${basePath}/components/${trigger?.component?.id}`}
              triggerType={trigger?.type as TActionConfigTriggerType}
              cronSchedule={trigger?.cron_schedule}
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
    enableSorting: false,
    accessorKey: 'labels',
    header: 'Labels',
    cell: (info) => info.getValue() as ReactNode,
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

interface IActionsTable {
  data: TActionRow[]
  isLoading: boolean
  filterActions?: ReactNode
  pagination: { hasNext?: boolean; offset: number; limit: number }
}

export const ActionsTable = ({
  data,
  isLoading,
  filterActions,
  pagination,
}: IActionsTable) => {
  if (isLoading) {
    return <ActionsTableSkeleton />
  }

  return (
    <Table<TActionRow>
      columns={columns}
      data={data}
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
      filterActions={filterActions}
      pagination={pagination}
      searchPlaceholder="Search action name..."
    />
  )
}

export const ActionsTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
