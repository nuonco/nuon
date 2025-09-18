import { api } from '@/lib/api'
import type { TSandboxRun, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstallSandboxRuns = ({
  installId,
  orgId,
  limit,
  offset,
}: {
  installId: string
  orgId: string
} & TPaginationParams) =>
  api<TSandboxRun[]>({
    path: `installs/${installId}/sandbox-runs${buildQueryParams({ limit, offset })}`,
    orgId,
  })
