import { api } from '@/lib/api'
import type { TAppBranch, TCreateAppBranchRequest } from '@/types'

export async function createAppBranch({
  appId,
  body,
  orgId,
}: {
  appId: string
  body: TCreateAppBranchRequest
  orgId: string
}) {
  return api<TAppBranch>({
    body,
    method: 'POST',
    orgId,
    path: `apps/${appId}/branches`,
  })
}
