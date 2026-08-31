import { api } from '@/lib/api'
import type { TAppBranchConfig } from '@/types'

export type TUpdateBranchConfigRequest = {
  // Send '' to clear; send '.*' to ignore all GitHub events.
  ignore_changes_regex?: string
  send_statuses_on_ignore?: boolean
}

export const updateBranchConfig = ({
  appId,
  branchId,
  configId,
  orgId,
  request,
}: {
  appId: string
  branchId: string
  configId: string
  orgId: string
  request: TUpdateBranchConfigRequest
}) =>
  api<TAppBranchConfig>({
    path: `apps/${appId}/branches/${branchId}/configs/${configId}`,
    orgId,
    method: 'PATCH',
    body: request,
  })
