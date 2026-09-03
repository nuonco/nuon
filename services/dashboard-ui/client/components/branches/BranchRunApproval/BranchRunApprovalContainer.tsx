import { useCallback } from 'react'
import { useSearchParams } from 'react-router'
import { useOrg } from '@/hooks/use-org'
import type { TInstallWorkflow } from '@/types'
import { GroupActionButton } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep/GroupApprovalActions'
import {
  BranchRunApproval,
  type IBranchRunApprovalItem,
} from './BranchRunApproval'

interface IBranchRunApprovalContainer {
  run: TInstallWorkflow
}

const getGroupName = (name?: string) =>
  name?.replace(/^plan install group:\s*/i, '').trim() || 'install group'

export const BranchRunApprovalContainer = ({
  run,
}: IBranchRunApprovalContainer) => {
  const { org } = useOrg()
  const orgId = org?.id ?? ''
  const [, setSearchParams] = useSearchParams()

  const openStep = useCallback(
    (stepId?: string) => {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev)
          next.set('workflow', run.id ?? '')
          if (stepId) next.set('step', stepId)
          return next
        },
        { replace: true }
      )
    },
    [run.id, setSearchParams]
  )

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
        onReview: () => openStep(step.id),
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
