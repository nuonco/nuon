import type { TDiffNode } from '@/lib/ctl-api/apps/get-app-config-diff'

export const DIFF_SECTION_KEYS: Record<string, string> = {
  components: 'Components',
  actions: 'Actions',
  inputs: 'Install inputs',
  secrets: 'Secrets',
  sandbox: 'Sandbox',
  runner: 'Runner',
  permissions: 'Permissions',
  stack: 'Stack',
}

export type DiffSectionData = {
  name: string
  additions: number
  removals: number
  changed: number
  entries: { op: string; name: string; description: string }[]
}

export function extractSections(node?: TDiffNode): DiffSectionData[] {
  if (!node?.children) return []

  const sections: DiffSectionData[] = []
  for (const child of node.children) {
    const displayName = DIFF_SECTION_KEYS[child.key]
    if (!displayName) continue

    const section: DiffSectionData = { name: displayName, additions: 0, removals: 0, changed: 0, entries: [] }
    collectDiffEntries(child, '', section)
    if (section.entries.length > 0) {
      sections.push(section)
    }
  }
  return sections
}

export function collectDiffEntries(node: TDiffNode, parentKey: string, section: DiffSectionData) {
  if (node.diff && node.diff.op !== 'noop' && node.diff.op !== '') {
    const entry = {
      op: node.diff.op,
      name: parentKey || node.key,
      description: parentKey ? node.diff.diff : node.diff.diff,
    }
    if (node.diff.op === 'add') section.additions++
    else if (node.diff.op === 'remove') section.removals++
    else if (node.diff.op === 'change') section.changed++
    section.entries.push(entry)
    return
  }

  if (node.children) {
    const hasLeaves = node.children.some((c) => c.diff && c.diff.op !== 'noop' && c.diff.op !== '')
    if (hasLeaves) {
      for (const c of node.children) {
        if (c.diff && c.diff.op !== 'noop' && c.diff.op !== '') {
          const entry = { op: c.diff.op, name: node.key, description: c.diff.diff }
          if (c.diff.op === 'add') section.additions++
          else if (c.diff.op === 'remove') section.removals++
          else if (c.diff.op === 'change') section.changed++
          section.entries.push(entry)
        }
      }
    } else {
      for (const c of node.children) {
        collectDiffEntries(c, node.key || parentKey, section)
      }
    }
  }
}

export function computeSummary(sections: DiffSectionData[]) {
  let added = 0, removed = 0, changed = 0
  for (const s of sections) {
    added += s.additions
    removed += s.removals
    changed += s.changed
  }
  return { added, removed, changed }
}
