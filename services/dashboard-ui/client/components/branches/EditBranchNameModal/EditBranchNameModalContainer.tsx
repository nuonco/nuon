import { useMemo, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { useVcsRepoBrowser } from '@/hooks/use-vcs-repo-browser'
import { createBranchConfig, updateBranch } from '@/lib'
import type { TCreateBranchConfigRequest } from '@/lib/ctl-api/apps/branches/create-branch-config'
import type { TAPIError, TAppBranch, TAppBranchConfig } from '@/types'
import {
  EditBranchNameModal,
  type IEditBranchNameModalSubmitData,
} from './EditBranchNameModal'

interface IEditBranchNameModalContainer extends IModal {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  onSuccess?: () => void
}

export const EditBranchNameModalContainer = ({
  branch,
  currentConfig,
  onSuccess,
  onSubmit: _onSubmit,
  ...props
}: IEditBranchNameModalContainer) => {
  const { app } = useApp()
  const { org } = useOrg()
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()
  const queryClient = useQueryClient()

  const [validationError, setValidationError] = useState<string | null>(null)

  const vcsConnections = org?.vcs_connections || []
  const existingConnectionId =
    currentConfig?.connected_github_vcs_config?.vcs_connection_id || ''

  const existingRepo =
    currentConfig?.connected_github_vcs_config?.repo ||
    currentConfig?.public_git_vcs_config?.repo ||
    ''
  const existingBranch =
    currentConfig?.connected_github_vcs_config?.branch ||
    currentConfig?.public_git_vcs_config?.branch ||
    'main'

  const repoOwner = existingRepo.split('/')[0] ?? ''
  const [vcsConnectionId, setVcsConnectionId] = useState(
    existingConnectionId ||
      vcsConnections.find(
        (c) => c.github_account_name?.toLowerCase() === repoOwner.toLowerCase()
      )?.id ||
      vcsConnections[0]?.id ||
      ''
  )

  const vcsBrowser = useVcsRepoBrowser({
    orgId: org.id,
    vcsConnectionId,
    enabled: !!vcsConnectionId,
    initialRepo: existingRepo,
    initialBranch: existingBranch,
  })

  const repos = useMemo(() => {
    const list = vcsBrowser.repos ?? []
    if (
      vcsBrowser.selectedRepo &&
      !list.some((r) => r.full_name === vcsBrowser.selectedRepo!.full_name)
    ) {
      return [vcsBrowser.selectedRepo, ...list]
    }
    return list
  }, [vcsBrowser.repos, vcsBrowser.selectedRepo])

  const formatError = (err: TAPIError | Error): string => {
    if ('error' in err && typeof err.error === 'string') return err.error
    if ('user_error' in err && typeof err.user_error === 'string') return err.user_error
    if ('message' in err && typeof err.message === 'string') return err.message
    return 'An error occurred'
  }

  const { mutate: handleSave, isPending: isSubmitting } = useMutation({
    mutationFn: async (data: IEditBranchNameModalSubmitData) => {
      if (data.branchName !== branch.name) {
        try {
          await updateBranch({
            appId: app.id,
            branchId: branch.id || '',
            orgId: org.id,
            request: { name: data.branchName },
          })
        } catch (err) {
          throw new Error(formatError(err as TAPIError))
        }
      }

      const request: TCreateBranchConfigRequest = {}

      if (data.useVcs && data.selectedRepo) {
        if (data.selectedRepo.private) {
          request.connected_github_vcs_config = {
            vcs_connection_id: data.selectedVcsConnectionId,
            repo: data.selectedRepo.full_name,
            branch: data.selectedBranch,
            directory: data.directory,
            path_filter: data.pathFilter || undefined,
          }
        } else {
          request.public_git_vcs_config = {
            repo: data.selectedRepo.full_name,
            branch: data.selectedBranch,
            directory: data.directory,
            path_filter: data.pathFilter || undefined,
          }
        }
      }

      if (currentConfig?.install_groups && currentConfig.install_groups.length > 0) {
        request.install_groups = currentConfig.install_groups.map((g, idx) => {
          const hasSelector =
            !!g.label_selector?.match_labels &&
            Object.keys(g.label_selector.match_labels).length > 0
          return {
            name: g.name ?? '',
            order: g.order ?? idx,
            max_parallel: g.max_parallel || 1,
            use_for_previews: g.use_for_previews,
            ...(hasSelector
              ? { label_selector: g.label_selector }
              : { install_ids: g.install_ids || [] }),
          }
        })
      }

      const hasVCS = request.connected_github_vcs_config || request.public_git_vcs_config
      const hasGroups = (request.install_groups?.length ?? 0) > 0

      if (hasVCS || hasGroups) {
        try {
          await createBranchConfig({
            appId: app.id,
            branchId: branch.id || '',
            orgId: org.id,
            request,
          })
        } catch (err) {
          throw new Error(formatError(err as TAPIError))
        }
      }
    },
    onSuccess: (_result, data) => {
      queryClient.invalidateQueries({ queryKey: ['app-branch', org.id, app.id, branch.id] })
      queryClient.invalidateQueries({ queryKey: ['app-branches', org.id, app.id] })
      queryClient.invalidateQueries({ queryKey: ['branch-configs', org.id, app.id, branch.id] })
      addToast(
        <Toast heading="Branch updated" theme="success">
          <Text>Updated branch {data.branchName}.</Text>
        </Toast>
      )
      setValidationError(null)
      onSuccess?.()
      removeModal(props.modalId)
    },
    onError: (error: Error) => {
      const msg = error?.message || 'An error occurred'
      setValidationError(msg)
      addToast(
        <Toast heading="Branch update failed" theme="error">
          <Text>{msg}</Text>
        </Toast>
      )
    },
  })

  return (
    <EditBranchNameModal
      branch={branch}
      currentConfig={currentConfig}
      vcsConnections={vcsConnections}
      repos={repos}
      branches={vcsBrowser.branches}
      loadingRepos={vcsBrowser.loadingRepos}
      loadingBranches={vcsBrowser.loadingBranches}
      reposError={vcsBrowser.reposError}
      branchesError={vcsBrowser.branchesError}
      selectedVcsConnectionId={vcsConnectionId}
      onVcsConnectionChange={setVcsConnectionId}
      selectedRepo={vcsBrowser.selectedRepo}
      onRepoChange={vcsBrowser.setSelectedRepo}
      selectedBranch={vcsBrowser.selectedBranch}
      onBranchChange={vcsBrowser.setSelectedBranch}
      isSubmitting={isSubmitting}
      validationError={validationError}
      onSubmit={(data) => handleSave(data)}
      onCancel={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const EditBranchButton = ({
  branch,
  currentConfig,
  onSuccess,
  ...props
}: { branch: TAppBranch; currentConfig?: TAppBranchConfig; onSuccess?: () => void } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <EditBranchNameModalContainer branch={branch} currentConfig={currentConfig} onSuccess={onSuccess} />
  return (
    <Button variant="secondary" onClick={() => addModal(modal)} {...props}>
      {props?.isMenuButton ? null : <Icon variant="PencilSimpleLineIcon" size={16} />}
      Edit branch
      {props?.isMenuButton ? <Icon variant="PencilSimpleLineIcon" size={16} /> : null}
    </Button>
  )
}
