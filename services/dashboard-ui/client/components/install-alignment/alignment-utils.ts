import { ACTIVE_DEPLOYMENT_STATUSES } from '@/components/installs/DeploymentDetail/deployment-progress'
import type {
  TInstallAlignmentResource,
  TInstallAlignmentResourceKind,
  TInstallOverviewCommit,
} from '@/types'

export const driftExplanation = (resource: TInstallAlignmentResource) =>
  resource.explanation?.summary?.trim() ||
  (resource.category === 'state-drift-detected'
    ? 'A drift check found differences between live resources and what was last applied.'
    : 'We couldn’t determine why these changes haven’t taken effect.')

export const explanationEvidence = (
  resource: TInstallAlignmentResource
): { workflowId: string; stepId?: string } | undefined => {
  const explanation = resource.explanation
  if (explanation?.workflow_id) {
    return { workflowId: explanation.workflow_id, stepId: explanation.step_id }
  }
  if (
    resource.category === 'state-drift-detected' &&
    resource.plan?.plan_only &&
    resource.plan?.workflow_id
  ) {
    return {
      workflowId: resource.plan.workflow_id,
      stepId: resource.plan.step_id,
    }
  }
  return undefined
}

export type TDriftFilter =
  | 'all'
  | 'changes-not-applied'
  | 'state-drift-detected'

export const hasOngoingDeployment = (resource?: TInstallAlignmentResource) =>
  !!resource?.open_deployment?.id &&
  ACTIVE_DEPLOYMENT_STATUSES.has(resource.open_deployment.status ?? '') &&
  !(
    resource.kind === 'image' &&
    resource.open_deployment.status === 'failed-pending-retry'
  )

export const isDrifted = (resource: TInstallAlignmentResource) =>
  resource.category === 'changes-not-applied' ||
  resource.category === 'state-drift-detected'

export const matchesSearch = (
  resource: TInstallAlignmentResource,
  search?: string
) => {
  const query = search?.trim().toLowerCase()
  if (!query) return true
  return [resource.name, resource.id, resource.component_id]
    .filter(Boolean)
    .some((value) => value!.toLowerCase().includes(query))
}

export const resourcesOfKinds = (
  resources: TInstallAlignmentResource[] | undefined,
  kinds: TInstallAlignmentResourceKind[]
) =>
  (resources ?? []).filter(
    (resource) => !!resource.kind && kinds.includes(resource.kind)
  )

export const filterDrifted = (
  drifted: TInstallAlignmentResource[],
  driftFilter: TDriftFilter
) =>
  drifted.filter(
    (resource) => driftFilter === 'all' || resource.category === driftFilter
  )

export const driftCounts = (drifted: TInstallAlignmentResource[]) => {
  const unapplied = drifted.filter(
    (resource) => resource.category === 'changes-not-applied'
  )
  return {
    all: drifted.length,
    unapplied: unapplied.length,
    stateDrift: drifted.filter(
      (resource) => resource.category === 'state-drift-detected'
    ).length,
  }
}

export const shortSha = (sha?: string) => sha?.slice(0, 7)

export const branchRunHref = ({
  appId,
  commit,
  orgId,
}: {
  appId?: string
  commit?: TInstallOverviewCommit
  orgId?: string
}) =>
  orgId && appId && commit?.branch_id && commit?.workflow_id
    ? `/${orgId}/apps/${appId}/branches/${commit.branch_id}/runs/${commit.workflow_id}`
    : undefined
