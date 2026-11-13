export default {
  title: 'Features / Branches / Approvals / Plan diff panel',
}

import type { TInstallWorkflowStep } from '@/types'
import { PlanDiffPanel } from './PlanDiffPanel'

const step = {
  id: 'step-plan-primary',
  name: 'plan install group: Primary region',
  execution_type: 'approval',
  install_workflow_id: 'wf-1',
  status: { status: 'approval-awaiting' },
  approval: { id: 'apr-primary', type: 'app_branch_plan' },
} as unknown as TInstallWorkflowStep

export const AwaitingApproval = () => (
  <PlanDiffPanel
    isVisible
    panelId="plan-diff-story"
    step={step}
    orgId="org-acme"
    workflowId="wf-1"
  />
)

export const Approved = () => (
  <PlanDiffPanel
    isVisible
    panelId="plan-diff-story"
    step={
      {
        ...step,
        status: { status: 'in-progress' },
        approval: { ...step.approval, response: { type: 'approve' } },
      } as unknown as TInstallWorkflowStep
    }
    orgId="org-acme"
    workflowId="wf-1"
  />
)
