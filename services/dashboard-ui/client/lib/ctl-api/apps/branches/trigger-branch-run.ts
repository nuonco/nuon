import { api } from '@/lib/api'
import type { TInstallWorkflow, TPreviewRunRequest } from '@/types'

export type TTriggerBranchRunRequest = {
  config_id?: string
  force?: boolean
  plan_only?: boolean
  preview_run?: TPreviewRunRequest
}

export const triggerBranchRun = ({
  appId,
  branchId,
  orgId,
  request = {},
}: {
  appId: string
  branchId: string
  orgId: string
  request?: TTriggerBranchRunRequest
}) =>
  api<TInstallWorkflow>({
    path: `apps/${appId}/branches/${branchId}/runs`,
    orgId,
    method: 'POST',
    body: request,
  })
