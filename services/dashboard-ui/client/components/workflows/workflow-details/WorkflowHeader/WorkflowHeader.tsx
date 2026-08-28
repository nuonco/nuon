import { Badge } from '@/components/common/Badge'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { humanize } from '@/utils/string-utils'
import { WorkflowActionButtons } from '../WorkflowActionButtons'
import type { TWorkflow, TInstall } from '@/types'

interface IWorkflowHeader {
  workflow: TWorkflow
  install?: TInstall
  readOnly?: boolean
}

export const WorkflowHeader = ({
  workflow,
  install,
  readOnly = false,
}: IWorkflowHeader) => {
  const hasDrift =
    install?.drifted_objects?.length &&
    install?.drifted_objects?.find(
      (d) => d?.install_workflow_id === workflow?.id
    )

  const approvalType = workflow?.metadata?.approval_type

  return (
    <DetailHeader
      title={
        workflow?.type === 'action_workflow_run' &&
        workflow?.metadata?.adhoc_action
          ? `Adhoc action run (${workflow?.metadata?.install_action_workflow_name})`
          : workflow?.name || humanize(workflow?.type)
      }
      description="Watch your app get updated here and provide needed approvals."
      status={
        <>
          {hasDrift ? (
            <Badge variant="code" theme="warn" size="sm">
              drift detected
            </Badge>
          ) : null}
          {workflow?.approval_option === 'approve-all' && approvalType ? (
            <Badge variant="code" size="sm">
              {approvalType === 'approve-workflow'
                ? 'auto-approve (workflow)'
                : approvalType === 'install-config'
                  ? 'auto-approve (config)'
                  : 'auto-approve'}
            </Badge>
          ) : null}
        </>
      }
      id={workflow?.id}
      actions={readOnly ? null : <WorkflowActionButtons />}
    />
  )
}
