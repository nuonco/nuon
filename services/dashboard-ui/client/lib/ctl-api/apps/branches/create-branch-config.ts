import { api } from '@/lib/api'
import type { TAppBranchConfig, TAppBranchInstallGroup } from '@/types'

export type TCreateBranchConfigRequest = {
  connected_github_vcs_config?: {
    vcs_connection_id: string
    repo: string
    branch: string
    directory?: string
    path_filter?: string
  }
  public_git_vcs_config?: {
    repo: string
    branch: string
    directory?: string
    path_filter?: string
  }
  install_groups?: Array<{
    name: string
    install_ids?: string[]
    label_selector?: { match_labels?: Record<string, string>; not_match_labels?: Record<string, string> } | null
    order: number
    max_parallel?: number
    use_for_previews?: boolean
  }>
}

export const createBranchConfig = ({
  appId,
  branchId,
  orgId,
  request,
}: {
  appId: string
  branchId: string
  orgId: string
  request: TCreateBranchConfigRequest
}) =>
  api<TAppBranchConfig>({
    path: `apps/${appId}/branches/${branchId}/configs`,
    orgId,
    method: 'POST',
    body: request,
  })
