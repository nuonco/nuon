import { DateTime, Duration } from 'luxon'
import type { TSpan } from '@/types'

export type TSpanNode = {
  span: TSpan
  children: TSpanNode[]
  hasErrorDescendant: boolean
  depth: number
}

const isError = (span: TSpan) =>
  (span.status_code ?? '').toLowerCase() === 'error'

export const buildSpanForest = (spans: TSpan[]): TSpanNode[] => {
  if (!spans?.length) return []

  const nodes = new Map<string, TSpanNode>()
  for (const span of spans) {
    nodes.set(span.span_id, {
      span,
      children: [],
      hasErrorDescendant: false,
      depth: 0,
    })
  }

  const roots: TSpanNode[] = []
  for (const node of nodes.values()) {
    const parentId = node.span.parent_span_id
    const parent = parentId ? nodes.get(parentId) : undefined
    if (parent) {
      parent.children.push(node)
    } else {
      roots.push(node)
    }
  }

  const sortByStart = (a: TSpanNode, b: TSpanNode) =>
    DateTime.fromISO(a.span.start_time).toMillis() -
    DateTime.fromISO(b.span.start_time).toMillis()

  const visit = (node: TSpanNode, depth: number) => {
    node.depth = depth
    node.children.sort(sortByStart)
    let hasErr = isError(node.span)
    for (const child of node.children) {
      visit(child, depth + 1)
      if (child.hasErrorDescendant) hasErr = true
    }
    node.hasErrorDescendant = hasErr
  }

  roots.sort(sortByStart)
  for (const root of roots) visit(root, 0)
  return roots
}

const buildChildIndex = (spans: TSpan[]): Map<string, string[]> => {
  const out = new Map<string, string[]>()
  for (const s of spans) {
    const parent = s.parent_span_id
    if (!parent) continue
    const existing = out.get(parent)
    if (existing) existing.push(s.span_id)
    else out.set(parent, [s.span_id])
  }
  return out
}

export const collectDescendantIds = (
  spans: TSpan[],
  spanId: string
): Set<string> => {
  const out = new Set<string>([spanId])
  if (!spans?.length || !spanId) return out
  const childIndex = buildChildIndex(spans)
  const queue = [spanId]
  while (queue.length > 0) {
    const id = queue.shift()!
    const children = childIndex.get(id)
    if (!children) continue
    for (const childId of children) {
      if (out.has(childId)) continue
      out.add(childId)
      queue.push(childId)
    }
  }
  return out
}

export const flattenForest = (forest: TSpanNode[]): TSpanNode[] => {
  const out: TSpanNode[] = []
  const walk = (node: TSpanNode) => {
    out.push(node)
    for (const child of node.children) walk(child)
  }
  for (const root of forest) walk(root)
  return out
}

export const traceStart = (spans: TSpan[]): number => {
  if (!spans?.length) return 0
  return spans.reduce((min, s) => {
    const t = DateTime.fromISO(s.start_time).toMillis()
    return t < min ? t : min
  }, Number.POSITIVE_INFINITY)
}

export const traceEnd = (spans: TSpan[]): number => {
  if (!spans?.length) return 0
  return spans.reduce((max, s) => {
    const t = DateTime.fromISO(s.end_time).toMillis()
    return t > max ? t : max
  }, Number.NEGATIVE_INFINITY)
}

// Runner-internal spans are the scaffolding the runner emits around every
// job: the per-step `step.<name>` spans (fetching, planning, applying, …)
// and any other span whose nuon.tool == "runner" that isn't itself a job
// root. End users almost always want to see the high-level job span plus
// the tool operations underneath it (terraform.plan, helm.install,
// git.clone, …) — not the intermediate runner machinery.
//
// `filterRunnerInternal` drops those scaffolding spans and re-parents
// their children to the nearest visible ancestor so the tree still
// reads top-to-bottom: job.deploy → terraform.plan instead of
// job.deploy → step.planning → terraform.plan.
//
// Job spans (nuon.tool == "runner" with nuon.job.type set) are kept —
// they're the user-recognized unit of work and anchor the timeline.
const isRunnerInternal = (span: TSpan): boolean => {
  const attrs = span.attributes
  if (!attrs) return false
  if (attrs['nuon.tool'] !== 'runner') return false
  if (attrs['nuon.job.type']) return false
  return true
}

export const filterRunnerInternal = (spans: TSpan[]): TSpan[] => {
  if (!spans?.length) return spans
  const byId = new Map<string, TSpan>()
  for (const s of spans) byId.set(s.span_id, s)

  const findVisibleAncestor = (id: string | undefined): string | undefined => {
    let cur = id
    while (cur) {
      const span = byId.get(cur)
      if (!span) return undefined
      if (!isRunnerInternal(span)) return cur
      cur = span.parent_span_id
    }
    return undefined
  }

  const out: TSpan[] = []
  for (const s of spans) {
    if (isRunnerInternal(s)) continue
    const newParent = findVisibleAncestor(s.parent_span_id)
    out.push(newParent === s.parent_span_id ? s : { ...s, parent_span_id: newParent })
  }
  return out
}

export const formatDurationNs = (ns: number): string => {
  if (!Number.isFinite(ns) || ns <= 0) return '—'
  const dur = Duration.fromMillis(ns / 1_000_000).shiftTo('minutes', 'seconds', 'milliseconds')
  const { minutes, seconds, milliseconds } = dur.toObject()
  if ((minutes ?? 0) > 0) return `${minutes}m ${Math.round(seconds ?? 0)}s`
  if ((seconds ?? 0) >= 1) return `${((seconds ?? 0) + (milliseconds ?? 0) / 1000).toFixed(2)} s`
  if ((milliseconds ?? 0) >= 1) return `${(milliseconds ?? 0).toFixed(1)} ms`
  return `${Math.round(ns)} ns`
}
