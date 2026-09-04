import type { TPulumiPlan, TPulumiResourceChange } from '@/types'
import {
  MISSING_DIFF_ERROR,
  NO_OP_NOTE,
  emptyDiffSummary,
  normalizeDiffOperation,
  type IPlanDiffDiagnostic,
  type IPlanDiffGroup,
  type IPlanDiffSection,
  type TDiffOperation,
} from '.'

export const PULUMI_DIFF_OPERATIONS = [
  'create',
  'update',
  'replace',
  'delete',
  'read',
  'no-op',
] as const

export const PULUMI_DEFAULT_DIFF_OPERATIONS = [
  'create',
  'update',
  'replace',
  'delete',
] as const

const NOTES: Partial<Record<TDiffOperation, string>> = {
  read: 'Pulumi will read this resource from the provider.',
  'no-op': NO_OP_NOTE,
}

const sortValue = (value: unknown): unknown => {
  if (Array.isArray(value)) return value.map(sortValue)
  if (!value || typeof value !== 'object') return value

  return Object.fromEntries(
    Object.entries(value)
      .sort(([left], [right]) => left.localeCompare(right))
      .map(([key, entry]) => [key, sortValue(entry)])
  )
}

const serialize = (value: unknown) =>
  value === undefined ? '' : JSON.stringify(sortValue(value), null, 2)

const contentsFor = (change: TPulumiResourceChange) => {
  if (change.old_inputs !== undefined || change.new_inputs !== undefined) {
    return {
      before: serialize(change.old_inputs),
      after: serialize(change.new_inputs),
    }
  }

  return {
    before: '',
    after: serialize(change.detailed_diff),
  }
}

const sectionFor = (
  change: TPulumiResourceChange,
  index: number
): IPlanDiffSection | undefined => {
  const operation = normalizeDiffOperation(change.action)
  if (!operation) return undefined

  const note = NOTES[operation]
  const contents = note ? { before: '', after: '' } : contentsFor(change)
  const diffs = change.diffs?.join(', ')
  const section: IPlanDiffSection = {
    id: `${change.urn}/${change.action}/${index}`,
    title: change.type,
    description: [change.name, diffs].filter(Boolean).join(' · '),
    operation,
    before: contents.before,
    after: contents.after,
    language: 'json',
    filename: `${change.name}.json`,
    searchable: [
      change.type,
      change.name,
      change.urn,
      change.action,
      ...(change.diffs ?? []),
    ],
    note,
  }

  return section.note || section.before || section.after
    ? section
    : { ...section, error: MISSING_DIFF_ERROR }
}

const diagnosticFor = (
  message: string,
  index: number
): IPlanDiffDiagnostic => ({
  id: `pulumi-diagnostic-${index}`,
  message,
  severity: message.startsWith('error')
    ? 'error'
    : message.startsWith('warning')
      ? 'warning'
      : 'info',
})

export const pulumiPlanDiff = (plan?: TPulumiPlan): IPlanDiffGroup => {
  const sections = (plan?.resource_changes ?? []).flatMap((change, index) => {
    const section = sectionFor(change, index)
    return section ? [section] : []
  })
  const summary = Object.entries(plan?.change_summary ?? {}).reduce(
    (counts, [action, count]) => {
      const operation = normalizeDiffOperation(action)
      if (operation) counts[operation] += count
      return counts
    },
    emptyDiffSummary()
  )

  return {
    id: 'pulumi-resources',
    title: 'Pulumi preview',
    searchPlaceholder: 'Search by type, name, or URN',
    sections,
    summary,
    diagnostics: plan?.diagnostics?.map(diagnosticFor),
  }
}
