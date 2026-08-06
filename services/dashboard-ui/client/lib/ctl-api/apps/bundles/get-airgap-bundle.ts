import { api } from '@/lib/api'
import type { TAirgapBundle } from '@/types'

export const getAirgapBundle = ({
  appId,
  bundleId,
  orgId,
}: {
  appId: string
  bundleId: string
  orgId: string
}) =>
  api<TAirgapBundle>({
    orgId,
    path: `apps/${appId}/airgap-bundles/${bundleId}`,
  })
