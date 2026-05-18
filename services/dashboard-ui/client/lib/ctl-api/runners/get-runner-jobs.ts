import { api } from '@/lib/api'
import type { TRunnerJob, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

type TJobGroup = TRunnerJob['group']
type TJobStatus = TRunnerJob['status']
type TJobExecutor = 'org-runner' | 'control-plane'

export const getRunnerJobs = ({
  executor,
  group,
  groups,
  limit = 10,
  offset,
  orgId,
  runnerId,
  status,
  statuses,
}: {
  runnerId: string
  orgId: string
  group?: TJobGroup
  groups?: TJobGroup[]
  executor?: TJobExecutor
  status?: TJobStatus
  statuses?: TJobStatus[]
} & TPaginationParams) =>
  api<TRunnerJob[]>({
    path: `runners/${runnerId}/jobs${buildQueryParams({ limit, offset, executor, group, groups, status, statuses })}`,
    orgId,
    paginated: true,
  })
