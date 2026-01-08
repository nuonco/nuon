'use server'

import { revalidatePath } from 'next/cache'
import { api } from '@/lib/api'
import type { TAppBranch, TCreateAppBranchRequest } from '@/types'

export async function createAppBranch(
  orgId: string,
  appId: string,
  data: TCreateAppBranchRequest
): Promise<{ success: boolean; error?: string; branch?: TAppBranch }> {
  try {
    const response = await api<TAppBranch>({
      path: `apps/${appId}/branches`,
      method: 'POST',
      orgId,
      body: data,
    })

    if (response.error) {
      return {
        success: false,
        error: response.error.error || 'Failed to create branch',
      }
    }

    // Revalidate the branches page to show the new branch
    revalidatePath(`/${orgId}/apps/${appId}/branches`)

    return {
      success: true,
      branch: response.data,
    }
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Unknown error occurred',
    }
  }
}
