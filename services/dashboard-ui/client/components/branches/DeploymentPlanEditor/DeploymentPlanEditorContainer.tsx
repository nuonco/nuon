import { useMemo } from 'react'
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { createBranchConfig, getAppInstalls, getRunbooks } from '@/lib'
import type { TAPIError, TAppBranch, TAppBranchConfig } from '@/types'
import { carryForwardBranchConfigRequest } from '@/components/branches/shared/branch-config-request'
import { DeploymentPlanEditor } from './DeploymentPlanEditor'
import type { IInstallGroup } from './types'

const toEditorGroups = (config?: TAppBranchConfig): IInstallGroup[] =>
  config?.install_groups?.map((g, idx) => {
    const hasLabelSelector = !!g.label_selector?.match_labels && Object.keys(g.label_selector.match_labels).length > 0
    return {
      id: g.id || `group-${idx}`,
      name: g.name || '',
      install_ids: g.install_ids || [],
      label_selector: g.label_selector || null,
      selection_mode: hasLabelSelector ? 'labels' as const : 'manual' as const,
      order: g.order ?? idx,
      max_parallel: g.max_parallel || 1,
      auto_approve_on_policies_passing: !!g.auto_approve_on_policies_passing,
    }
  }) || []

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
  const { app, labelColors } = useApp()
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()
  const queryClient = useQueryClient()

  const { data: installsResult, isLoading: loadingInstalls } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-installs', org.id, app.id],
    queryFn: () =>
      getAppInstalls({ appId: app.id!, orgId: org.id!, limit: 100 }),
    enabled: !!org.id && !!app.id,
  })

  const availableInstalls = useMemo(
    () =>
      (installsResult?.data ?? []).filter(
        (i) => !i.app_branch_id || i.app_branch_id === branch.id
      ),
    [installsResult, branch.id]
  )

  const { data: runbooksResult, isLoading: loadingRunbooks } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['runbooks', org?.id, app?.id, 'deployment-plan'],
    queryFn: () => getRunbooks({ appId: app!.id, orgId: org!.id, limit: 100 }),
    enabled: !!org?.id && !!app?.id,
  })

  const initialGroups = useMemo(
    () => toEditorGroups(currentConfig),
    [currentConfig]
  )

  const initialPostDeployRunbookIds = useMemo(
    () => currentConfig?.post_deploy_runbook_ids ?? [],
    [currentConfig]
  )

  const { mutate: save, isPending: isSaving } = useMutation({
    mutationFn: async ({
      groups,
      postDeployRunbookIds,
    }: {
      groups: IInstallGroup[]
      postDeployRunbookIds: string[]
    }) => {
      const installGroupsForApi = groups.map((group, index) => {
        const matchLabels = group.label_selector?.match_labels
        const useLabels =
          group.selection_mode === 'labels' && !!matchLabels && Object.keys(matchLabels).length > 0

        return {
          name: group.name,
          install_ids: useLabels ? [] : group.install_ids || [],
          label_selector: useLabels ? group.label_selector : undefined,
          order: index,
          max_parallel: group.max_parallel || 1,
          auto_approve_on_policies_passing:
            group.auto_approve_on_policies_passing,
        }
      })

      return createBranchConfig({
        appId: app.id!,
        branchId: branch.id || '',
        orgId: org.id!,
        request: carryForwardBranchConfigRequest(currentConfig, {
          install_groups: installGroupsForApi,
          post_deploy_runbook_ids: postDeployRunbookIds,
        }),
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['app-branch', org.id, app.id, branch.id] })
      queryClient.invalidateQueries({ queryKey: ['branch-configs', org.id, app.id, branch.id] })
      addToast(
        <Toast heading="Deployment plan saved" theme="success">
          <Text>A new config version has been created.</Text>
        </Toast>
      )
      onSuccess?.()
      removeModal(props.modalId)
    },
    onError: (error: TAPIError) => {
      addToast(
        <Toast heading="Deployment plan save failed" theme="error">
          <Text>{error.description || error.error || 'Unable to save deployment plan.'}</Text>
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
      labelColors={labelColors}
      orgId={org.id!}
      runbooks={runbooksResult?.data ?? []}
      loadingRunbooks={loadingRunbooks}
      initialPostDeployRunbookIds={initialPostDeployRunbookIds}
      onSave={(groups, postDeployRunbookIds) =>
        save({ groups, postDeployRunbookIds })
      }
      onCancel={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const EditDeploymentPlanButton = ({
  branch,
  currentConfig,
  onSuccess,
  label = 'Deployment plan',
  ...props
}: {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  onSuccess?: () => void
  label?: string
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
      {props?.isMenuButton ? null : <Icon variant="SlidersHorizontalIcon" size={16} />}
      {label}
      {props?.isMenuButton ? <Icon variant="SlidersHorizontalIcon" size={16} /> : null}
    </Button>
  )
}
