import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { TemporalLink } from '@/components/admin/TemporalLink'
import { LabelFilterDropdown } from '@/components/common/LabelFilterDropdown'
import { RunAdhocActionButton } from '@/components/installs/management/RunAdhocAction'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
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
  const { install } = useInstall()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result } = useQuery({
    queryKey: [
      'install-actions',
      org?.id,
      install?.id,
      offset,
      searchParams.get('q'),
      searchParams.get('trigger_types'),
      searchParams.get('labels'),
    ],
    queryFn: () =>
      getInstallActionsLatestRuns({
        orgId: org.id,
        installId: install.id,
        limit: LIMIT,
        offset,
        q: searchParams.get('q') || undefined,
        trigger_types: searchParams.get('trigger_types') || undefined,
        labels: searchParams.get('labels') || undefined,
      }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id,
  })

  const actions = result?.data ?? []
  const pagination = { hasNext: result?.pagination?.hasNext ?? false, offset, limit: LIMIT }

  return (
    <InstallActionsTable
      data={parseInstallActionsLatestRunsToTableData(
        actions,
        org?.id ?? '',
        install?.id ?? ''
      )}
      filterActions={
        <div className="flex items-center gap-4 flex-wrap">
          <LabelFilterDropdown
            queryKey={['action-label-keys', org.id, install?.app_id]}
            queryFn={() => getActionLabelKeys({ orgId: org.id, appId: install.app_id })}
          />
          <TemporalLink
            namespace="installs"
            eventLoopId={`${install?.id}-action-workflows`}
          />
          <RunAdhocActionButton />
          <TriggeredByFilter />
        </div>
      }
      pagination={pagination}
    />
  )
}
