'use server'

import {
  createAppBranch as create,
  type TCreateAppBranchRequest,
} from '@/lib/ctl-api/apps/branches'

export async function createAppBranch(
  orgId: string,
  appId: string,
  body: TCreateAppBranchRequest
) {
  return create({
    appId,
    body,
    orgId,
  })
}