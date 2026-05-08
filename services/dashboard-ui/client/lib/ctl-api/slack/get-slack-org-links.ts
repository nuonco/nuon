import { api } from '@/lib/api'
import type { TSlackOrgLink } from '@/types'

export const getSlackOrgLinks = ({ orgId }: { orgId: string }) =>
  api<TSlackOrgLink[]>({
    orgId,
    path: `orgs/${orgId}/slack/org-links`,
  })
