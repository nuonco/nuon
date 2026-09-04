export default {
  title: 'Workflows/WorkflowChangesSummary',
}

import { TerraformDiff } from '@/components/approvals/plan-diffs/terraform/TerraformDiff'
import { Banner } from '@/components/common/Banner'
import type { TStepChangeSummary } from '@/types'
import { WorkflowChangesSummary } from './WorkflowChangesSummary'
import { PLAN_TYPE_META } from './change-summary-utils'

const noCounts = { create: 0, update: 0, delete: 0, replace: 0, noop: 0 }

const mk = (over: Partial<TStepChangeSummary>): TStepChangeSummary => ({
  step_id: `step-${Math.random().toString(36).slice(2, 9)}`,
  approval_id: `approval-${Math.random().toString(36).slice(2, 9)}`,
  step_name: 'Sync and plan',
  component_name: 'api',
  plan_type: 'terraform_plan',
  status: 'pending-approval',
  counts: noCounts,
  has_detail: true,
  ...over,
})

const smallTerraformPlan = {
  resource_changes: [
    {
      address: 'aws_s3_bucket.assets',
      type: 'aws_s3_bucket',
      name: 'assets',
      change: { actions: ['create'], before: null, after: { bucket: 'assets' } },
    },
    {
      address: 'aws_instance.web',
      type: 'aws_instance',
      name: 'web',
      change: {
        actions: ['update'],
        before: { instance_type: 't3.micro' },
        after: { instance_type: 't3.small' },
      },
    },
  ],
  output_changes: {},
}

const largeTerraformPlan = {
  resource_changes: Array.from({ length: 120 }, (_, i) => {
    const action = ['create', 'update', 'delete'][i % 3]
    return {
      address: `aws_instance.node_${i}`,
      type: 'aws_instance',
      name: `node_${i}`,
      change: {
        actions: [action],
        before: action === 'create' ? null : { tags: { index: i - 1 } },
        after: action === 'delete' ? null : { tags: { index: i } },
      },
    }
  }),
  output_changes: {},
}

const renderMockDetail = (summary: TStepChangeSummary) => {
  if (summary.plan_type === 'terraform_plan') {
    return <TerraformDiff plan={smallTerraformPlan as any} />
  }
  return (
    <Banner theme="neutral">
      {PLAN_TYPE_META[summary.plan_type].label} diff renders here on expand.
    </Banner>
  )
}

export const MixedTypes = () => (
  <WorkflowChangesSummary
    renderDetail={renderMockDetail}
    summaries={[
      mk({
        step_name: 'Sync and plan api',
        component_name: 'api',
        plan_type: 'terraform_plan',
        status: 'pending-approval',
        counts: { create: 3, update: 2, delete: 0, replace: 1, noop: 4 },
      }),
      mk({
        step_name: 'Sync and plan ingress',
        component_name: 'ingress',
        plan_type: 'helm_approval',
        status: 'pending-approval',
        counts: { create: 0, update: 4, delete: 0, replace: 0, noop: 1 },
      }),
      mk({
        step_name: 'Sync and plan configmaps',
        component_name: 'configmaps',
        plan_type: 'kubernetes_manifest_approval',
        status: 'approved',
        counts: { create: 2, update: 0, delete: 1, replace: 0, noop: 0 },
      }),
      mk({
        step_name: 'Sync and plan platform',
        component_name: 'platform',
        plan_type: 'pulumi_plan',
        status: 'pending-approval',
        counts: { create: 0, update: 6, delete: 2, replace: 0, noop: 3 },
      }),
    ]}
  />
)

export const DriftScanClean = () => (
  <WorkflowChangesSummary
    renderDetail={renderMockDetail}
    summaries={Array.from({ length: 8 }, (_, i) =>
      mk({
        step_id: `clean-${i}`,
        step_name: `Sync and plan component-${i}`,
        component_name: `component-${i}`,
        status: 'applied',
        counts: noCounts,
      })
    )}
  />
)

export const MostlyNoOp = () => (
  <WorkflowChangesSummary
    renderDetail={renderMockDetail}
    summaries={[
      mk({
        step_id: 'changed-1',
        step_name: 'Sync and plan api',
        component_name: 'api',
        counts: { create: 1, update: 2, delete: 0, replace: 0, noop: 8 },
      }),
      mk({
        step_id: 'changed-2',
        step_name: 'Sync and plan worker',
        component_name: 'worker',
        plan_type: 'helm_approval',
        counts: { create: 0, update: 1, delete: 0, replace: 0, noop: 2 },
      }),
      mk({
        step_id: 'changed-3',
        step_name: 'Sync and plan db',
        component_name: 'db',
        counts: { create: 0, update: 0, delete: 1, replace: 0, noop: 5 },
      }),
      ...Array.from({ length: 17 }, (_, i) =>
        mk({
          step_id: `noop-${i}`,
          step_name: `Sync and plan component-${i}`,
          component_name: `component-${i}`,
          status: 'applied',
          counts: noCounts,
        })
      ),
    ]}
  />
)

export const LargePlan = () => (
  <WorkflowChangesSummary
    renderDetail={() => <TerraformDiff plan={largeTerraformPlan as any} />}
    summaries={[
      mk({
        step_id: 'large-1',
        step_name: 'Sync and plan monolith',
        component_name: 'monolith',
        status: 'pending-approval',
        counts: { create: 40, update: 40, delete: 40, replace: 0, noop: 0 },
      }),
    ]}
  />
)

export const InProgress = () => (
  <WorkflowChangesSummary
    renderDetail={renderMockDetail}
    summaries={[
      mk({
        step_id: 'gen-1',
        step_name: 'Sync and plan api',
        component_name: 'api',
        status: 'generating',
      }),
      mk({
        step_id: 'gen-2',
        step_name: 'Sync and plan worker',
        component_name: 'worker',
        plan_type: 'helm_approval',
        status: 'generating',
      }),
      mk({
        step_id: 'done-1',
        step_name: 'Sync and plan ingress',
        component_name: 'ingress',
        status: 'applied',
        counts: { create: 2, update: 1, delete: 0, replace: 0, noop: 0 },
      }),
    ]}
  />
)

export const PartialFailure = () => (
  <WorkflowChangesSummary
    renderDetail={renderMockDetail}
    summaries={[
      mk({
        step_id: 'ok-1',
        step_name: 'Sync and plan api',
        component_name: 'api',
        counts: { create: 1, update: 3, delete: 0, replace: 0, noop: 2 },
      }),
      mk({
        step_id: 'err-1',
        step_name: 'Sync and plan worker',
        component_name: 'worker',
        plan_type: 'pulumi_plan',
        status: 'error',
      }),
      mk({
        step_id: 'err-2',
        step_name: 'Sync and plan db',
        component_name: 'db',
        status: 'error',
      }),
    ]}
  />
)

export const ApprovalStateMix = () => (
  <WorkflowChangesSummary
    renderDetail={renderMockDetail}
    summaries={[
      mk({
        step_id: 'pending-1',
        step_name: 'Sync and plan api',
        component_name: 'api',
        status: 'pending-approval',
        counts: { create: 2, update: 0, delete: 0, replace: 0, noop: 1 },
      }),
      mk({
        step_id: 'denied-1',
        step_name: 'Sync and plan worker',
        component_name: 'worker',
        plan_type: 'helm_approval',
        status: 'denied',
        counts: { create: 0, update: 0, delete: 3, replace: 0, noop: 0 },
      }),
      mk({
        step_id: 'applied-1',
        step_name: 'Sync and plan ingress',
        component_name: 'ingress',
        status: 'applied',
        counts: { create: 1, update: 1, delete: 0, replace: 0, noop: 2 },
      }),
    ]}
  />
)

export const BranchRun = () => (
  <WorkflowChangesSummary
    renderDetail={renderMockDetail}
    summaries={[
      mk({
        step_id: 'branch-1',
        step_name: 'Plan app branch',
        component_name: 'main',
        plan_type: 'app_branch_plan',
        status: 'pending-approval',
        counts: { create: 2, update: 1, delete: 0, replace: 0, noop: 0 },
      }),
      mk({
        step_id: 'install-1',
        step_name: 'Create install',
        component_name: undefined,
        plan_type: 'install_creation',
        status: 'pending-approval',
        counts: noCounts,
        has_detail: false,
      }),
    ]}
  />
)

export const SingleStepTrivial = () => (
  <WorkflowChangesSummary
    renderDetail={renderMockDetail}
    summaries={[
      mk({
        step_id: 'only-1',
        step_name: 'Sync and plan api',
        component_name: 'api',
        status: 'pending-approval',
        counts: { create: 1, update: 0, delete: 0, replace: 0, noop: 0 },
      }),
    ]}
  />
)

export const Loading = () => (
  <WorkflowChangesSummary
    renderDetail={renderMockDetail}
    summaries={[]}
    isLoading
    loadingSteps={[
      { stepName: 'Sync and plan api', componentName: 'api', planType: 'terraform_plan' },
      { stepName: 'Sync and plan worker', componentName: 'worker', planType: 'helm_approval' },
      { stepName: 'Sync and plan db', componentName: 'db', planType: 'terraform_plan' },
    ]}
  />
)
