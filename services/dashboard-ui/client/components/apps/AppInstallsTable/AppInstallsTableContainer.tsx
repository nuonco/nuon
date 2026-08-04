import type { ReactNode } from 'react'
import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { useOrg } from '@/hooks/use-org'
import { getAppInstalls } from '@/lib'
import { CreateInstallButton } from '../CreateInstall'
import { AppInstallsTable, parseInstallsToTableData } from './AppInstallsTable'

const LIMIT = 20

export const AppInstallsTableContainer = ({
  appId,
  filterBranchId,
  emptyTitle,
  emptyMessage,
  emptyAction,
  pollInterval = 20000,
  shouldPoll = true,
}: {
  appId?: string
  filterBranchId?: string
  emptyTitle?: string
  emptyMessage?: string
  emptyAction?: ReactNode
  pollInterval?: number
  shouldPoll?: boolean
}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const hasNewAppIA = useNewAppIA()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: [
      'app-installs',
      org.id,
      appId,
      offset,
      searchParams.get('q'),
      filterBranchId,
    ],
    queryFn: () => getAppInstalls({
      orgId: org.id,
      appId,
      offset,
      limit: LIMIT,
      q: searchParams.get('q') || undefined,
      app_branch_id: filterBranchId,
    }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!appId,
  })

  return (
    <AppInstallsTable
      data={parseInstallsToTableData(result?.data ?? [], org.id, appId)}
      isLoading={isLoading}
      emptyTitle={emptyTitle}
      emptyMessage={emptyMessage}
      emptyAction={emptyAction ?? <CreateInstallButton />}
      pagination={{ hasNext: result?.pagination?.hasNext ?? false, offset, limit: LIMIT }}
      showBranchColumn={hasNewAppIA && !filterBranchId}
    />
  )
}
