import { api } from '@/lib/api'
import type { TInstallComponent, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export async function getInstallComponents({
  installId,
  limit,
  orgId,
  offset,
  q,
  types,
}: {
  installId: string
  orgId: string
  q?: string
  types?: string
} & TPaginationParams) {
  return api<TInstallComponent[]>({
    orgId,
    path: `installs/${installId}/components${buildQueryParams({ limit, offset, q, types })}`,
  })
}
