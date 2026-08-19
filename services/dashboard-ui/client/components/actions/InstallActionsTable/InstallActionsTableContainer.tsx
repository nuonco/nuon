import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { LabelFilterDropdown } from '@/components/common/LabelFilterDropdown'
import { SyncedFilterContainer } from '@/components/common/SyncedFilter'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSyncedOnlyFilter } from '@/hooks/use-synced-only-filter'
import { getInstallActionsLatestRuns, getActionLabelKeys } from '@/lib'
import { TriggeredByFilter } from '../TriggeredByFilter'
import {
  InstallActionsTable,
  parseInstallActionsLatestRunsToTableData,
} from './InstallActionsTable'

const LIMIT = 20

export const InstallActionsTableContainer = ({
  pollInterval = 20000,
  shouldPoll,
}: {
  pollInterval?: number
  shouldPoll?: boolean
}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const { install, labelColors } = useInstall()
  const offset = Number(searchParams.get('offset') ?? 0)
  const q = searchParams.get('q') || undefined
  const trigger_types = searchParams.get('trigger_types') || undefined
  const labels = searchParams.get('labels') || undefined
  const { syncedOnly } = useSyncedOnlyFilter()

  const { data: result, isLoading } = useQuery({
    queryKey: ['install-actions', org?.id, install?.id, offset, q, trigger_types, labels],
    queryFn: () =>
      getInstallActionsLatestRuns({
        orgId: org.id,
        installId: install.id,
        limit: LIMIT,
        offset,
        q,
        trigger_types,
        labels,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id,
  })

  const { data: removedResult } = useQuery({
    queryKey: ['install-actions-removed', org?.id, install?.id, q, trigger_types, labels],
    queryFn: () =>
      getInstallActionsLatestRuns({
        orgId: org.id,
        installId: install.id,
        limit: 100,
        offset: 0,
        q,
        trigger_types,
        labels,
        synced: false,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id && !syncedOnly,
  })

  const actions = result?.data ?? []
  const removedActions =
    offset === 0 && !syncedOnly ? removedResult?.data ?? [] : []
  const pagination = { hasNext: result?.pagination?.hasNext ?? false, offset, limit: LIMIT }

  const removedRows = parseInstallActionsLatestRunsToTableData(
    removedActions,
    org?.id ?? '',
    install?.id ?? '',
    labelColors,
    true
  )
  const currentRows = parseInstallActionsLatestRunsToTableData(
    actions,
    org?.id ?? '',
    install?.id ?? '',
    labelColors
  )

  return (
    <InstallActionsTable
      isLoading={isLoading}
      data={[...removedRows, ...currentRows]}
      filterActions={
        <div className="flex items-center gap-4 flex-wrap">
          <AdminDashboardLink path={`/queues?owner_id=${install.id}`} label="View queues" />
          <SyncedFilterContainer />
          <LabelFilterDropdown
            queryKey={['action-label-keys', org.id, install?.app_id]}
            queryFn={() => getActionLabelKeys({ orgId: org.id, appId: install.app_id })}
          />
          <TriggeredByFilter />
        </div>
      }
      pagination={pagination}
    />
  )
}
