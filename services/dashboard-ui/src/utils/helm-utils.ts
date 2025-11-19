import type {
  THelmPlan,
  THelmPlanChange,
  THelmPlanSummary,
  THelmK8sChangeAction,
} from '@/types'

export function parseHelmPlan(plan: THelmPlan): {
  changes: THelmPlanChange[]
  summary: THelmPlanSummary
} {
  const { plan: planText, helm_content_diff: diffs } = plan

  const changes: THelmPlanChange[] = []
  const summary: THelmPlanSummary = { add: 0, change: 0, destroy: 0 }
  const lines = planText.replace(/\u001b\[\d+;?\d*m/g, '').split('\n')

  lines.forEach((line) => {
    const match = line.match(
      /^([^,]+),\s*([^,]+),\s*([^(]+)\s*\(([^)]+)\)\s*to\s*be\s*(\w+)/
    )
    if (match) {
      // Try to find a matching diff item
      const diff = diffs?.find(
        (d) =>
          d.kind === match[3].trim() &&
          d.name === match[2].trim() &&
          d.namespace === match[1].trim()
      )

      // Extract before/after from the diff entries if available
      const { before, after } = diff
        ? buildBeforeAfterStrings(diff.entries || [])
        : { before: null, after: null }

      changes.push({
        workspace: match[1].trim(),
        release: match[2].trim(),
        resource: match[3].trim(),
        resourceType: match[4].trim(),
        action: match[5].trim() as unknown as THelmK8sChangeAction,
        before: before,
        after: after,
      })
    }

    const summaryMatch = line.match(
      /Plan: (\d+) to add, (\d+) to change, (\d+) to destroy/
    )

    if (summaryMatch) {
      summary.add = parseInt(summaryMatch[1])
      summary.change = parseInt(summaryMatch[2])
      summary.destroy = parseInt(summaryMatch[3])
    }
  })
  return { changes, summary }
}

function buildBeforeAfterStrings(entries: any[]): {
  before: string | null
  after: string | null
} {
  const beforeLines: string[] = []
  const afterLines: string[] = []

  entries.forEach((entry) => {
    if (entry.type === 1) {
      // Before value (removal) - lines that existed before
      if (entry.payload) {
        beforeLines.push(`${entry.payload}`)
      }
    } else if (entry.type === 2) {
      // After value (addition) - lines that will exist after
      if (entry.payload) {
        afterLines.push(`${entry.payload}`)
      }
    }
  })

  return {
    before: beforeLines.length > 0 ? beforeLines.join('\n') : null,
    after: afterLines.length > 0 ? afterLines.join('\n') : null,
  }
}

export function getHelmOutputStatus(deployments: Record<string, any>): string {
  for (const namespace of Object.values(deployments)) {
    for (const deployment of Object.values(namespace as any)) {
      const status = (deployment as any).status
      const replicas = {
        desired: status?.replicas || 0,
        ready: status?.readyReplicas || 0,
        available: status?.availableReplicas || 0,
      }

      if (
        replicas.ready !== replicas.desired ||
        replicas.available !== replicas.desired
      ) {
        return 'pending'
      }
    }
  }
  return 'healthy'
}
