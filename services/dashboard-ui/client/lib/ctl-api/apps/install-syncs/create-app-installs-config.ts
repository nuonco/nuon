import { api } from '@/lib/api'
import type { TAppInstallsConfig } from '@/types'

export type TCreateAppInstallsConfigBody = {
  vcs_type: 'connected' | 'public'
  vcs_connection_id?: string
  repo: string
  branch: string
  directory?: string
}

export const createAppInstallsConfig = ({
  appId,
  body,
  orgId,
}: {
  appId: string
  body: TCreateAppInstallsConfigBody
  orgId: string
}) =>
  api<TAppInstallsConfig>({
    path: `apps/${appId}/installs-configs`,
    method: 'POST',
    body,
    orgId,
  })
