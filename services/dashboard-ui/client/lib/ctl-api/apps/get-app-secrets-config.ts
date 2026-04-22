import { api } from '@/lib/api'
import type { TAppSecretsConfig } from '@/types'

export const getAppSecretsConfig = ({
  appId,
  orgId,
}: {
  orgId: string
  appId: string
}) =>
  api<TAppSecretsConfig>({
    path: `apps/${appId}/latest-secrets-config`,
    orgId,
  })
