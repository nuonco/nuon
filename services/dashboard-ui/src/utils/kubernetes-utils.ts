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
  
  // Handle the new structure where the plan data is in k8s_content_diff
  const diffItems = plan?.k8s_content_diff || []
  
  diffItems.forEach((item) => {
    let action: THelmK8sChangeAction
    
    // Determine action type based on op and type
    if (item.op === 'delete') {
      action = 'destroyed'
      summary.destroy += 1
    } else if (item.op === 'apply') {
      // type: 1 = add, 2 = delete, 3 = change
      if (item.type === 1) {
        action = 'added'
        summary.add += 1
      } else if (item.type === 3) {
        action = 'changed'
        summary.change += 1
      } else if (item.type === 2) {
        action = 'destroyed'
        summary.destroy += 1
      } else {
        // Default to changed if type is present but unknown
        action = 'changed'
        summary.change += 1
      }
    } else {
      action = item.op as THelmK8sChangeAction
    }

    // Extract before/after from entries if available
    const before = item.entries?.[0]?.original || null
    const after = item.entries?.[0]?.applied || null

    changes.push({
      namespace: item.namespace,
      name: item.name,
      resource: item.kind,
      resourceType: item.api,
      action: action,
      before: before,
      after: after,
    })
  })

  return { changes, summary }
}
