import { useMemo, useState } from 'react'
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createBranchConfig, getAppInstalls } from '@/lib'
import type { TAPIError, TAppBranch, TAppBranchConfig } from '@/types'
import {
  PreviewDefaultsEditor,
  previewDefaultsFromConfig,
  previewDefaultsToConfig,
  type IPreviewDefaults,
} from '@/components/branches/shared/PreviewDefaultsEditor'
import { carryForwardBranchConfigRequest } from '@/components/branches/shared/branch-config-request'
import { formatPreviewDefaultsSummary } from '@/components/branches/shared/preview-run-utils'

interface IPreviewConfigSection {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  orgId: string
  appId: string
  onSuccess?: () => void
}

export const PreviewConfigSection = ({
  branch,
  currentConfig,
  orgId,
  appId,
  onSuccess,
}: IPreviewConfigSection) => {
  const { addModal } = useSurfaces()

  const openEditModal = () => {
    addModal(
      <PreviewConfigEditorContainer
        branch={branch}
        currentConfig={currentConfig}
        orgId={orgId}
        appId={appId}
        onSuccess={onSuccess}
      />
    )
  }

  return (
    <PreviewConfigReadView
      currentConfig={currentConfig}
      orgId={orgId}
      appId={appId}
      branch={branch}
      onEdit={openEditModal}
    />
  )
}

interface IPreviewConfigReadView {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  orgId: string
  appId: string
  onEdit: () => void
}

const PreviewConfigReadView = ({
  branch,
  currentConfig,
  orgId,
  appId,
  onEdit,
}: IPreviewConfigReadView) => {
  const { data: installsResult, isLoading } = useQuery({
    queryKey: ['app-installs', orgId, appId],
    queryFn: () => getAppInstalls({ appId, orgId, limit: 100 }),
    enabled: !!orgId && !!appId,
  })

  const availableInstalls = useMemo(
    () =>
      (installsResult?.data ?? []).filter(
        (i) => !i.app_branch_id || i.app_branch_id === branch.id
      ),
    [installsResult, branch.id]
  )

  const hasGithubVCS = !!(
    currentConfig?.connected_github_vcs_config || currentConfig?.public_git_vcs_config
  )

  const summary = useMemo(
    () =>
      formatPreviewDefaultsSummary(currentConfig?.preview_config, availableInstalls, {
        includeGithub: hasGithubVCS,
      }),
    [currentConfig?.preview_config, availableInstalls, hasGithubVCS]
  )

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <Text variant="base" weight="strong">
            Preview settings
          </Text>
          {currentConfig?.config_number != null && (
            <Badge theme="info" size="sm">
              v{currentConfig.config_number}
            </Badge>
          )}
        </div>
        <Button variant="secondary" onClick={onEdit} disabled={isLoading}>
          Edit preview settings
        </Button>
      </div>

      <Card className="p-4 flex flex-col gap-2">
        <Text variant="subtext" theme="neutral">
          Defaults used when triggering preview runs from this branch.
        </Text>
        <Text variant="base">{summary}</Text>
        {!currentConfig?.preview_config && (
          <Text variant="subtext" theme="neutral">
            Using platform defaults until you save custom settings.
          </Text>
        )}
      </Card>
    </div>
  )
}

interface IPreviewConfigEditorContainer extends IModal {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  orgId: string
  appId: string
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

  const { data: installsResult, isLoading: loadingInstalls } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-installs', orgId, appId],
    queryFn: () => getAppInstalls({ appId, orgId, limit: 100 }),
    enabled: !!orgId && !!appId,
  })

  const availableInstalls = useMemo(
    () =>
      (installsResult?.data ?? []).filter(
        (i) => !i.app_branch_id || i.app_branch_id === branch.id
      ),
    [installsResult, branch.id]
  )

  const initialDefaults = useMemo(
    () => previewDefaultsFromConfig(currentConfig?.preview_config, availableInstalls),
    [currentConfig?.preview_config, availableInstalls]
  )

  const [previewDefaults, setPreviewDefaults] = useState(initialDefaults)

  const hasGithubVCS = !!(
    currentConfig?.connected_github_vcs_config || currentConfig?.public_git_vcs_config
  )

  const { mutate: save, isPending: isSaving } = useMutation({
    mutationFn: () =>
      createBranchConfig({
        appId,
        branchId: branch.id!,
        orgId,
        request: carryForwardBranchConfigRequest(currentConfig, {
          preview_config: previewDefaultsToConfig(previewDefaults, availableInstalls),
        }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['app-branch', orgId, appId, branch.id] })
      queryClient.invalidateQueries({ queryKey: ['branch-configs', orgId, appId, branch.id] })
      addToast(
        <Toast heading="Preview settings saved" theme="success">
          <Text>A new config version has been created.</Text>
        </Toast>
      )
      onSuccess?.()
      removeModal(props.modalId)
    },
    onError: (error: TAPIError) => {
      addToast(
        <Toast heading="Save failed" theme="error">
          <Text>{error.description || error.error || 'Unable to save preview settings.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <Modal
      heading="Edit preview settings"
      primaryActionTrigger={{
        children: isSaving ? 'Saving...' : 'Save',
        disabled: isSaving || loadingInstalls,
        onClick: () => save(),
        variant: 'primary',
      }}
      secondaryActionTrigger={{
        children: 'Cancel',
        onClick: () => removeModal(props.modalId),
        disabled: isSaving,
      }}
      {...props}
    >
      <PreviewDefaultsEditor
        value={previewDefaults}
        onChange={setPreviewDefaults}
        availableInstalls={availableInstalls}
        hasGithubVCS={hasGithubVCS}
        disabled={isSaving || loadingInstalls}
        showHeader={false}
      />
    </Modal>
  )
}
