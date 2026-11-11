import { GroupApprovalActions } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep/GroupApprovalActions'
import { BranchRunApproval } from './BranchRunApproval'

export default {
  title: 'Features / Branches / Branch run approval',
}

const target = (groupName: string, stepId: string) => ({
  orgId: 'org-mock-001',
  workflowId: 'wf-1',
  stepId,
  approvalId: `apr-${stepId}`,
  groupName,
})

export const Default = () => (
  <BranchRunApproval
    items={[
      {
        key: 'step-1',
        groupName: 'tag',
        onReview: () => {},
        actions: <GroupApprovalActions target={target('tag', 'step-1')} />,
      },
    ]}
  />
)

export const MultipleGroups = () => (
  <BranchRunApproval
    items={[
      {
        key: 'step-1',
        groupName: 'uat',
        onReview: () => {},
        actions: <GroupApprovalActions target={target('uat', 'step-1')} />,
      },
      {
        key: 'step-2',
        groupName: 'production',
        onReview: () => {},
        actions: (
          <GroupApprovalActions target={target('production', 'step-2')} />
        ),
      },
    ]}
  />
)

export const Empty = () => <BranchRunApproval items={[]} />
