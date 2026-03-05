'use server'

import {
  createAppBranch as create,
  createBranchConfig,
  type TCreateAppBranchRequest,
  type TCreateBranchConfigRequest,
} from '@/lib/ctl-api/apps/branches'

export async function createAppBranch(
  orgId: string,
  appId: string,
  body: TCreateAppBranchRequest & {
    vcs_connection_id?: string
    connected_github_vcs_config?: any
    public_git_vcs_config?: any
  }
) {
  // Step 1: Create the branch with just the name
  const { data: branch, error: branchError } = await create({
    appId,
    body: { name: body.name },
    orgId,
  })

  if (branchError || !branch) {
    return { data: null, error: branchError }
  }

  // Step 2: If VCS config provided, create a branch config
  if (body.connected_github_vcs_config || body.public_git_vcs_config) {
    const configBody: TCreateBranchConfigRequest = {}

    if (body.connected_github_vcs_config) {
      configBody.connected_github_vcs_config = {
        ...body.connected_github_vcs_config,
        vcs_connection_id: body.vcs_connection_id || '',
      }
    }

    if (body.public_git_vcs_config) {
      configBody.public_git_vcs_config = body.public_git_vcs_config
    }

    const { error: configError } = await createBranchConfig({
      appId,
      branchId: branch.id,
      request: configBody,
      orgId,
    })

    if (configError) {
      // Branch was created but config failed - log error but still return success
      // The user can create a config later from the branch page
      // eslint-disable-next-line no-console
      console.error('Failed to create branch config:', configError)
      // Return branch successfully but note the config creation failed
      return {
        data: branch,
        error: null, // Still return success since branch was created
      }
    }
  }

  return { data: branch, error: null }
}
