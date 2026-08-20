import { useSearchParams } from 'react-router'
import { keepPreviousData, useQueries, useQuery } from '@tanstack/react-query'
import { LabelFilterDropdown } from '@/components/common/LabelFilterDropdown'
import { useOrg } from '@/hooks/use-org'
import { getInstalls, getInstallLabelKeys, getAppLabels, toLabelColorMap } from '@/lib'
import { CreateInstallButton } from '../CreateInstall'
import { InstallBranchFilter } from '../InstallBranchFilter'
import { InstallsTable, parseInstallsToTableData } from './InstallsTable'

const LIMIT = 20

export const InstallsTableContainer = ({
  pollInterval = 20000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: [
      'installs',
      org.id,
      offset,
      searchParams.get('q'),
      searchParams.get('labels'),
      searchParams.get('branch_status'),
    ],
    queryFn: () =>
      getInstalls({
        orgId: org.id,
        offset,
        limit: LIMIT,
        q: searchParams.get('q') || undefined,
        labels: searchParams.get('labels') || undefined,
        branch_status: searchParams.get('branch_status') || undefined,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  const installs = result?.data ?? []
  const appIds = [...new Set(installs.map((i) => i.app_id).filter(Boolean))] as string[]

  const labelQueries = useQueries({
    queries: appIds.map((appId) => ({
      queryKey: ['app-labels', org.id, appId],
      queryFn: () => getAppLabels({ orgId: org.id, appId }),
      enabled: !!org.id && !!appId,
    })),
  })

  const labelColorsByApp: Record<string, Record<string, string>> = {}
  appIds.forEach((appId, i) => {
    labelColorsByApp[appId] = toLabelColorMap(labelQueries[i]?.data)
  })

  return (
    <InstallsTable
      data={parseInstallsToTableData(installs, org.id, labelColorsByApp)}
      isLoading={isLoading}
      emptyStateAction={<CreateInstallButton />}
      filterActions={
        <div className="flex items-center gap-3">
          <LabelFilterDropdown
            queryKey={['install-label-keys', org.id]}
            queryFn={() => getInstallLabelKeys({ orgId: org.id })}
          />
          <InstallBranchFilter />
          <CreateInstallButton
            className="!w-full !flex !justify-center md:!w-fit"
            variant="primary"
          />
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
