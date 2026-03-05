import type {
  TKubernetesPlan,
  TKubernetesPlanChange,
  TKubernetesPlanSummary,
  TKubernetesPlanError,
  THelmK8sChangeAction,
} from '@/types'

export function parseKubernetesPlan(plan: TKubernetesPlan): {
  changes: TKubernetesPlanChange[]
  errors: TKubernetesPlanError[]
  summary: TKubernetesPlanSummary
} {
  const changes: TKubernetesPlanChange[] = []
  const errors: TKubernetesPlanError[] = []
  const summary: TKubernetesPlanSummary = { add: 0, change: 0, destroy: 0 }

  // Handle the new structure where the plan data is in k8s_content_diff
  const diffItems = plan?.k8s_content_diff || []

  diffItems.forEach((item) => {
    // Skip items with errors - handle separately
    if (item.error) {
      errors.push({
        namespace: item.namespace,
        name: item.name,
        resource: item.kind,
        resourceType: item.api,
        error: item.error,
      })
      return
    }

    let action: THelmK8sChangeAction

    // Determine action type based on op and type
    if (item.op === 'delete') {
      action = 'destroyed'
      summary.destroy += 1
    } else if (item.op === 'apply') {
      if (item.type === 2) {
        action = 'added'
        summary.add += 1
      } else if (item.type === 3) {
        action = 'changed'
        summary.change += 1
      } else if (item.type === 1) {
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

    // Extract before/after from entries by building a formatted string
    const { before, after } = buildBeforeAfterStrings(item.entries || [])

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

  return { changes, errors, summary }
}

function buildBeforeAfterStrings(entries: any[]): {
  before: string | null
  after: string | null
} {
  const beforeLines: string[] = []
  const afterLines: string[] = []

  // Group entries by path to handle before/after pairs
  const pathGroups = new Map<string, { before?: string; after?: string }>()

  entries.forEach((entry) => {
    const path = entry.path

    // Handle content-based diffs (no path)
    if (!path) {
      if (entry.type === 1) {
        // Before value (removal)
        beforeLines.push(entry.payload || '')
      } else if (entry.type === 2) {
        // After value (addition)
        afterLines.push(entry.payload || '')
      }
      return
    }

    const existing = pathGroups.get(path) || {}

    if (entry.type === 1) {
      // Before value (removal)
      existing.before = entry.payload || null
    } else if (entry.type === 2) {
      // After value (addition)
      existing.after = entry.payload || null
    }

    pathGroups.set(path, existing)
  })

  // Build the formatted strings
  pathGroups.forEach((values, path) => {
    if (values.before !== undefined) {
      beforeLines.push(`${path}: ${values.before || ''}`)
    }
    if (values.after !== undefined) {
      afterLines.push(`${path}: ${values.after || ''}`)
    }
  })

  return {
    before: beforeLines.length > 0 ? beforeLines.join('\n') : null,
    after: afterLines.length > 0 ? afterLines.join('\n') : null,
  }
}
