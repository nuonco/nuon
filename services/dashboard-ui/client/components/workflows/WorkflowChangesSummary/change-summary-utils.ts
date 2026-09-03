import type { TIconVariant } from '@/components/common/Icon'
import type {
  TStepChangeCounts,
  TStepChangePlanType,
  TStepChangeStatus,
  TStepChangeSummary,
} from '@/types'
import { WORKFLOW_BADGE_MAP, type TBadgeCfg } from '@/utils/workflow-utils'

export const PLAN_TYPE_META: Record<
  TStepChangePlanType,
  { icon: TIconVariant; label: string }
> = {
  terraform_plan: { icon: 'Terraform', label: 'Terraform' },
  pulumi_plan: { icon: 'Pulumi', label: 'Pulumi' },
  helm_approval: { icon: 'Helm', label: 'Helm' },
  kubernetes_manifest_approval: { icon: 'Kubernetes', label: 'Kubernetes' },
  app_branch_plan: { icon: 'GitBranchIcon', label: 'App branch' },
  install_creation: { icon: 'CubeIcon', label: 'Install' },
}

export const STATUS_META: Record<TStepChangeStatus, TBadgeCfg> = {
  'pending-approval': WORKFLOW_BADGE_MAP['approval-awaiting'],
  approved: WORKFLOW_BADGE_MAP['approved'],
  denied: WORKFLOW_BADGE_MAP['approval-denied'],
  applied: WORKFLOW_BADGE_MAP['success'],
  generating: { children: 'Generating', theme: 'info' },
  error: WORKFLOW_BADGE_MAP['error'],
}

export const emptyCounts = (): TStepChangeCounts => ({
  create: 0,
  update: 0,
  delete: 0,
  replace: 0,
  noop: 0,
})

export const hasChanges = (counts: TStepChangeCounts): boolean =>
  counts.create + counts.update + counts.delete + counts.replace > 0

export const sumCounts = (summaries: TStepChangeSummary[]): TStepChangeCounts =>
  summaries.reduce(
    (acc, summary) => ({
      create: acc.create + summary.counts.create,
      update: acc.update + summary.counts.update,
      delete: acc.delete + summary.counts.delete,
      replace: acc.replace + summary.counts.replace,
      noop: acc.noop + summary.counts.noop,
    }),
    emptyCounts()
  )

const pluralize = (count: number, word: string): string =>
  `${count} ${count === 1 ? word : `${word}s`}`

export const formatAggregate = (
  counts: TStepChangeCounts,
  changedStepCount: number
): string => {
  const parts: string[] = []
  if (counts.create > 0) parts.push(pluralize(counts.create, 'create'))
  if (counts.update > 0) parts.push(pluralize(counts.update, 'update'))
  if (counts.delete > 0) parts.push(pluralize(counts.delete, 'delete'))
  if (counts.replace > 0) parts.push(pluralize(counts.replace, 'replace'))

  const scope = pluralize(changedStepCount, 'component')
  return `${parts.join(' · ')} across ${scope}`
}
