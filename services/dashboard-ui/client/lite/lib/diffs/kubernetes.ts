import type { TKubernetesPlan, TKubernetesPlanItem } from '@/types'
import { humanize } from '@/utils/string-utils'
import {
  MISSING_DIFF_ERROR,
  emptyDiffSummary,
  normalizeDiffOperation,
  type IPlanDiffGroup,
  type IPlanDiffSection,
  type TDiffOperation,
} from '.'

export const KUBERNETES_DIFF_OPERATIONS = [
  'create',
  'update',
  'delete',
] as const

type TKubernetesEntry = Partial<
  NonNullable<TKubernetesPlanItem['entries']>[number]
> & {
  original?: unknown
  applied?: unknown
}

type TKubernetesContentDiff = Partial<Omit<TKubernetesPlanItem, 'entries'>> & {
  entries?: TKubernetesEntry[] | null
}

const operationFor = (item: TKubernetesContentDiff): TDiffOperation => {
  if (item?.op === 'apply') {
    if (item?.type === 1) return 'delete'
    if (item?.type === 2) return 'create'
    return 'update'
  }

  return normalizeDiffOperation(item?.op ?? '') ?? 'update'
}

const value = (entryValue: unknown) => {
  if (entryValue === null || entryValue === undefined) return ''
  return typeof entryValue === 'string'
    ? entryValue
    : JSON.stringify(entryValue)
}

const entryContents = (entries?: TKubernetesEntry[] | null) => {
  const before: string[] = []
  const after: string[] = []
  const errors: string[] = []

  entries?.forEach((entry) => {
    if (entry?.type === 4) {
      if (entry?.payload) errors.push(entry.payload)
      return
    }

    if (
      entry?.path &&
      (entry?.original !== undefined || entry?.applied !== undefined)
    ) {
      if (entry?.original !== undefined) {
        before.push(`${entry.path}: ${value(entry.original)}`)
      }
      if (entry?.applied !== undefined) {
        after.push(`${entry.path}: ${value(entry.applied)}`)
      }
      return
    }

    if (entry?.type === 3) {
      if (entry?.original !== undefined) before.push(value(entry.original))
      if (entry?.applied !== undefined) after.push(value(entry.applied))
      if (
        entry?.original === undefined &&
        entry?.applied === undefined &&
        entry?.payload
      ) {
        before.push(entry.payload)
        after.push(entry.payload)
      }
      return
    }

    const line = entry?.path
      ? `${entry.path}: ${entry.payload ?? ''}`
      : (entry.payload ?? '')

    if (entry?.type === 0 || entry?.type === 1) before.push(line)
    if (entry?.type === 0 || entry?.type === 2) after.push(line)
  })

  return {
    before: before.join('\n'),
    after: after.join('\n'),
    error: errors.length ? errors.join('\n') : undefined,
  }
}

const sectionFor = (
  item: TKubernetesContentDiff,
  index: number
): IPlanDiffSection => {
  const contents = entryContents(item?.entries)
  const name = item?.name ?? `resource-${index + 1}`
  const namespace = item?.namespace ?? 'unknown namespace'
  const kind = item?.kind ?? 'Unknown resource'
  const api = item?.api ?? 'unknown API'
  const errors = [item?.error, contents.error].filter(Boolean)
  if (!errors.length && !contents.before && !contents.after) {
    errors.push(MISSING_DIFF_ERROR)
  }

  return {
    id: `${namespace}/${api}/${kind}/${name}/${index}`,
    title: name,
    description: `${kind} · ${api} · ${namespace}`,
    operation: operationFor(item),
    before: contents.before,
    after: contents.after,
    language: 'yaml',
    filename: `${name}.yaml`,
    searchable: [
      name,
      namespace,
      kind,
      api,
      item?.resource ?? '',
      item?.op ?? '',
    ],
    error: errors.length ? errors.join('\n') : undefined,
  }
}

export const kubernetesPlanDiff = (plan?: TKubernetesPlan): IPlanDiffGroup => {
  const items = (plan?.k8s_content_diff ?? []) as TKubernetesContentDiff[]
  const sections = items.map(sectionFor)
  const summary = items.reduce((counts, item) => {
    if (!item?.error) counts[operationFor(item)] += 1
    return counts
  }, emptyDiffSummary())

  return {
    id: 'kubernetes',
    title: 'Kubernetes changes',
    description: plan?.op ? `Operation: ${humanize(plan.op)}` : undefined,
    searchPlaceholder: 'Search by name, resource, type, or namespace',
    sections,
    summary,
  }
}
