'use server'

import { executeServerAction } from '@/actions/execute-server-action'
import {
  createAppBranch as create,
  type TCreateAppBranchRequest,
} from '@/lib/ctl-api/apps/branches'
import type { TAppBranch } from '@/types'

export async function createAppBranch(
  orgId: string,
  appId: string,
  data: TCreateAppBranchRequest
): Promise<{ success: boolean; error?: string; branch?: TAppBranch }> {
  const result = await executeServerAction({
    action: create,
    args: {
      appId,
      body: data,
      orgId,
    },
    path: `/${orgId}/apps/${appId}/branches`,
  })

  if (result.error) {
    return {
      success: false,
      error: result.error.error || 'Failed to create branch',
    }
  }

  return {
    success: true,
    branch: result.data,
  }
}
