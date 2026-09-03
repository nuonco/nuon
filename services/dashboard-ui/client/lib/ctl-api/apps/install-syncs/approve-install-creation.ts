import { api } from '@/lib/api'

export const respondInstallCreationApproval = ({
  appId,
  syncId,
  approvalId,
  responseType,
  orgId,
}: {
  appId: string
  syncId: string
  approvalId: string
  responseType: 'approve' | 'deny'
  orgId: string
}) =>
  api<{ status: string }>({
    path: `apps/${appId}/install-syncs/${syncId}/approvals/${approvalId}/response`,
    method: 'POST',
    body: { response_type: responseType },
    orgId,
  })

export const approveInstallCreation = ({
  appId,
  syncId,
  approvalId,
  orgId,
}: {
  appId: string
  syncId: string
  approvalId: string
  orgId: string
}) =>
  respondInstallCreationApproval({
    appId,
    syncId,
    approvalId,
    responseType: 'approve',
    orgId,
  })
