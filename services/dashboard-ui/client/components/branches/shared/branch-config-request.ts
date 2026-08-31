import type { TCreateBranchConfigRequest } from '@/lib/ctl-api/apps/branches/create-branch-config'
import type { TAppBranchConfig } from '@/types'

export const installGroupsForApi = (config?: TAppBranchConfig): TCreateBranchConfigRequest['install_groups'] =>
  config?.install_groups?.map((g, idx) => {
    const hasSelector =
      !!g.label_selector?.match_labels &&
      Object.keys(g.label_selector.match_labels).length > 0
    return {
      name: g.name ?? '',
      order: g.order ?? idx,
      max_parallel: g.max_parallel || 1,
      ...(hasSelector
        ? { label_selector: g.label_selector }
        : { install_ids: g.install_ids || [] }),
    }
  })

export const vcsConfigForApi = (
  config?: TAppBranchConfig
): Pick<TCreateBranchConfigRequest, 'connected_github_vcs_config' | 'public_git_vcs_config'> => {
  if (config?.connected_github_vcs_config) {
    return {
      connected_github_vcs_config: {
        vcs_connection_id: config.connected_github_vcs_config.vcs_connection_id || '',
        repo: config.connected_github_vcs_config.repo || '',
        branch: config.connected_github_vcs_config.branch || '',
        directory: config.connected_github_vcs_config.directory,
        path_filter: config.connected_github_vcs_config.path_filter,
      },
    }
  }
  if (config?.public_git_vcs_config) {
    return {
      public_git_vcs_config: {
        repo: config.public_git_vcs_config.repo || '',
        branch: config.public_git_vcs_config.branch || '',
        directory: config.public_git_vcs_config.directory,
        path_filter: config.public_git_vcs_config.path_filter,
      },
    }
  }
  return {}
}

export const carryForwardBranchConfigRequest = (
  currentConfig: TAppBranchConfig | undefined,
  overrides: Partial<TCreateBranchConfigRequest> = {}
): TCreateBranchConfigRequest => ({
  install_groups: installGroupsForApi(currentConfig) ?? [],
  post_deploy_runbook_ids: currentConfig?.post_deploy_runbook_ids ?? [],
  preview_config: currentConfig?.preview_config,
  ignore_changes_regex: currentConfig?.ignore_changes_regex,
  send_statuses_on_ignore: currentConfig?.send_statuses_on_ignore,
  ...vcsConfigForApi(currentConfig),
  ...overrides,
})
