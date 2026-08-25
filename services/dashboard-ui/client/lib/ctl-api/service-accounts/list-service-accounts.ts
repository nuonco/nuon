import { api } from '@/lib/api'
import type { TAccount, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const listServiceAccounts = ({
  orgId,
  limit,
  offset,
  includeRunners,
  includeStacks,
}: {
  orgId: string
  includeRunners?: boolean
  includeStacks?: boolean
} & TPaginationParams) =>
  api<TAccount[]>({
    path: `service-accounts${buildQueryParams({
      limit,
      offset,
      include_runners: includeRunners ? 'true' : undefined,
      include_stacks: includeStacks ? 'true' : undefined,
    })}`,
    orgId,
  })
