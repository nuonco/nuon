import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'

/**
 * Merges branch-level comparison sections (source, build, input changes) into
 * per-install tree-diff sections. Entities already in installSections are
 * preserved as-is; only missing sections and missing component entities are
 * added from branchSections.
 */
export function mergeBranchSections(
  installSections: DiffSectionData[],
  branchSections: DiffSectionData[]
): DiffSectionData[] {
  if (!branchSections.length) return installSections

  const result = [...installSections]
  const installKeySet = new Set(result.map((s) => s.sectionKey))

  for (const branchSection of branchSections) {
    if (branchSection.sectionKey === 'components') {
      const idx = result.findIndex((s) => s.sectionKey === 'components')
      if (idx < 0) {
        result.unshift({ ...branchSection })
      } else {
        const existing = new Set(result[idx].entities.map((e) => e.name))
        const extra = branchSection.entities.filter(
          (e) => !existing.has(e.name)
        )
        if (extra.length > 0) {
          result[idx] = {
            ...result[idx],
            entities: [...result[idx].entities, ...extra],
            changed:
              result[idx].changed +
              extra.filter((e) => e.op === 'change').length,
            additions:
              result[idx].additions +
              extra.filter((e) => e.op === 'add').length,
            removals:
              result[idx].removals +
              extra.filter((e) => e.op === 'remove').length,
          }
        }
      }
    } else if (!installKeySet.has(branchSection.sectionKey)) {
      result.push({ ...branchSection })
      installKeySet.add(branchSection.sectionKey)
    }
  }

  return result
}
