import { useQuery } from '@tanstack/react-query'
import { InstallActionManualRunButton } from '@/components/actions/InstallActionManualRun'
import { Icon } from '@/components/common/Icon'
import { useOrg } from '@/hooks/use-org'
import { getInstallAction } from '@/lib'
import { ActionRunMetadata } from './ActionRunMetadata'
import type { IActionRunMetadata } from '../types'

export const ActionRunMetadataContainer = (props: IActionRunMetadata) => {
  const { org } = useOrg()
  const { actionRun, step } = props

  const isAdhoc = actionRun?.trigger_type === 'adhoc'
  const installId = step?.owner_id
  const actionWorkflowId = actionRun?.config?.action_workflow_id

  const { data: installAction } = useQuery({
    queryKey: ['action', org?.id, actionWorkflowId],
    queryFn: () =>
      getInstallAction({
        orgId: org!.id,
        installId: installId!,
        actionId: actionWorkflowId!,
      }),
    enabled: !!org?.id && !!installId && !!actionWorkflowId && !isAdhoc,
  })

  const actionWorkflow = installAction?.action_workflow
  const hasManualTrigger =
    actionWorkflow?.configs?.[0]?.triggers?.find((t) => t.type === 'manual') ||
    actionRun?.triggered_by_type === 'manual'

  const rerunButton =
    hasManualTrigger && actionWorkflow && !isAdhoc ? (
      <InstallActionManualRunButton
        action={actionWorkflow}
        actionConfigId={
          actionWorkflow.configs?.[0]?.id ??
          actionRun?.action_workflow_config_id ??
          ''
        }
      >
        Re-run action
        <Icon variant="PlayIcon" />
      </InstallActionManualRunButton>
    ) : null

  return <ActionRunMetadata {...props} orgId={org?.id} rerunButton={rerunButton} />
}
