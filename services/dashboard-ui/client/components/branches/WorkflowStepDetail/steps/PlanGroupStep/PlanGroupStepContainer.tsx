import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { approveWorkflowStep } from '@/lib'
import type { TInstallWorkflowStep, TAPIError } from '@/types'
import { PlanGroupStep } from './PlanGroupStep'

interface IPlanGroupStepContainer {
  step: TInstallWorkflowStep
  metadata: Record<string, any>
}

export const PlanGroupStepContainer = ({ step, metadata }: IPlanGroupStepContainer) => {
  const { org } = useOrg()
  const orgId = org?.id ?? ''
  const { addToast } = useToast()
  const queryClient = useQueryClient()

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

  const { mutate: respond, isPending: isResponding } = useMutation({
    mutationFn: (responseType: 'approve' | 'deny' | 'deny-skip-current') =>
      approveWorkflowStep({
        orgId,
        workflowId: step.install_workflow_id,
        workflowStepId: step.id,
        approvalId: approvalId!,
        body: { response_type: responseType, note: '' },
      }),
    onSuccess: () => {
      addToast(
        <Toast heading="Plan approved" theme="success">
          <Text>Approved install group plan.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['branch-run'] })
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Approval failed" theme="error">
          <Text>{err?.error || 'Unable to respond to approval.'}</Text>
        </Toast>
      )
    },
  })

  const installs = (plan?.installs || metadata.installs || []) as any[]
  const groupName = plan?.install_group || metadata.install_group_name || step.name?.replace(/^plan install group:\s*/i, '')

  return (
    <PlanGroupStep
      installs={installs}
      groupName={groupName}
      orgId={orgId}
      hasResponse={hasResponse}
      responseType={step.approval?.response?.response_type}
      showApproveBar={hasApproval && isAwaiting && !hasResponse}
      isResponding={isResponding}
      isInProgress={step.status?.status === 'in-progress'}
      onRespond={respond}
    />
  )
}
