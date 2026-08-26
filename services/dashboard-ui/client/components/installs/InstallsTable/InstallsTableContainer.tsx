import type { ReactNode } from 'react'
import { useSearchParams } from 'react-router'
import { keepPreviousData, useQueries, useQuery } from '@tanstack/react-query'
import { LabelFilterDropdown } from '@/components/common/LabelFilterDropdown'
import { useOrg } from '@/hooks/use-org'
import {
  getAppInstalls,
  getInstalls,
  getInstallLabelKeys,
  getBranches,
  getAppLabels,
  toLabelColorMap,
} from '@/lib'
import { CreateInstallButton } from '../CreateInstall'
import { InstallBranchFilter } from '../InstallBranchFilter'
import {
  InstallsTable,
  parseInstallsToTableData,
  type TInstallsTableScope,
} from './InstallsTable'

const LIMIT = 20

export const InstallsTableContainer = ({
  appId,
  branchId,
  emptyStateAction,
  emptyTitle,
  emptyMessage,
  pollInterval = 20000,
  shouldPoll = true,
}: {
  appId?: string
  branchId?: string
  emptyStateAction?: ReactNode
  emptyTitle?: string
  emptyMessage?: string
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const offset = Number(searchParams.get('offset') ?? 0)
  const q = searchParams.get('q') || undefined

  const scope: TInstallsTableScope = branchId
    ? 'branch'
    : appId
      ? 'app'
      : 'org'

  const { data: result, isLoading } = useQuery({
    queryKey: [
      'installs',
      org.id,
      appId,
      branchId,
      offset,
      q,
      searchParams.get('labels'),
      searchParams.get('branches'),
    ],
    queryFn: () =>
      appId
        ? getAppInstalls({
            orgId: org.id,
            appId,
            offset,
            limit: LIMIT,
            q,
            app_branch_id: branchId,
          })
        : getInstalls({
            orgId: org.id,
            offset,
            limit: LIMIT,
            q,
            labels: searchParams.get('labels') || undefined,
            branches: searchParams.get('branches') || undefined,
          }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  const installs = result?.data ?? []
  const appIds = [
    ...new Set(installs.map((i) => i.app_id).filter(Boolean)),
  ] as string[]

  const labelQueries = useQueries({
    queries: appIds.map((id) => ({
      queryKey: ['app-labels', org.id, id],
      queryFn: () => getAppLabels({ orgId: org.id, appId: id }),
      enabled: !!org.id && !!id,
    })),
  })

  const labelColorsByApp: Record<string, Record<string, string>> = {}
  appIds.forEach((id, i) => {
    labelColorsByApp[id] = toLabelColorMap(labelQueries[i]?.data)
  })

  return (
    <InstallsTable
      data={parseInstallsToTableData(installs, org.id, labelColorsByApp)}
      isLoading={isLoading}
      emptyStateAction={emptyStateAction ?? <CreateInstallButton />}
      emptyTitle={emptyTitle}
      emptyMessage={emptyMessage}
      filterActions={
        scope === 'org' ? (
          <div className="flex items-center gap-3">
            <LabelFilterDropdown
              queryKey={['install-label-keys', org.id]}
              queryFn={() => getInstallLabelKeys({ orgId: org.id })}
            />
            <InstallBranchFilter
              queryKey={['org-branch-names', org.id]}
              queryFn={async () => {
                const { data } = await getBranches({ orgId: org.id, limit: 100 })
                return [...new Set(data.map((b) => b.name).filter(Boolean))].sort()
              }}
            />
          </div>
        ) : undefined
      }
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
      scope={scope}
    />
  )
}
