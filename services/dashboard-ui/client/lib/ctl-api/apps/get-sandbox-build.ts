import { api } from '@/lib/api'
import type { TAppSandboxBuild } from '@/types'

export const getSandboxBuild = ({
  appId,
  buildId,
  orgId,
}: {
  appId: string
  buildId: string
  orgId: string
}) =>
  api<TAppSandboxBuild>({
    path: `apps/${appId}/sandbox/builds/${buildId}`,
    orgId,
  })
