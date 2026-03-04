import { api } from '@/lib/api'
import type { TDeploy } from '@/types'

export const getDeploy = ({
  installId,
  deployId,
  orgId,
}: {
  installId: string
  deployId: string
  orgId: string
}) =>
  api<TDeploy>({
    path: `installs/${installId}/deploys/${deployId}`,
    orgId,
  })
