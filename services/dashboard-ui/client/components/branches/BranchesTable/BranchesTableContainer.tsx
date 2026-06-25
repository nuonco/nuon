import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppBranches } from '@/lib'
import { BranchesTable, parseBranchesToTableData } from './BranchesTable'

const LIMIT = 20

export const BranchesTableContainer = ({
  pollInterval = 20000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const { org } = useOrg()
  const { app } = useApp()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: ['app-branches', org.id, app.id, offset],
    queryFn: () =>
      getAppBranches({ orgId: org.id!, appId: app.id!, limit: LIMIT, offset }),
    enabled: !!org.id && !!app.id,
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  return (
    <BranchesTable
      data={parseBranchesToTableData(result?.data ?? [], org.id!, app.id!)}
      isLoading={isLoading}
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
    />
  )
}
