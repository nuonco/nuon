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
import { createBranchConfig, updateBranch } from '@/lib'
import type { TAppBranch, TAppBranchConfig } from '@/types'
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

  const formatError = (err: any): string => {
    if (!err) return 'An error occurred'
    if (typeof err === 'string') return err
    return (
      err.user_error && typeof err.user_error === 'string'
        ? err.user_error
        : err.error || err.description || err.message || 'An error occurred'
    )
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
          throw new Error(formatError(err))
        }
      }

      const request: any = {}

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
          throw new Error(formatError(err))
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
      orgId={org.id}
      vcsConnections={org?.vcs_connections || []}
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
