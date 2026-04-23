import { api } from '@/lib/api'

export const getComponentLabelKeys = ({
  appId,
  orgId,
}: {
  appId: string
  orgId: string
}) =>
  api<Record<string, string[]>>({
    path: `apps/${appId}/components/label-keys`,
    orgId,
  })
