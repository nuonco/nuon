import type {
  TKubernetesPlan,
  TKubernetesPlanChange,
  TKubernetesPlanSummary,
  TKubernetesPlanError,
  THelmK8sChangeAction,
} from '@/types'
import { diffLines } from '@/utils/code-utils'

export function parseKubernetesPlan(plan: TKubernetesPlan): {
  changes: TKubernetesPlanChange[]
  errors: TKubernetesPlanError[]
  summary: TKubernetesPlanSummary
} {
  const changes: TKubernetesPlanChange[] = []
  const errors: TKubernetesPlanError[] = []
  const summary: TKubernetesPlanSummary = { add: 0, change: 0, destroy: 0 }

  const diffItems = plan?.k8s_content_diff || []

  diffItems.forEach((item) => {
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
        action = 'changed'
        summary.change += 1
      }
    } else {
      action = item.op as THelmK8sChangeAction
    }

    const { before, after } = buildBeforeAfterStrings(item.entries || [])

    changes.push({
      namespace: item.namespace,
      name: item.name,
      resource: item.kind,
      resourceType: item.api,
      action: action,
      before: before,
      after: after,
      diff: diffLines(before, after),
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
