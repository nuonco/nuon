import { api } from '@/lib/api'
import type { TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'
import type { TRunbook } from '@/lib/ctl-api/apps/runbooks'

export type TInstallRunbook = TRunbook & {
  install_id?: string
}

export async function getInstallRunbooks({
  installId,
  orgId,
  limit,
  offset,
}: {
  installId: string
  orgId: string
} & TPaginationParams) {
  return api<TInstallRunbook[]>({
    orgId,
    path: `installs/${installId}/runbooks${buildQueryParams({ limit, offset })}`,
    paginated: true,
  })
}
