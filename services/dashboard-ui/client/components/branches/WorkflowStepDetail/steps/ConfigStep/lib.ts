import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'

export {
  extractSections,
  computeSummary,
} from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
export type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'

// Stack, Install inputs, and Secrets are intentionally excluded from the
// redesigned config-diff presentation.
const EXCLUDED_SECTION_KEYS = new Set(['stack', 'inputs', 'secrets'])

export function filterExcludedSections(
  sections: DiffSectionData[]
): DiffSectionData[] {
  return sections.filter(
    ({ sectionKey }) => !EXCLUDED_SECTION_KEYS.has(sectionKey)
  )
}
