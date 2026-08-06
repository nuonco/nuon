import { api } from '@/lib/api'
import type { TAirgapBundle, TCreateAirgapBundleRequest } from '@/types'

export const createAirgapBundle = ({
  appId,
  body,
  orgId,
}: {
  appId: string
  body: TCreateAirgapBundleRequest
  orgId: string
}) =>
  api<TAirgapBundle>({
    body,
    method: 'POST',
    orgId,
    path: `apps/${appId}/airgap-bundles`,
  })
