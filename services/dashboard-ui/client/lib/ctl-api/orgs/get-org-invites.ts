import { api } from '@/lib/api'
import type { TOrgInvite, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getOrgInvites = ({
  orgId,
  limit,
  offset,
}: { orgId: string } & TPaginationParams) =>
  api<TOrgInvite[]>({
    path: `orgs/current/invites${buildQueryParams({ limit, offset })}`,
    orgId,
  })
