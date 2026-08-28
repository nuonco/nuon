import { api } from '@/lib/api'
import type {
  TCustomerManagedBundle,
  TCreateCustomerManagedBundleRequest,
} from '@/types'

export const createCustomerManagedBundle = ({
  appId,
  body,
  orgId,
}: {
  appId: string
  body: TCreateCustomerManagedBundleRequest
  orgId: string
}) =>
  api<TCustomerManagedBundle>({
    body,
    method: 'POST',
    orgId,
    path: `apps/${appId}/customer-managed-bundles`,
  })
