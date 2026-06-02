import { api } from '@/lib/api'
import type { TAppSandboxConfig, TCreateAppSandboxConfigBody } from '@/types'

export const createAppSandboxConfig = ({
  appId,
  orgId,
  body,
}: {
  appId: string
  orgId: string
  body: TCreateAppSandboxConfigBody
}) =>
  api<TAppSandboxConfig>({
    path: `apps/${appId}/sandbox/configs/v2`,
    method: 'POST',
    orgId,
    body,
  })
