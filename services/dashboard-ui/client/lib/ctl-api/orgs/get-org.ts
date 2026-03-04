import { api } from '@/lib/api'
import type { TOrg } from '@/types'

export const getOrg = ({ orgId }: { orgId: string }) =>
  api<TOrg>({
    path: `orgs/current`,
    orgId,
  })
