import { stringify } from 'yaml'
import type { THelmPlan } from '@/types'
import { humanize } from '@/utils/string-utils'
import {
  emptyDiffSummary,
  normalizeDiffOperation,
  summarizeDiffSections,
  type IPlanDiffGroup,
  type IPlanDiffSummary,
} from '.'

export const HELM_DIFF_OPERATIONS = ['create', 'update', 'delete'] as const

const ANSI_ESCAPE = new RegExp(`${String.fromCharCode(27)}\\[[0-9;]*m`, 'g')

type THelmContentDiff = NonNullable<THelmPlan['helm_content_diff']>[number] & {
  before?: unknown
  after?: unknown
  error?: string
  entries?: Array<{
    type?: number
    payload?: string
    original?: unknown
    applied?: unknown
  }>
}

interface IHelmDiffContents {
  before: string
  after: string
  error?: string
}

const content = (value: unknown) => {
  if (typeof value === 'string') return value
  if (value === null || value === undefined) return ''
  return stringify(value).trimEnd()
}

const entryContents = (
  entries: THelmContentDiff['entries']
): IHelmDiffContents => {
  const before: string[] = []
  const after: string[] = []
  const errors: string[] = []

  entries?.forEach((entry) => {
    if (entry.type === 3) {
      if (entry.original !== undefined) before.push(content(entry.original))
      if (entry.applied !== undefined) after.push(content(entry.applied))
      if (
        entry.original === undefined &&
        entry.applied === undefined &&
        entry.payload
      ) {
        before.push(entry.payload)
        after.push(entry.payload)
      }
      return
    }
    if (entry.type === 4) {
      if (entry.payload) errors.push(entry.payload)
      return
    }
    if (entry.payload) {
      if (entry.type === 0 || entry.type === 1) before.push(entry.payload)
      if (entry.type === 0 || entry.type === 2) after.push(entry.payload)
    }
  })

  return {
    before: before.join('\n'),
    after: after.join('\n'),
    error: errors.length ? errors.join('\n') : undefined,
  }
}

const diffContents = (diff?: THelmContentDiff): IHelmDiffContents => {
  if (!diff) {
    return {
      before: '',
      after: '',
      error: 'Diff not available from planner',
    }
  }

  if (diff.error) {
    return {
      before: '',
      after: '',
      error: diff.error,
    }
  }

  if (diff.before !== undefined || diff.after !== undefined) {
    return {
      before: content(diff.before),
      after: content(diff.after),
    }
  }

  return entryContents(diff.entries)
}

const planSummary = (planText: string): IPlanDiffSummary | undefined => {
  const match = planText.match(
    /Plan:\s*(\d+)\s+to add,\s*(\d+)\s+to change,\s*(\d+)\s+to destroy/i
  )
  if (!match) return

  return {
    ...emptyDiffSummary(),
    create: Number(match[1]),
    update: Number(match[2]),
    delete: Number(match[3]),
  }
}

export const helmPlanDiff = (plan?: THelmPlan): IPlanDiffGroup => {
  const planText = plan?.plan ?? ''
  const diffs = (plan?.helm_content_diff ?? []) as THelmContentDiff[]
  const lines = planText.replace(ANSI_ESCAPE, '').split('\n')

  const sections = lines.flatMap((line, index) => {
    const match = line.match(
      /^([^,]+),\s*([^,]+),\s*([^(]+)\s*\(([^)]+)\)\s*to\s*be\s*([\w-]+)/
    )
    if (!match) return []

    const workspace = match[1]?.trim() ?? ''
    const release = match[2]?.trim() ?? ''
    const resource = match[3]?.trim() ?? ''
    const resourceType = match[4]?.trim() ?? ''
    const rawOperation = match[5]?.trim() ?? ''
    const operation = normalizeDiffOperation(rawOperation)
    if (!operation) return []

    const diff = diffs.find(
      (item) =>
        item?.kind === resource &&
        item?.name === release &&
        item?.namespace === workspace
    )
    const values = diffContents(diff)

    return [
      {
        id: `${workspace}/${release}/${resourceType}/${resource}/${index}`,
        title: release,
        description: `${resource} · ${resourceType} · ${workspace}`,
        operation,
        before: values.before,
        after: values.after,
        language: 'yaml',
        filename: `${release}.yaml`,
        searchable: [workspace, release, resource, resourceType, rawOperation],
        error: values.error,
      },
    ]
  })

  return {
    id: 'helm',
    title: 'Helm changes',
    description: plan?.op ? `Operation: ${humanize(plan.op)}` : undefined,
    sections,
    summary: planSummary(planText) ?? summarizeDiffSections(sections),
  }
}
