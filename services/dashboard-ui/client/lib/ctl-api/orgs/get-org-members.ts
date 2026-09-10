import { api } from '@/lib/api'
import type { TOrgMember, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getOrgMembers = ({
  orgId,
  q,
  status,
  role_type,
  limit,
  offset,
}: {
  orgId: string
  q?: string
  status?: string
  role_type?: string
} & TPaginationParams) =>
  api<TOrgMember[]>({
    path: `orgs/current/members${buildQueryParams({ limit, offset, q, status, role_type })}`,
    orgId,
    paginated: true,
  })
