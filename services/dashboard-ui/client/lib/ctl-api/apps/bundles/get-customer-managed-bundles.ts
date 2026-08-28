import { api } from '@/lib/api'
import { buildQueryParams } from '@/utils/build-query-params'
import type { TCustomerManagedBundle, TPaginationParams } from '@/types'

export interface IGetCustomerManagedBundles extends TPaginationParams {
  appId: string
  orgId: string
}

export async function getCustomerManagedBundles({
  appId,
  orgId,
  limit,
  offset,
}: IGetCustomerManagedBundles) {
  return api<TCustomerManagedBundle[]>({
    orgId,
    path: `apps/${appId}/customer-managed-bundles${buildQueryParams({ limit, offset })}`,
    paginated: true,
  })
}
