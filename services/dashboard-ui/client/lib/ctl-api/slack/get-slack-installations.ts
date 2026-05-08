import { api } from '@/lib/api'
import type { TSlackInstallation } from '@/types'

export const getSlackInstallations = ({ orgId }: { orgId: string }) =>
  api<TSlackInstallation[]>({
    orgId,
    path: `orgs/${orgId}/slack/installations`,
  })
