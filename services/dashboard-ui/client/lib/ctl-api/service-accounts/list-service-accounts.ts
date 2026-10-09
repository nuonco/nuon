import { api } from '@/lib/api'
import type {
  TPaginationParams,
  TServiceAccount,
  TServiceAccountManagement,
} from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const listServiceAccounts = ({
  orgId,
  limit,
  offset,
  management,
  purpose,
  q,
}: {
  orgId: string
  management?: TServiceAccountManagement
  purpose?: string
  q?: string
} & TPaginationParams) =>
  api<TServiceAccount[]>({
    path: `service-accounts${buildQueryParams({
      limit,
      offset,
      management,
      purpose: purpose || undefined,
      q: q || undefined,
    })}`,
    orgId,
  })
