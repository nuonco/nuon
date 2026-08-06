import { api } from '@/lib/api'
import type { TAirgapInstall } from '@/types'

export async function getAirgapInstalls({
  appId,
  bundleId,
  orgId,
}: {
  appId: string
  bundleId: string
  orgId: string
}) {
  return api<TAirgapInstall[]>({
    method: 'GET',
    orgId,
    path: `apps/${appId}/airgap-bundles/${bundleId}/installs`,
  })
}
