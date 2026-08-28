import { api } from '@/lib/api'
import type { TCustomerManagedBundleDownloadGrant } from '@/types'

export const createCustomerManagedBundleDownloadGrant = ({
  appId,
  bundleId,
  orgId,
}: {
  appId: string
  bundleId: string
  orgId: string
}) =>
  api<TCustomerManagedBundleDownloadGrant>({
    method: 'POST',
    orgId,
    path: `apps/${appId}/customer-managed-bundles/${bundleId}/download-grants`,
  })
