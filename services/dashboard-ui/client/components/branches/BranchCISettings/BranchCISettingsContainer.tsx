import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { createBranchConfig } from '@/lib'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { Text } from '@/components/common/Text'
import { carryForwardBranchConfigRequest } from '@/components/branches/shared/branch-config-request'
import type { TAPIError, TAppBranch, TAppBranchConfig } from '@/types'
import { BranchCISettingsCard, BranchCISettingsModal } from './BranchCISettings'
import type { BranchCISettingsValues } from './schema'

export const BranchCISettingsContainer = ({
  branch,
  currentConfig,
  orgId,
  appId,
  onSuccess,
}: {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  orgId: string
  appId: string
  onSuccess?: () => void
}) => {
  const { addModal } = useSurfaces()
  const ignoreChangesRegex = currentConfig?.ignore_changes_regex ?? ''
  const sendStatusesOnIgnore = currentConfig?.send_statuses_on_ignore ?? false

  return (
    <BranchCISettingsCard
      ignoreChangesRegex={ignoreChangesRegex}
      sendStatusesOnIgnore={sendStatusesOnIgnore}
      onEdit={() =>
        addModal(
          <BranchCISettingsEditor
            branch={branch}
            currentConfig={currentConfig}
            orgId={orgId}
            appId={appId}
            ignoreChangesRegex={ignoreChangesRegex}
            sendStatusesOnIgnore={sendStatusesOnIgnore}
            onSuccess={onSuccess}
          />
        )
      }
    />
  )
}

const BranchCISettingsEditor = ({
  branch,
  currentConfig,
  orgId,
  appId,
  ignoreChangesRegex,
  sendStatusesOnIgnore,
  onSuccess,
  onSubmit: _onSubmit,
  ...props
}: {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  orgId: string
  appId: string
  ignoreChangesRegex: string
  sendStatusesOnIgnore: boolean
  onSuccess?: () => void
} & IModal) => {
  const queryClient = useQueryClient()
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()

  const {
    mutate: save,
    isPending,
    error,
  } = useMutation<TAppBranchConfig, TAPIError, BranchCISettingsValues>({
    mutationFn: (values: BranchCISettingsValues) =>
      createBranchConfig({
        appId,
        branchId: branch.id ?? '',
        orgId,
        request: carryForwardBranchConfigRequest(currentConfig, {
          ignore_changes_regex: values.ignoreChangesRegex,
          send_statuses_on_ignore: values.sendStatusesOnIgnore,
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
        <Toast heading="CI triggers updated" theme="success">
          <Text>Updated CI triggers for {branch.name}.</Text>
        </Toast>
      )
      onSuccess?.()
      removeModal(props.modalId)
    },
  })

  return (
    <BranchCISettingsModal
      ignoreChangesRegex={ignoreChangesRegex}
      sendStatusesOnIgnore={sendStatusesOnIgnore}
      isPending={isPending}
      error={error ?? null}
      onSubmit={(values) => save(values)}
      onCancel={() => removeModal(props.modalId)}
      {...props}
    />
  )
}
