import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router'
import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import { getRunnerProcesses } from '@/lib'
import { RunnerProcessesTable } from './RunnerProcessesTable'

const LIMIT = 20

export const RunnerProcessesTableContainer = ({
  shouldPoll = true,
  pollInterval = 20000,
  filterStatus,
}: {
  shouldPoll?: boolean
  pollInterval?: number
  filterStatus?: string
}) => {
  const { org } = useOrg()
  const { runner } = useRunner()
  const { user, isLoading: isAuthLoading } = useAuth()
  const config = useConfig()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: ['runner-processes', org?.id, runner?.id, offset, filterStatus],
    queryFn: () =>
      getRunnerProcesses({
        orgId: org.id,
        runnerId: runner.id,
        limit: LIMIT,
        offset,
        status: filterStatus,
      }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!runner?.id,
  })

  const isAdminVisible = !isAuthLoading && !!user?.email?.endsWith('@nuon.co')
  const adminDashboardUrl = isAdminVisible ? (config.adminDashboardUrl ?? '') : undefined

  return (
    <RunnerProcessesTable
      processes={result?.data ?? []}
      isLoading={isLoading}
      adminDashboardUrl={adminDashboardUrl || undefined}
    />
  )
}
