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
  branches,
  cloud_connection_id,
  health,
  lifecycle_phase,
}: {
  orgId: string
  q?: string
  labels?: string
  runner_id?: string
  branches?: string
  cloud_connection_id?: string
  health?: string
  lifecycle_phase?: string
} & TPaginationParams) =>
  api<TInstall[]>({
    path: `installs${buildQueryParams({ limit, offset, q, labels, runner_id, branches, cloud_connection_id, health, lifecycle_phase, include_components: false })}`,
    orgId,
    paginated: true,
  })
