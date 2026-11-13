import { api } from '@/lib/api'
import type { TDiffNode, TDiffSummary } from '@/lib/ctl-api/apps/get-app-config-diff'

export type TInstallAppConfigTreeDiffResponse = {
  changed?: string
  config_id?: string
  diff?: TDiffNode
  summary?: TDiffSummary
}

export const getInstallAppConfigTreeDiff = ({
  installId,
  configId,
  orgId,
}: {
  installId: string
  configId: string
  orgId: string
}) =>
  api<TInstallAppConfigTreeDiffResponse>({
    orgId,
    path: `installs/${installId}/app-configs/${configId}/diff`,
  })
