import { api } from '@/lib/api'
import type { TOrgStats } from '@/types'

export const getOrgStats = ({ orgId }: { orgId: string }) =>
  api<TOrgStats>({
    path: `orgs/current/stats`,
    orgId,
  })
