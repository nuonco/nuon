import { useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import type { TInstallWorkflowStep } from '@/types'
import { PlanGroupStep } from './PlanGroupStep'
import { GroupApprovalActions } from './GroupApprovalActions'

interface IPlanGroupStepContainer {
  step: TInstallWorkflowStep
  metadata: Record<string, any>
}

export const PlanGroupStepContainer = ({ step, metadata }: IPlanGroupStepContainer) => {
  const { org } = useOrg()
  const { labelColors } = useApp()
  const orgId = org?.id ?? ''

  const approvalId = step.approval?.id
  const hasApproval = step.execution_type === 'approval' && !!approvalId
  const hasResponse = !!step.approval?.response
  const isAwaiting = step.status?.status === 'approval-awaiting'

  const { data: plan } = useQuery({
    queryKey: ['approval-plan', orgId, step.id, approvalId],
    queryFn: async () => {
      const res = await fetch(
        `/api/orgs/${orgId}/workflows/${step.install_workflow_id}/steps/${step.id}/approvals/${approvalId}/contents`
      )
      if (!res.ok) throw new Error(`Failed to fetch approval contents: ${res.status}`)
      return res.json()
    },
    enabled: !!orgId && !!step.id && !!step.install_workflow_id && !!approvalId,
  })

  const installs = (plan?.installs || metadata.installs || []) as any[]
  const groupName = plan?.install_group || metadata.install_group_name || step.name?.replace(/^plan install group:\s*/i, '')
  const showApproveBar = hasApproval && isAwaiting && !hasResponse

  return (
    <PlanGroupStep
      installs={installs}
      groupName={groupName}
      orgId={orgId}
      labelColors={labelColors}
      hasResponse={hasResponse}
      responseType={step.approval?.response?.response_type}
      showApproveBar={showApproveBar}
      isInProgress={step.status?.status === 'in-progress'}
      actions={
        showApproveBar ? (
          <GroupApprovalActions
            target={{
              orgId,
              workflowId: step.install_workflow_id,
              stepId: step.id,
              approvalId: approvalId!,
              groupName: groupName || 'install group',
            }}
          />
        ) : undefined
      }
    />
  )
}
