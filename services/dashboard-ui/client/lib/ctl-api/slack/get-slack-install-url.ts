import { api } from '@/lib/api'
import type { TSlackInstallURLResponse } from '@/types'

export const getSlackInstallURL = ({ orgId }: { orgId: string }) =>
  api<TSlackInstallURLResponse>({
    orgId,
    path: `orgs/${orgId}/slack/install-url`,
  })
