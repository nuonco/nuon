import { api } from '@/lib/api'
import type { TInstallComponent, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export async function getInstallComponents({
  installId,
  ids,
  labels,
  limit,
  orgId,
  offset,
  q,
  synced,
  types,
}: {
  installId: string
  ids?: string
  labels?: string
  orgId: string
  q?: string
  synced?: boolean
  types?: string
} & TPaginationParams) {
  return api<TInstallComponent[]>({
    orgId,
    path: `installs/${installId}/components${buildQueryParams({ limit, offset, q, types, labels, synced, ids })}`,
    paginated: true,
  })
}
