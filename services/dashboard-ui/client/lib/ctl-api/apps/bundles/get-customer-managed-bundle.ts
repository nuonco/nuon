import { api } from '@/lib/api'
import type { TCustomerManagedBundle } from '@/types'

export const getCustomerManagedBundle = ({
  appId,
  bundleId,
  orgId,
}: {
  appId: string
  bundleId: string
  orgId: string
}) =>
  api<TCustomerManagedBundle>({
    orgId,
    path: `apps/${appId}/customer-managed-bundles/${bundleId}`,
  })
