import type { TConfigDiffKey, TConfigDiffNode } from '@/types'

export interface IConfigDiffChange {
  path: string
  op: TConfigDiffKey['op']
  diff: string
}

export function collectChangedLeaves(
  node: TConfigDiffNode | null | undefined,
  prefix = ''
): IConfigDiffChange[] {
  if (!node) return []

  const path = prefix ? `${prefix}.${node.key}` : node.key

  if (node.diff) {
    if (node.diff.op === 'noop' || node.diff.op === '') return []
    return [{ path, op: node.diff.op, diff: node.diff.diff }]
  }

  if (node.children) {
    return node.children
      .filter(Boolean)
      .flatMap((child) => collectChangedLeaves(child, path))
  }

  return []
}

export function getDiffSummary(node: TConfigDiffNode): {
  added: number
  changed: number
  removed: number
} {
  const leaves = collectChangedLeaves(node)
  return {
    added: leaves.filter((l) => l.op === 'add').length,
    changed: leaves.filter((l) => l.op === 'change').length,
    removed: leaves.filter((l) => l.op === 'remove').length,
  }
}
