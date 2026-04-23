import { api } from '@/lib/api'

export const getActionLabelKeys = ({
  appId,
  orgId,
}: {
  appId: string
  orgId: string
}) =>
  api<Record<string, string[]>>({
    path: `apps/${appId}/actions/label-keys`,
    orgId,
  })
