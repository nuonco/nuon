import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { LabelFilterDropdown } from '@/components/common/LabelFilterDropdown'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getActions, getActionLabelKeys } from '@/lib'
import { TriggeredByFilter } from '../TriggeredByFilter'
import { ActionsTable, parseActionsToTableData } from './ActionsTable'

const LIMIT = 20

export const ActionsTableContainer = ({
  filterIds,
  branchId,
  pollInterval = 20000,
  shouldPoll = true,
}: {
  filterIds?: string[]
  branchId?: string
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const { app, labelColors } = useApp()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: [
      'actions',
      org?.id,
      app?.id,
      offset,
      searchParams.get('q'),
      searchParams.get('labels'),
      searchParams.get('trigger_types'),
    ],
    queryFn: () =>
      getActions({
        orgId: org.id,
        appId: app.id,
        offset,
        limit: LIMIT,
        q: searchParams.get('q') || undefined,
        labels: searchParams.get('labels') || undefined,
        trigger_types: searchParams.get('trigger_types') || undefined,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!app?.id,
  })

  return (
    <ActionsTable
      data={parseActionsToTableData(
        (result?.data ?? []).filter(
          (row) => !filterIds || filterIds.includes(row.id ?? '')
        ),
        org.id,
        app.id,
        labelColors,
        branchId
      )}
      isLoading={isLoading}
      filterActions={
        <div className="flex items-center gap-4 flex-wrap">
          <LabelFilterDropdown
            queryKey={['action-label-keys', org.id, app.id]}
            queryFn={() => getActionLabelKeys({ orgId: org.id, appId: app.id })}
          />
          <TriggeredByFilter />
        </div>
      }
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
    />
  )
}
