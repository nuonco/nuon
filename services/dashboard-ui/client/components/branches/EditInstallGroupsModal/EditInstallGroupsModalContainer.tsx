import { useMutation, useQuery } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { createBranchConfig, getAppInstalls } from '@/lib'
import type { TCreateBranchConfigRequest } from '@/lib/ctl-api/apps/branches/create-branch-config'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TAppBranch, TAppBranchConfig } from '@/types'
import { EditInstallGroupsModal, type IInstallGroup } from './EditInstallGroupsModal'

interface IEditInstallGroupsModalContainer extends IModal {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  onSuccess?: () => void
}

export const EditInstallGroupsModalContainer = ({
  branch,
  currentConfig,
  onSuccess,
  ...props
}: IEditInstallGroupsModalContainer) => {
  const { app } = useApp()
  const { org } = useOrg()
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()

  const { data: installsResult, isLoading: loadingInstalls } = useQuery({
    queryKey: ['app-installs', org.id, app.id],
    queryFn: () => getAppInstalls({ appId: app.id, orgId: org.id, limit: 100 }),
    enabled: !!org.id && !!app.id,
  })

  const availableInstalls = installsResult?.data ?? []

  const initialGroups: IInstallGroup[] =
    currentConfig?.install_groups?.map((group, idx) => ({
      id: group.id || `group-${idx}`,
      name: group.name || '',
      install_ids: group.install_ids || [],
      order: group.order || idx,
      max_parallel: group.max_parallel || 1,
      requires_approval: group.requires_approval || false,
      rollback_on_failure: group.rollback_on_failure || false,
    })) || []

  const { mutate: saveMutation, isPending: isSaving } = useMutation({
    mutationFn: async (groups: IInstallGroup[]) => {
      if (groups.length === 0) {
        throw new Error('At least one install group is required')
      }

      if (groups.some((g) => !g.name)) {
        throw new Error('All install groups must have a name')
      }

      const installGroupsForApi = groups.map((group, index) => ({
        name: group.name,
        install_ids: group.install_ids || [],
        order: index,
        max_parallel: group.max_parallel || 1,
        requires_approval: group.requires_approval || false,
        rollback_on_failure: group.rollback_on_failure || false,
      }))

      const request: TCreateBranchConfigRequest = { install_groups: installGroupsForApi }

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
        appId: app.id,
        branchId: branch.id || '',
        orgId: org.id,
        request,
      })
    },
    onSuccess: () => {
      addToast(
        <Toast heading="Install groups saved successfully" theme="success">
          <Text>Your install group configuration has been updated.</Text>
        </Toast>
      )
      onSuccess?.()
      removeModal(props.modalId)
    },
    onError: (error: Error) => {
      addToast(
        <Toast heading="Failed to save install groups" theme="error">
          <Text>{error.message || 'An unknown error occurred.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <EditInstallGroupsModal
      initialGroups={initialGroups}
      availableInstalls={availableInstalls}
      loadingInstalls={loadingInstalls}
      isSaving={isSaving}
      onSave={(groups) => saveMutation(groups)}
      onCancel={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const EditInstallGroupsButton = ({
  branch,
  currentConfig,
  onSuccess,
  ...props
}: { branch: TAppBranch; currentConfig?: TAppBranchConfig; onSuccess?: () => void } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <EditInstallGroupsModalContainer branch={branch} currentConfig={currentConfig} onSuccess={onSuccess} />
  return (
    <Button variant="secondary" onClick={() => addModal(modal)} {...props}>
      <Icon variant="SlidersHorizontalIcon" size={16} />
      Manage installs
    </Button>
  )
}
