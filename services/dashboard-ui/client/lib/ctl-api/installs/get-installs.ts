import { api } from '@/lib/api'
import type { TInstall, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstalls = ({
  limit,
  offset,
  orgId,
  q,
  labels,
  runner_id,
}: { orgId: string; q?: string; labels?: string; runner_id?: string } & TPaginationParams) =>
  api<TInstall[]>({
    path: `installs${buildQueryParams({ limit, offset, q, labels, runner_id })}`,
    orgId,
    paginated: true,
  })
