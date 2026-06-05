import { useMemo } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { createBranchConfig, getAppInstalls } from '@/lib'
import type { TCreateBranchConfigRequest } from '@/lib/ctl-api/apps/branches/create-branch-config'
import type { TAppBranch, TAppBranchConfig } from '@/types'
import { DeploymentPlanEditor } from './DeploymentPlanEditor'
import type { IInstallGroup } from './types'

const toEditorGroups = (config?: TAppBranchConfig): IInstallGroup[] =>
  config?.install_groups?.map((g, idx) => ({
    id: g.id || `group-${idx}`,
    name: g.name || '',
    install_ids: g.install_ids || [],
    order: g.order ?? idx,
    max_parallel: g.max_parallel || 1,
    requires_approval: g.requires_approval || false,
    rollback_on_failure: g.rollback_on_failure || false,
  })) || []

interface IDeploymentPlanEditorContainer extends IModal {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  onSuccess?: () => void
}

export const DeploymentPlanEditorContainer = ({
  branch,
  currentConfig,
  onSuccess,
  ...props
}: IDeploymentPlanEditorContainer) => {
  const { org } = useOrg()
  const { app } = useApp()
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()

  const { data: installsResult, isLoading: loadingInstalls } = useQuery({
    queryKey: ['app-installs', org.id, app.id],
    queryFn: () =>
      getAppInstalls({ appId: app.id!, orgId: org.id!, limit: 100 }),
    enabled: !!org.id && !!app.id,
  })

  const availableInstalls = installsResult?.data ?? []

  const initialGroups = useMemo(
    () => toEditorGroups(currentConfig),
    [currentConfig]
  )

  const { mutate: save, isPending: isSaving } = useMutation({
    mutationFn: async (groups: IInstallGroup[]) => {
      const installGroupsForApi = groups.map((group, index) => ({
        name: group.name,
        install_ids: group.install_ids || [],
        order: index,
        max_parallel: group.max_parallel || 1,
        requires_approval: group.requires_approval || false,
        rollback_on_failure: group.rollback_on_failure || false,
      }))

      const request: TCreateBranchConfigRequest = {
        install_groups: installGroupsForApi,
      }

      if (currentConfig?.connected_github_vcs_config) {
        request.connected_github_vcs_config = {
          vcs_connection_id:
            currentConfig.connected_github_vcs_config.vcs_connection_id || '',
          repo: currentConfig.connected_github_vcs_config.repo || '',
          branch: currentConfig.connected_github_vcs_config.branch || '',
          directory: currentConfig.connected_github_vcs_config.directory,
          path_filter: currentConfig.connected_github_vcs_config.path_filter,
        }
      } else if (currentConfig?.public_git_vcs_config) {
        request.public_git_vcs_config = {
          repo: currentConfig.public_git_vcs_config.repo || '',
          branch: currentConfig.public_git_vcs_config.branch || '',
          directory: currentConfig.public_git_vcs_config.directory,
          path_filter: currentConfig.public_git_vcs_config.path_filter,
        }
      }

      return createBranchConfig({
        appId: app.id!,
        branchId: branch.id || '',
        orgId: org.id!,
        request,
      })
    },
    onSuccess: () => {
      addToast(
        <Toast heading="Deployment plan saved" theme="success">
          <Text>A new config version has been created.</Text>
        </Toast>
      )
      onSuccess?.()
      removeModal(props.modalId)
    },
    onError: (error: Error) => {
      addToast(
        <Toast heading="Deployment plan save failed" theme="error">
          <Text>{error.message || 'An unknown error occurred.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <DeploymentPlanEditor
      initialGroups={initialGroups}
      availableInstalls={availableInstalls}
      loadingInstalls={loadingInstalls}
      isSaving={isSaving}
      onSave={(groups) => save(groups)}
      onCancel={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const EditDeploymentPlanButton = ({
  branch,
  currentConfig,
  onSuccess,
  ...props
}: {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  onSuccess?: () => void
} & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = (
    <DeploymentPlanEditorContainer
      branch={branch}
      currentConfig={currentConfig}
      onSuccess={onSuccess}
    />
  )
  return (
    <Button
      variant="secondary"
      onClick={() => addModal(modal)}
      {...props}
    >
      <Icon variant="SlidersHorizontalIcon" size={16} />
      Deployment plan
    </Button>
  )
}
