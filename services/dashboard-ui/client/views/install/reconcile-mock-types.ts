// Mockup-only types for the install out-of-sync reconcile feature (ladle stories only).
//
// This describes a PROPOSED extension of the `service.InstallReconciliationResponse`
// API shape introduced in #3245. Nothing here is wired to a real endpoint — it exists
// so step-1 ladle stories can render the feature before the API/backend work lands.
//
// Proposed additions:
// - `status` gains `out_of_sync`
// - `out_of_sync: boolean` (so `behind` + `out_of_sync` can combine)
// - `failed_workflow_id?: string` — the update workflow that partially failed
// - `deploy_plan: TMockDeployPlanItem[]` — ordered list of what reconcile will deploy
import type { TInstallReconciliation } from '@/types'

export type TMockReconcileStatus = TInstallReconciliation['status'] | 'out_of_sync'

export type TMockDeployPlanKind = 'component' | 'image' | 'sandbox' | 'stack'

export type TMockDeployPlanReason =
  | 'deploy_failed'
  | 'skipped'
  | 'older_build_deployed'
  | 'config_changed'
  | 'not_attempted'
  | 'dependency_changed'

export interface TMockDeployPlanItem {
  kind: TMockDeployPlanKind
  id: string
  name: string
  order: number
  reason: TMockDeployPlanReason
  reason_message: string
}

export interface TMockInstallReconciliation
  extends Omit<TInstallReconciliation, 'status'> {
  status: TMockReconcileStatus
  out_of_sync?: boolean
  failed_workflow_id?: string
  deploy_plan?: TMockDeployPlanItem[]
}

export const DEPLOY_PLAN_REASON_LABELS: Record<TMockDeployPlanReason, string> = {
  deploy_failed: 'Deploy failed',
  skipped: 'Step skipped',
  older_build_deployed: 'Older build deployed',
  config_changed: 'Config changed',
  not_attempted: 'Not attempted',
  dependency_changed: 'Dependency changed',
}

export const DEPLOY_PLAN_KIND_LABELS: Record<TMockDeployPlanKind, string> = {
  component: 'Component',
  image: 'Image',
  sandbox: 'Sandbox',
  stack: 'Stack',
}
