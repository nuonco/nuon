import { api } from '@/lib/api'
import type { TInstallOverview } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstallOverview = ({
  installId,
  orgId,
  includeAlignment = true,
}: {
  installId: string
  orgId: string
  includeAlignment?: boolean
}) =>
  api<TInstallOverview>({
    orgId,
    path: `installs/${installId}/overview${buildQueryParams({ include_alignment: includeAlignment })}`,
  })
