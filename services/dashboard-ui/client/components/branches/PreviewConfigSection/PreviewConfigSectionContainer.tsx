import { useMemo } from 'react'
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import {
  Button,
  type IButtonAsButton,
} from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import type { IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { createBranchConfig, getAppInstalls } from '@/lib'
import type {
  TAPIError,
  TAppBranch,
  TAppBranchConfig,
  TAppBranchPreviewConfig,
} from '@/types'
import { carryForwardBranchConfigRequest } from '@/components/branches/shared/branch-config-request'
import { PreviewConfigEditorModal } from './PreviewConfigEditorModal'
import { PreviewConfigSection } from './PreviewConfigSection'

interface IPreviewConfigSectionContainer {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  orgId: string
  appId: string
}

const useAvailableInstalls = ({
  branch,
  orgId,
  appId,
}: Pick<IPreviewConfigSectionContainer, 'branch' | 'orgId' | 'appId'>) => {
  const query = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-installs', orgId, appId],
    queryFn: () => getAppInstalls({ appId, orgId, limit: 100 }),
    enabled: !!orgId && !!appId,
  })

  const installs = useMemo(
    () =>
      (query.data?.data ?? []).filter(
        (install) => !install.app_branch_id || install.app_branch_id === branch.id
      ),
    [query.data, branch.id]
  )

  return { installs, isLoading: query.isLoading }
}

export const PreviewConfigSectionContainer = ({
  branch,
  currentConfig,
  orgId,
  appId,
}: IPreviewConfigSectionContainer) => {
  const { installs, isLoading } = useAvailableInstalls({
    branch,
    orgId,
    appId,
  })
  const hasGithubVCS = !!(
    currentConfig?.connected_github_vcs_config ||
    currentConfig?.public_git_vcs_config
  )

  return (
    <PreviewConfigSection
      currentConfig={currentConfig}
      installs={installs}
      hasGithubVCS={hasGithubVCS}
      isLoading={isLoading}
    />
  )
}

interface IPreviewConfigEditorContainer
  extends IPreviewConfigSectionContainer,
    IModal {
  onSuccess?: () => void
}

const PreviewConfigEditorContainer = ({
  branch,
  currentConfig,
  orgId,
  appId,
  onSuccess,
  ...props
}: IPreviewConfigEditorContainer) => {
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()
  const queryClient = useQueryClient()
  const { installs, isLoading } = useAvailableInstalls({
    branch,
    orgId,
    appId,
  })
  const hasGithubVCS = !!(
    currentConfig?.connected_github_vcs_config ||
    currentConfig?.public_git_vcs_config
  )

  const { mutate: save, isPending, error } = useMutation<
    Awaited<ReturnType<typeof createBranchConfig>>,
    TAPIError,
    TAppBranchPreviewConfig
  >({
    mutationFn: (previewConfig) =>
      createBranchConfig({
        appId,
        branchId: branch.id!,
        orgId,
        request: carryForwardBranchConfigRequest(currentConfig, {
          preview_config: previewConfig,
        }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['app-branch', orgId, appId, branch.id],
      })
      queryClient.invalidateQueries({
        queryKey: ['branch-configs', orgId, appId, branch.id],
      })
      addToast(
        <Toast heading="Preview settings saved" theme="success">
          <Text>
            Saved preview defaults for {branch.name}. A new config version was
            created.
          </Text>
        </Toast>
      )
      onSuccess?.()
      removeModal(props.modalId)
    },
  })

  return (
    <PreviewConfigEditorModal
      {...props}
      currentConfig={currentConfig}
      installs={installs}
      hasGithubVCS={hasGithubVCS}
      isPending={isPending}
      isLoading={isLoading}
      error={error}
      onSubmit={save}
      onCancel={() => removeModal(props.modalId)}
    />
  )
}

export const EditPreviewConfigButton = ({
  branch,
  currentConfig,
  orgId,
  appId,
  onSuccess,
  ...props
}: IPreviewConfigSectionContainer &
  Pick<IPreviewConfigEditorContainer, 'onSuccess'> &
  IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = (
    <PreviewConfigEditorContainer
      branch={branch}
      currentConfig={currentConfig}
      orgId={orgId}
      appId={appId}
      onSuccess={onSuccess}
    />
  )

  return (
    <Button variant="secondary" onClick={() => addModal(modal)} {...props}>
      Edit preview settings
    </Button>
  )
}
