import { api } from '@/lib/api'
import type { TAirgapInstall, TCreateAirgapInstallRequest } from '@/types'

export const createAirgapInstall = ({
  appId,
  body,
  bundleId,
  orgId,
}: {
  appId: string
  body: TCreateAirgapInstallRequest
  bundleId: string
  orgId: string
}) =>
  api<TAirgapInstall>({
    body,
    method: 'POST',
    orgId,
    path: `apps/${appId}/airgap-bundles/${bundleId}/installs`,
  })
