import { api } from '@/lib/api'
import type { TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'
import type { TRunbookRun } from './get-install-runbook-run'

export async function getInstallRunbookRuns({
  installId,
  orgId,
  limit,
  offset,
}: {
  installId: string
  orgId: string
} & TPaginationParams) {
  return api<TRunbookRun[]>({
    orgId,
    path: `installs/${installId}/runbook-runs${buildQueryParams({ limit, offset })}`,
    paginated: true,
  })
}
