import { api } from '@/lib/api'

export const getAppConfigGraph = ({
  appId,
  appConfigId,
  orgId,
}: {
  orgId: string
  appId: string
  appConfigId: string
}) =>
  api<string>({
    path: `apps/${appId}/config/${appConfigId}/graph`,
    orgId,
  })
