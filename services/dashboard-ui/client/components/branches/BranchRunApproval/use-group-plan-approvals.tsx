import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { GroupActionButton } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep/GroupApprovalActions'
import type { TTrackGroup } from '@/components/branches/BranchOverview/RolloutTrack'
import type { TInstallWorkflowStep } from '@/types'
import { BranchRunApproval } from './BranchRunApproval'
import { PlanDiffPanel } from './PlanDiffPanel'
import { getGroupName } from './use-group-plan-href'
import type { IGroupPlanApproval } from '@/components/branches/BranchOverview/RolloutGroupsCard'

type TApprovalRun = {
  id: string
  status?: { status?: string }
  steps?: TInstallWorkflowStep[]
}

const factsFor = (group?: TTrackGroup) => {
  if (!group) return undefined
  return Object.fromEntries(
    group.installs.map((install) => [
      install.id,
      {
        labels: install.labels,
        region: install.region,
        status: install.status,
        detail: install.detail,
        appliedConfigId: install.appliedConfigId,
      },
    ])
  )
}

export const useGroupPlanApprovals = (
  run: TApprovalRun | undefined,
  groups: TTrackGroup[]
): IGroupPlanApproval[] => {
  const { org } = useOrg()
  const { addPanel } = useSurfaces()
  if (!run || run.status?.status === 'cancelled') return []

  return (run.steps ?? [])
    .filter(
      (step) =>
        step.execution_type === 'approval' &&
        step.status?.status === 'approval-awaiting' &&
        !step.approval?.response &&
        step.approval?.id
    )
    .map((step) => {
      const groupName = getGroupName(step.name)
      const group = groups.find(
        (item) => item.name.toLowerCase() === groupName.toLowerCase()
      )
      return {
        groupName,
        banner: (
          <BranchRunApproval
            items={[
              {
                key: `${run.id}-${step.id}`,
                groupName,
                onReview: () => {
                  addPanel(
                    <PlanDiffPanel
                      step={step}
                      installFacts={factsFor(group)}
                    />,
                    `plan-diff-${step.id}`
                  )
                },
                actions: (
                  <GroupActionButton
                    action="approve"
                    target={{
                      orgId: org?.id ?? '',
                      workflowId: run.id,
                      stepId: step.id,
                      approvalId: step.approval!.id!,
                      groupName,
                    }}
                  />
                ),
              },
            ]}
          />
        ),
      }
    })
}
