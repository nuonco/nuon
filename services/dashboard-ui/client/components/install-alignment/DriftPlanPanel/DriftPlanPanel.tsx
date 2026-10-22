import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { PlanComponent } from '@/components/approvals/Plan'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { useInstall } from '@/hooks/use-install'
import { useInstallOverview } from '@/hooks/use-install-overview'
import { useInstallLink } from '@/hooks/use-install-path'
import { useOrg } from '@/hooks/use-org'
import { useQueryApprovalPlan } from '@/hooks/use-query-approval-plan'
import { useResolveInstallDrift } from '@/hooks/use-resolve-install-drift'
import type { TInstallAlignmentPlan, TWorkflowStep } from '@/types'
import { branchRunHref, isDrifted } from '../alignment-utils'
import { DriftPlanDetails } from './DriftPlanDetails'

const PREVIEWABLE_PLAN_TYPES = new Set([
  'terraform_plan',
  'helm_approval',
  'kubernetes_manifest_approval',
  'pulumi_plan',
])

const planStep = (plan: TInstallAlignmentPlan) =>
  ({
    id: plan.step_id,
    install_workflow_id: plan.workflow_id,
    execution_type: 'approval',
    finished: true,
    approval: { id: plan.approval_id, type: plan.type },
  }) as TWorkflowStep

export const DriftPlanPanel = ({
  resourceId,
  ...props
}: IPanel & { resourceId: string }) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const installLink = useInstallLink()
  const { data: overview } = useInstallOverview()
  const resource = overview?.alignment?.resources?.find(
    (item) => item.id === resourceId
  )
  const plan = resource?.plan
  const previewable = PREVIEWABLE_PLAN_TYPES.has(plan?.type ?? '')
  const contents = useQueryApprovalPlan({
    step: plan ? planStep(plan) : undefined,
    enabled: previewable,
  })
  const resolve = useResolveInstallDrift()
  const workflowHref = (workflowId?: string) =>
    workflowId
      ? installLink({
          installId: install?.id,
          appId: install?.app_id,
          suffix: `/workflows/${workflowId}`,
        })
      : undefined
  const selectedCommit = overview?.branch_tracking?.selected_commit
  const title =
    resource?.category === 'state-drift-detected'
      ? 'Drift scan plan'
      : resource?.kind === 'image'
        ? 'Sync plan'
        : 'Deployment plan'

  return (
    <Panel
      {...props}
      heading={`${resource?.name ?? 'Resource'} · ${title}`}
      size="half"
      className="!h-auto max-h-screen"
      data-testid="drift-plan-panel"
    >
      {!resource || !plan ? (
        <EmptyState
          variant="table"
          emptyTitle="Plan unavailable"
          emptyMessage={
            overview?.alignment?.error
              ? 'Resource alignment is unavailable. Close this panel and try again.'
              : 'This resource no longer has a recorded plan. Close this panel to see its latest state.'
          }
        />
      ) : (
        <DriftPlanDetails
          resource={resource}
          selectedCommit={selectedCommit}
          selectedCommitHref={branchRunHref({
            orgId: org?.id,
            appId: install?.app_id,
            commit: selectedCommit,
          })}
          planWorkflowHref={workflowHref(plan.workflow_id)}
          deploymentHref={workflowHref(resource.open_deployment?.id)}
          resolving={resolve.isPending}
          resolveError={resolve.error}
          resolveDisabled={
            !contents.isSuccess || !contents.plan || !!contents.error
          }
          onApproveAndResolve={
            isDrifted(resource) && previewable && plan.approval_id
              ? () =>
                  resolve.mutate({ resource, approvedPlanId: plan.approval_id })
              : undefined
          }
          plan={
            PREVIEWABLE_PLAN_TYPES.has(plan.type ?? '') ? (
              <PlanComponent
                step={planStep(plan)}
                plan={contents.plan}
                isLoading={contents.isLoading}
                error={contents.error}
              />
            ) : (
              <EmptyState
                variant="table"
                emptyTitle="Plan preview unavailable"
                emptyMessage="This plan type can't be previewed here. Open the planning deployment to review it."
              />
            )
          }
        />
      )}
      {contents.error ? (
        <div role="alert">
          <Text variant="subtext" theme="error">
            {contents.error.message}
          </Text>
        </div>
      ) : null}
    </Panel>
  )
}
