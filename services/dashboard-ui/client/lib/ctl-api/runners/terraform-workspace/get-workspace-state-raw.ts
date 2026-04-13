import { api } from '@/lib/api'

export const getWorkspaceStateRaw = ({
  workspaceId,
  stateId,
  orgId,
}: {
  workspaceId: string
  stateId: string
  orgId: string
}) =>
  api<any>({
    path: `terraform-workspaces/${workspaceId}/state-json/${stateId}/raw`,
    orgId,
  })
