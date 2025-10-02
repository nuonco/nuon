import type {
  TKubernetesPlan,
  TKubernetesPlanChange,
  TKubernetesPlanSummary,
  THelmK8sChangeAction,
} from '@/types'

export function parseKubernetesPlan(plan: TKubernetesPlan): {
  changes: TKubernetesPlanChange[]
  summary: TKubernetesPlanSummary
} {
  const changes: TKubernetesPlanChange[] = []
  const summary: TKubernetesPlanSummary = { add: 0, change: 0, destroy: 0 }

  plan.forEach((item) => {
    let action: string
    // Determine action type from op and presence of before/after
    if (item.op === 'delete') {
      action = 'destroyed'
      summary.destroy += 1
    } else if (item.op === 'apply') {
      if (!item.before && item.after) {
        action = 'added'
        summary.add += 1
      } else if (item.before && item.after) {
        action = 'changed'
        summary.change += 1
      } else if (item.before && !item.after) {
        action = 'destroyed'
        summary.destroy += 1
      } else {
        action = item.op // fallback
      }
    } else {
      action = item.op
    }

    changes.push({
      namespace: item.namespace,
      name: item.name,
      resource: item.group_version_kind.Kind,
      resourceType: item.group_version_kind.Version,
      action: action as THelmK8sChangeAction,
      before: item.before,
      after: item.after,
    })
  })

  return { changes, summary }
}
