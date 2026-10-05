import { useCallback } from 'react'
import { useNavigate, useSearchParams } from 'react-router'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { useOrg } from '@/hooks/use-org'
import type { TInstallWorkflow, TInstallWorkflowStep } from '@/types'
import { GroupActionButton } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep/GroupApprovalActions'
import {
  BranchRunApproval,
  type IBranchRunApprovalItem,
} from './BranchRunApproval'
import { getGroupName, useGroupPlanHref } from './use-group-plan-href'

interface IBranchRunApprovalContainer {
  run: TInstallWorkflow
}

export const BranchRunApprovalContainer = ({
  run,
}: IBranchRunApprovalContainer) => {
  const { org } = useOrg()
  const orgId = org?.id ?? ''
  const hasNewAppIA = useNewAppIA()
  const navigate = useNavigate()
  const groupPlanHref = useGroupPlanHref()
  const [, setSearchParams] = useSearchParams()

  const openStep = useCallback(
    (step: TInstallWorkflowStep) => {
      if (hasNewAppIA) {
        navigate(groupPlanHref(step.name))
        return
      }
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev)
          next.set('workflow', run.id ?? '')
          if (step.id) next.set('step', step.id)
          return next
        },
        { replace: true }
      )
    },
    [hasNewAppIA, navigate, groupPlanHref, run.id, setSearchParams]
  )

  if (run.status?.status === 'cancelled') {
    return null
  }

  const items: IBranchRunApprovalItem[] = (run.steps ?? [])
    .filter(
      (step) =>
        step.execution_type === 'approval' &&
        step.status?.status === 'approval-awaiting' &&
        !step.approval?.response &&
        !!step.approval?.id
    )
    .map((step) => {
      const groupName = getGroupName(step.name)
      return {
        key: step.id ?? step.approval!.id!,
        groupName,
        onReview: () => openStep(step),
        actions: (
          <GroupActionButton
            action="approve"
            target={{
              orgId,
              workflowId: step.install_workflow_id ?? run.id ?? '',
              stepId: step.id ?? '',
              approvalId: step.approval!.id!,
              groupName,
            }}
          />
        ),
      }
    })

  return <BranchRunApproval items={items} />
}
