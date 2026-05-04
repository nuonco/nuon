import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
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

  const [validationError, setValidationError] = useState<string | null>(null)

  const vcsConnections = org?.vcs_connections || []
  const existingConnectionId =
    currentConfig?.connected_github_vcs_config?.vcs_connection_id || ''
  const [vcsConnectionId, setVcsConnectionId] = useState(
    existingConnectionId || vcsConnections[0]?.id || ''
  )

  const existingRepo =
    currentConfig?.connected_github_vcs_config?.repo ||
    currentConfig?.public_git_vcs_config?.repo ||
    ''
  const existingBranch =
    currentConfig?.connected_github_vcs_config?.branch ||
    currentConfig?.public_git_vcs_config?.branch ||
    'main'

  const vcsBrowser = useVcsRepoBrowser({
    orgId: org.id,
    vcsConnectionId,
    enabled: !!vcsConnectionId,
    initialRepo: existingRepo,
    initialBranch: existingBranch,
  })

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
        request.install_groups = currentConfig.install_groups.map((g, idx) => ({
          name: g.name,
          install_ids: g.install_ids || [],
          order: g.order ?? idx,
          max_parallel: g.max_parallel || 1,
          requires_approval: g.requires_approval || false,
          rollback_on_failure: g.rollback_on_failure || false,
        }))
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
    onSuccess: () => {
      addToast(
        <Toast heading="Branch updated successfully" theme="success">
          <Text>Updated branch: {branch.name}</Text>
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
      repos={vcsBrowser.repos}
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
      <Icon variant="PencilSimpleLineIcon" size={16} />
      Edit branch
    </Button>
  )
}
