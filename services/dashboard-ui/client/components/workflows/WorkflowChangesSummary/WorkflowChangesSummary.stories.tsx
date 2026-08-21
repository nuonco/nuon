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
  stepId: `step-${Math.random().toString(36).slice(2, 9)}`,
  stepName: 'Sync and plan',
  componentName: 'api',
  planType: 'terraform_plan',
  status: 'pending-approval',
  counts: noCounts,
  hasDetail: true,
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
  if (summary.planType === 'terraform_plan') {
    return <TerraformDiff plan={smallTerraformPlan as any} />
  }
  return (
    <Banner theme="neutral">
      {PLAN_TYPE_META[summary.planType].label} diff renders here on expand.
    </Banner>
  )
}

export const MixedTypes = () => (
  <WorkflowChangesSummary
    renderDetail={renderMockDetail}
    summaries={[
      mk({
        stepName: 'Sync and plan api',
        componentName: 'api',
        planType: 'terraform_plan',
        status: 'pending-approval',
        counts: { create: 3, update: 2, delete: 0, replace: 1, noop: 4 },
      }),
      mk({
        stepName: 'Sync and plan ingress',
        componentName: 'ingress',
        planType: 'helm_approval',
        status: 'pending-approval',
        counts: { create: 0, update: 4, delete: 0, replace: 0, noop: 1 },
      }),
      mk({
        stepName: 'Sync and plan configmaps',
        componentName: 'configmaps',
        planType: 'kubernetes_manifest_approval',
        status: 'approved',
        counts: { create: 2, update: 0, delete: 1, replace: 0, noop: 0 },
      }),
      mk({
        stepName: 'Sync and plan platform',
        componentName: 'platform',
        planType: 'pulumi_plan',
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
        stepId: `clean-${i}`,
        stepName: `Sync and plan component-${i}`,
        componentName: `component-${i}`,
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
        stepId: 'changed-1',
        stepName: 'Sync and plan api',
        componentName: 'api',
        counts: { create: 1, update: 2, delete: 0, replace: 0, noop: 8 },
      }),
      mk({
        stepId: 'changed-2',
        stepName: 'Sync and plan worker',
        componentName: 'worker',
        planType: 'helm_approval',
        counts: { create: 0, update: 1, delete: 0, replace: 0, noop: 2 },
      }),
      mk({
        stepId: 'changed-3',
        stepName: 'Sync and plan db',
        componentName: 'db',
        counts: { create: 0, update: 0, delete: 1, replace: 0, noop: 5 },
      }),
      ...Array.from({ length: 17 }, (_, i) =>
        mk({
          stepId: `noop-${i}`,
          stepName: `Sync and plan component-${i}`,
          componentName: `component-${i}`,
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
        stepId: 'large-1',
        stepName: 'Sync and plan monolith',
        componentName: 'monolith',
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
        stepId: 'gen-1',
        stepName: 'Sync and plan api',
        componentName: 'api',
        status: 'generating',
      }),
      mk({
        stepId: 'gen-2',
        stepName: 'Sync and plan worker',
        componentName: 'worker',
        planType: 'helm_approval',
        status: 'generating',
      }),
      mk({
        stepId: 'done-1',
        stepName: 'Sync and plan ingress',
        componentName: 'ingress',
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
        stepId: 'ok-1',
        stepName: 'Sync and plan api',
        componentName: 'api',
        counts: { create: 1, update: 3, delete: 0, replace: 0, noop: 2 },
      }),
      mk({
        stepId: 'err-1',
        stepName: 'Sync and plan worker',
        componentName: 'worker',
        planType: 'pulumi_plan',
        status: 'error',
      }),
      mk({
        stepId: 'err-2',
        stepName: 'Sync and plan db',
        componentName: 'db',
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
        stepId: 'pending-1',
        stepName: 'Sync and plan api',
        componentName: 'api',
        status: 'pending-approval',
        counts: { create: 2, update: 0, delete: 0, replace: 0, noop: 1 },
      }),
      mk({
        stepId: 'denied-1',
        stepName: 'Sync and plan worker',
        componentName: 'worker',
        planType: 'helm_approval',
        status: 'denied',
        counts: { create: 0, update: 0, delete: 3, replace: 0, noop: 0 },
      }),
      mk({
        stepId: 'applied-1',
        stepName: 'Sync and plan ingress',
        componentName: 'ingress',
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
        stepId: 'branch-1',
        stepName: 'Plan app branch',
        componentName: 'main',
        planType: 'app_branch_plan',
        status: 'pending-approval',
        counts: { create: 2, update: 1, delete: 0, replace: 0, noop: 0 },
      }),
      mk({
        stepId: 'install-1',
        stepName: 'Create install',
        componentName: undefined,
        planType: 'install_creation',
        status: 'pending-approval',
        counts: noCounts,
        hasDetail: false,
      }),
    ]}
  />
)

export const SingleStepTrivial = () => (
  <WorkflowChangesSummary
    renderDetail={renderMockDetail}
    summaries={[
      mk({
        stepId: 'only-1',
        stepName: 'Sync and plan api',
        componentName: 'api',
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
