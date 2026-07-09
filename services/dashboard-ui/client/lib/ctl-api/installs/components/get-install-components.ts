import { api } from '@/lib/api'
import type { TInstallComponent, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export async function getInstallComponents({
  installId,
  labels,
  limit,
  orgId,
  offset,
  q,
  synced,
  types,
}: {
  installId: string
  labels?: string
  orgId: string
  q?: string
  synced?: boolean
  types?: string
} & TPaginationParams) {
  return api<TInstallComponent[]>({
    orgId,
    path: `installs/${installId}/components${buildQueryParams({ limit, offset, q, types, labels, synced })}`,
    paginated: true,
  })
}
