import { useContext } from 'react'
import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useInstallLink } from '@/hooks/use-install-path'
import { useInstallOverview } from '@/hooks/use-install-overview'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { isDrifted } from '@/components/install-alignment/alignment-utils'
import { reconcileInstallAppVersion } from '@/lib'
import { ActiveWorkflowsContext } from '@/providers/active-workflows-provider'
import {
  ReconcileAppVersionModal,
  type TReconcileAppVersionValues,
} from './ReconcileAppVersion'

export const reconcileAppVersionDisabledReason = ({
  appliedConfigId,
  targetConfigId,
  hasActiveWorkflow,
}: {
  appliedConfigId?: string
  targetConfigId?: string
  hasActiveWorkflow: boolean
}): string | undefined => {
  if (hasActiveWorkflow) {
    return 'A reconcile or app branch config update is already running for this install'
  }
  if (appliedConfigId && targetConfigId && appliedConfigId === targetConfigId) {
    return 'This install is already on its pinned app config'
  }
  return undefined
}

export const ReconcileAppVersionModalContainer = ({
  ...props
}: Omit<IModal, 'onSubmit'>) => {
  const navigate = useNavigate()
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const { install } = useInstall()
  const installLink = useInstallLink()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const { data: overview } = useInstallOverview()

  const appliedConfigId = install?.app_config_ref?.applied_config_id
  const targetConfigId = install?.app_config_id
  const unappliedCount = (overview?.alignment?.resources ?? []).filter(
    (resource) => resource.category === 'changes-not-applied' || isDrifted(resource)
  ).length

  const { mutate, isPending, error } = useMutation({
    mutationFn: (values: TReconcileAppVersionValues) =>
      reconcileInstallAppVersion({
        orgId: org.id,
        installId: install.id,
        body: { plan_only: values.planOnly },
      }),
    onSuccess: (result) => {
      addToast(
        <Toast heading={`${install.name} reconcile was started`} theme="success">
          <Text>Reconciling {install.name}'s app version workflow was created.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      queryClient.invalidateQueries({ queryKey: ['install-overview'] })
      removeModal(props.modalId)
      const workflowId = result.data.workflow_id
      if (workflowId) {
        navigate(installLink({ installId: install.id, appId: install.app_id, suffix: `/workflows/${workflowId}` }))
      } else {
        navigate(installLink({ installId: install.id, appId: install.app_id, suffix: `/workflows` }))
      }
    },
    onError: () => {
      addToast(
        <Toast heading="Reconcile not started" theme="error">
          <Text>Unable to reconcile {install.name}'s app version.</Text>
        </Toast>
      )
    },
  })

  return (
    <ReconcileAppVersionModal
      installName={install.name}
      appliedConfigId={appliedConfigId}
      targetConfigId={targetConfigId}
      unappliedCount={unappliedCount}
      isPending={isPending}
      error={error}
      onSubmit={(values) => mutate(values)}
      {...props}
    />
  )
}

export const ReconcileAppVersionButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const { install } = useInstall()
  const activeWorkflows = useContext(ActiveWorkflowsContext)?.activeWorkflows ?? []

  const appliedConfigId = install?.app_config_ref?.applied_config_id
  const targetConfigId = install?.app_config_id
  const hasActiveWorkflow = activeWorkflows.some(
    (workflow) =>
      workflow.owner_id === install?.id &&
      (workflow.type === 'reconcile_app_version' ||
        workflow.type === 'app_branch_config_update')
  )

  const disabledReason = reconcileAppVersionDisabledReason({
    appliedConfigId,
    targetConfigId,
    hasActiveWorkflow,
  })

  return (
    <Button
      onClick={() => addModal(<ReconcileAppVersionModalContainer />)}
      disabled={!!disabledReason}
      tooltipProps={disabledReason ? { tipContent: disabledReason } : undefined}
      {...props}
    >
      Reconcile app version
      <Icon variant="ArrowsClockwiseIcon" />
    </Button>
  )
}
