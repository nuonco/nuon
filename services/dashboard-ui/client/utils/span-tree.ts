import type { TSpan } from '@/types'

export type TSpanNode = {
  span: TSpan
  children: TSpanNode[]
  // True iff this span or any descendant has status_code === 'Error'.
  // The runner only sets Error on the failing leaf — the UI bubbles the
  // ✗ marker up the ancestor chain so a failed deep span surfaces at the
  // top of the tree.
  hasErrorDescendant: boolean
  depth: number
}

const isError = (span: TSpan) =>
  (span.status_code ?? '').toLowerCase() === 'error'

// Build a forest from a flat span list. Roots are spans whose
// parent_span_id is missing or refers to a span outside the slice
// (e.g. a parent span ingested separately). Siblings are sorted by
// start_time so the tree reads top-to-bottom in time order.
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
    new Date(a.span.start_time).getTime() -
    new Date(b.span.start_time).getTime()

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

// Build a reverse adjacency map (parent_span_id → child span ids) directly
// from a flat span list, without paying for the full forest construction.
// Used by collectDescendantIds for the span→logs cross-link.
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

// Returns the set { spanId, ...descendants } for a given span.
//
// Used by useLogFilters to expand a single ?span_id=... URL param into the
// full set of span ids whose logs should be shown when the user clicks a
// parent span in the trace tree. Most runner-emitted logs are tagged with
// the parent step span_id (step.execute, step.fetching, …) rather than a
// leaf tool span — without this expansion, clicking step.execute would only
// match logs that happened to come from explicit op.Start scopes, which is
// almost none of them.
//
// If the span is missing from the list (e.g. URL ?span_id=… points at a
// span that hasn't been ingested yet), we return just { spanId } so the
// existing exact-match behavior is preserved.
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

// Flatten a forest into an iteration list (depth-first, in tree order).
// Useful for keyboard navigation and Gantt rendering.
export const flattenForest = (forest: TSpanNode[]): TSpanNode[] => {
  const out: TSpanNode[] = []
  const walk = (node: TSpanNode) => {
    out.push(node)
    for (const child of node.children) walk(child)
  }
  for (const root of forest) walk(root)
  return out
}

// Earliest start across all spans (epoch ms). Used as the timeline origin.
export const traceStart = (spans: TSpan[]): number => {
  if (!spans?.length) return 0
  return spans.reduce((min, s) => {
    const t = new Date(s.start_time).getTime()
    return t < min ? t : min
  }, Number.POSITIVE_INFINITY)
}

// Latest end across all spans (epoch ms). Used for the timeline width.
export const traceEnd = (spans: TSpan[]): number => {
  if (!spans?.length) return 0
  return spans.reduce((max, s) => {
    const t = new Date(s.end_time).getTime()
    return t > max ? t : max
  }, Number.NEGATIVE_INFINITY)
}

export const formatDurationNs = (ns: number): string => {
  if (!Number.isFinite(ns) || ns <= 0) return '—'
  const ms = ns / 1_000_000
  if (ms < 1) return `${ns} ns`
  if (ms < 1000) return `${ms.toFixed(1)} ms`
  const s = ms / 1000
  if (s < 60) return `${s.toFixed(2)} s`
  const m = Math.floor(s / 60)
  const r = s - m * 60
  return `${m}m ${r.toFixed(0)}s`
}
