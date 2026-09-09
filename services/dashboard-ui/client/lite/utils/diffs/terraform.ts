import type { TTerraformChangeAction, TTerraformPlan } from '@/types'
import {
  MISSING_DIFF_ERROR,
  NO_OP_NOTE,
  emptyDiffSummary,
  normalizeDiffOperation,
  terraformDiff,
  type IPlanDiffGroup,
  type IPlanDiffSection,
  type TDiffOperation,
} from '.'

export const TERRAFORM_DIFF_OPERATIONS = [
  'create',
  'update',
  'replace',
  'delete',
  'read',
  'no-op',
] as const

export const TERRAFORM_DEFAULT_DIFF_OPERATIONS = [
  'create',
  'update',
  'replace',
  'delete',
] as const

export const TERRAFORM_READ_NOTE =
  'Terraform will refresh this resource from the provider.'

const NOTES: Partial<Record<TDiffOperation, string>> = {
  read: TERRAFORM_READ_NOTE,
  'no-op': NO_OP_NOTE,
}

export interface ITerraformPlanDiff {
  drift: IPlanDiffGroup
  resources: IPlanDiffGroup
  outputs: IPlanDiffGroup
}

type TTerraformChange = {
  actions?: TTerraformChangeAction[] | string[]
  before?: unknown
  after?: unknown
  after_unknown?: unknown
  after_sensitive?: unknown
  before_sensitive?: unknown
}

type TTerraformResource = {
  address?: string
  module_address?: string | null
  type?: string
  name?: string
  change?: TTerraformChange
}

const isEmptyCollection = (value: unknown) => {
  if (value == null) return true
  if (Array.isArray(value)) return value.length === 0
  if (typeof value === 'object') return Object.keys(value).length === 0
  return false
}

const semanticEqual = (left: unknown, right: unknown): boolean => {
  if (left === right) return true
  if (isEmptyCollection(left) && isEmptyCollection(right)) return true
  if (left == null || right == null) return false
  if (typeof left !== typeof right) return false

  if (Array.isArray(left) && Array.isArray(right)) {
    if (left.length !== right.length) return false
    return left.every((item, index) => semanticEqual(item, right[index]))
  }

  if (typeof left === 'object' && typeof right === 'object') {
    const leftRecord = left as Record<string, unknown>
    const rightRecord = right as Record<string, unknown>
    const keys = new Set([
      ...Object.keys(leftRecord),
      ...Object.keys(rightRecord),
    ])
    return [...keys].every((key) =>
      semanticEqual(leftRecord[key], rightRecord[key])
    )
  }

  return false
}

const isReplaceActions = (actions: string[]) =>
  actions.length === 2 &&
  ((actions[0] === 'delete' && actions[1] === 'create') ||
    (actions[0] === 'create' && actions[1] === 'delete'))

const operationsFor = (actions: string[]): TDiffOperation[] => {
  if (isReplaceActions(actions)) return ['replace']
  return actions.flatMap((action) => {
    const operation = normalizeDiffOperation(action)
    return operation ? [operation] : []
  })
}

const serialized = (change: TTerraformChange | undefined, filename?: string) =>
  terraformDiff({
    before: change?.before,
    after: change?.after,
    beforeSensitive: change?.before_sensitive,
    afterSensitive: change?.after_sensitive,
    afterUnknown: change?.after_unknown,
    filename,
  })

const described = (section: IPlanDiffSection): IPlanDiffSection => {
  const note = NOTES[section.operation]
  if (note) return { ...section, before: '', after: '', note }

  return section.error || section.before || section.after
    ? section
    : { ...section, error: MISSING_DIFF_ERROR }
}

const resourceSections = (
  items: TTerraformResource[] | undefined,
  skipCosmetic: boolean
): IPlanDiffSection[] =>
  (items ?? []).flatMap((item, index) => {
    const change = item?.change
    if (skipCosmetic && semanticEqual(change?.before, change?.after)) return []

    const address = item?.address ?? `resource-${index + 1}`
    const name = item?.name ?? address
    const type = item?.type ?? 'unknown'
    const module = item?.module_address ?? undefined
    const filename = `${name}.tf`
    const values = serialized(change, filename)

    return operationsFor(change?.actions ?? []).map((operation, actionIndex) =>
      described({
        id: `${address}/${operation}/${index}/${actionIndex}`,
        title: address,
        description: module
          ? `${type} · ${name} · ${module}`
          : `${type} · ${name}`,
        operation,
        before: values.before,
        after: values.after,
        language: values.language,
        filename,
        searchable: [address, name, type, module ?? '', operation],
      })
    )
  })

const outputSections = (
  outputs: TTerraformPlan['output_changes'] | undefined
): IPlanDiffSection[] =>
  Object.entries(outputs ?? {}).flatMap(([name, change], index) => {
    const filename = `${name}.tf`
    const values = serialized(change, filename)
    const knownAfterApply = change?.after_unknown ? 'known after apply' : ''

    return operationsFor(change?.actions ?? []).map((operation, actionIndex) =>
      described({
        id: `output/${name}/${operation}/${index}/${actionIndex}`,
        title: name,
        description:
          [
            change?.before_sensitive ? 'sensitive before' : null,
            change?.after_sensitive ? 'sensitive after' : null,
            knownAfterApply,
          ]
            .filter(Boolean)
            .join(' · ') || undefined,
        operation,
        before: values.before,
        after: values.after,
        language: values.language,
        filename,
        searchable: [name, operation],
      })
    )
  })

const group = (
  id: string,
  title: string,
  searchPlaceholder: string,
  sections: IPlanDiffSection[]
): IPlanDiffGroup => ({
  id,
  title,
  searchPlaceholder,
  sections,
  summary: sections.reduce((counts, section) => {
    counts[section.operation] += 1
    return counts
  }, emptyDiffSummary()),
})

export const terraformPlanDiff = (
  plan?: TTerraformPlan
): ITerraformPlanDiff => ({
  drift: group(
    'terraform-drift',
    'Resource drift',
    'Search by address, resource, or name',
    resourceSections(plan?.resource_drift, true)
  ),
  resources: group(
    'terraform-resources',
    'Resource changes',
    'Search by address, resource, or name',
    resourceSections(plan?.resource_changes, false)
  ),
  outputs: group(
    'terraform-outputs',
    'Output changes',
    'Search outputs by name',
    outputSections(plan?.output_changes)
  ),
})
