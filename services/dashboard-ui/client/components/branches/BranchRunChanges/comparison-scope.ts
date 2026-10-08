import type {
  TBranchRunComparisonConfigDiff,
  TSourceArchiveFileDiff,
} from '@/lib'

export type TComparisonScope = {
  sections: string[]
  components?: string[]
}

export const scopedComparisonConfigDiff = (
  config: TBranchRunComparisonConfigDiff | undefined | null,
  scope?: TComparisonScope
): TBranchRunComparisonConfigDiff | undefined | null => {
  if (!config || !scope) return config
  return {
    ...config,
    sections: config.sections
      .filter((section) => scope.sections.includes(section.name))
      .map((section) => {
        const components = scope.components
        if (section.name !== 'Components' || !components) return section
        const entries = section.entries.filter((entry) =>
          components.includes(entry.name)
        )
        return {
          ...section,
          entries,
          additions: entries.filter((entry) => entry.op === 'add').length,
          removals: entries.filter((entry) => entry.op === 'remove').length,
          changed: entries.filter(
            (entry) => !['add', 'remove'].includes(entry.op)
          ).length,
        }
      })
      .filter((section) => section.entries.length > 0),
  }
}

export const scopedComparisonFiles = (
  files: TSourceArchiveFileDiff[] | undefined,
  config: TBranchRunComparisonConfigDiff | undefined | null,
  scope?: TComparisonScope
) => {
  if (!scope || !config) return files
  const selected = scopedComparisonConfigDiff(config, scope)
  if (!selected) return files
  const filePaths = (diff: TBranchRunComparisonConfigDiff) =>
    diff.sections.flatMap((section) =>
      section.entries.flatMap((entry) => (entry.file ? [entry.file] : []))
    )
  const known = new Set(filePaths(config))
  const included = new Set(filePaths(selected))
  // Archive files have no resource ownership; keep shared source dependencies.
  return files?.filter(
    (file) => !known.has(file.path) || included.has(file.path)
  )
}
