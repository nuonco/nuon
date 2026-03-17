import { api } from '@/lib/api'
import type { TAppSandboxBuild } from '@/types'

export const createSandboxBuild = ({
  appId,
  orgId,
}: {
  appId: string
  orgId: string
}) =>
  api<TAppSandboxBuild>({
    path: `apps/${appId}/sandbox/builds`,
    method: 'POST',
    orgId,
    body: {},
  })
