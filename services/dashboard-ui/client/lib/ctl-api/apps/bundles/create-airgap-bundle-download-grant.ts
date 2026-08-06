import { api } from '@/lib/api'
import type { TAirgapBundleDownloadGrant } from '@/types'

export const createAirgapBundleDownloadGrant = ({
  appId,
  bundleId,
  orgId,
}: {
  appId: string
  bundleId: string
  orgId: string
}) =>
  api<TAirgapBundleDownloadGrant>({
    method: 'POST',
    orgId,
    path: `apps/${appId}/airgap-bundles/${bundleId}/download-grants`,
  })
