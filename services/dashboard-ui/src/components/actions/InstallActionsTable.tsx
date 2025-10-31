'use client'

import { useSearchParams } from 'next/navigation'
import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { StatusWithDescription } from '@/components/common/StatusWithDescription'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { type IPagination } from '@/components/common/Pagination'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useQueryParams } from '@/hooks/use-query-params'
import type { TActionConfigTriggerType, TInstallAction } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'
import { ActionTriggerType } from './ActionTriggerType'
import { TriggeredByFilter } from './TriggeredByFilter'

// NOTE: old stuff
import { InstallActionTriggerFilter } from '@/components/old/InstallActionTriggerFilter'

export type InstallActionRow = {
  actionId: string
  actionName: string
  actionStatus: ReactNode
  actionTrigger: ReactNode
  actionRunDatetime: ReactNode
  actionRunDuration: ReactNode
  href: string
}

function parseInstallActionsLatestRunsToTableData(
  actionsWithRuns: TInstallAction[],
  orgId: string,
  installId: string
): InstallActionRow[] {
  return actionsWithRuns.map((actionWithRuns) => {
    const basePath = `/${orgId}/installs/${installId}`
    const recentRun = actionWithRuns?.runs?.at(0)

    return {
      actionId: actionWithRuns.action_workflow_id,
      actionName: actionWithRuns.action_workflow?.name,
      actionRunDatetime: recentRun ? (
        <Text className="!flex items-center gap-2">
          <Icon variant="CalendarBlankIcon" />
          <Time
            time={recentRun?.created_at}
            format="relative"
            variant="subtext"
          />
        </Text>
      ) : (
        <Icon variant="Minus" />
      ),
      actionRunDuration: recentRun ? (
        <Text className="!flex items-center gap-2">
          <Icon variant="TimerIcon" />
          <Duration nanoseconds={recentRun?.execution_time} variant="subtext" />
        </Text>
      ) : (
        <Icon variant="Minus" />
      ),
      actionStatus: recentRun ? (
        <StatusWithDescription
          statusProps={{ status: recentRun?.status_v2?.status }}
          tooltipProps={{
            position: 'top',
            tipContent: toSentenceCase(
              recentRun?.status_v2?.status_human_description
            ),
          }}
        />
      ) : (
        <Icon variant="Minus" />
      ),
      actionTrigger: recentRun ? (
        <ActionTriggerType
          componentName={recentRun?.run_env_vars?.COMPONENT_NAME}
          componentPath={`${basePath}/components/${recentRun?.run_env_vars?.COMPONENT_ID}`}
          triggerType={recentRun?.triggered_by_type as TActionConfigTriggerType}
        />
      ) : (
        <Icon variant="Minus" />
      ),
      href: `${basePath}/actions/${actionWithRuns.action_workflow_id}`,
    }
  })
}

const columns: ColumnDef<InstallActionRow>[] = [
  {
    accessorKey: 'actionName',
    header: 'Action name',
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
    accessorKey: 'actionRunDatetime',
    header: 'Time since last run',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    enableSorting: false,
    accessorKey: 'actionRunDuration',
    header: 'Run duration',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    accessorKey: 'actionTrigger',
    header: 'Recent trigger',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    accessorKey: 'actionStatus',
    header: 'Run status',
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

export const InstallActionsTable = ({
  actionsWithRuns: initActionsWithRuns,
  pagination,
  pollInterval = 20000,
  shouldPoll,
}: {
  actionsWithRuns: TInstallAction[]
  pagination: IPagination
} & IPollingProps) => {
  const searchParams = useSearchParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const queryParams = useQueryParams({
    offset: pagination?.offset,
    limit: pagination?.limit,
    q: searchParams.get('q'),
    trigger_types: searchParams.get('trigger_types'),
  })
  const { data: actions } = usePolling({
    dependencies: [queryParams],
    initData: initActionsWithRuns,
    path: `/api/orgs/${org.id}/installs/${install.id}/actions${queryParams}`,
    pollInterval,
    shouldPoll,
  })
  return (
    <Table<InstallActionRow>
      columns={columns}
      data={parseInstallActionsLatestRunsToTableData(
        actions,
        org.id,
        install.id
      )}
      filterActions={<TriggeredByFilter />}
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

export const InstallActionsTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
