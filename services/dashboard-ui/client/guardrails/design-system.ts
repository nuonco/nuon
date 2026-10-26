import { readdirSync, readFileSync } from 'node:fs'
import { join, relative } from 'node:path'

const CLIENT_DIR = join(import.meta.dir, '..')
const ROOTS = ['views', 'components']
const SKIP_DIRS = new Set(['onboarding', 'playground', '__stories__', '__fixtures__'])

const PRIMITIVE_DIRS = ['components/common/', 'components/layout/', 'components/surfaces/']
const isPrimitive = (rel: string) => PRIMITIVE_DIRS.some((dir) => rel.startsWith(dir))
const isView = (rel: string) => rel.startsWith('views/')
const isFilterComponent = (rel: string) => /filter/i.test(rel) && !isPrimitive(rel)

export interface IRule {
  id: string
  pattern: RegExp
  scope: (rel: string) => boolean
}

export const RULES: IRule[] = [
  {
    id: 'stock-color',
    pattern:
      /\b(?:bg|text|border|from|to|via|ring|outline|fill|stroke|divide|placeholder|decoration|shadow)-(?:slate|gray|zinc|neutral|stone|amber|yellow|lime|emerald|teal|cyan|sky|indigo|violet|purple|fuchsia|pink|rose)-\d+\b/g,
    scope: (rel) => !isPrimitive(rel),
  },
  {
    id: 'arbitrary-px',
    pattern: /-\[\d+(?:\.\d+)?px\]/g,
    scope: (rel) => !isPrimitive(rel),
  },
  {
    id: 'border-width',
    pattern: /\bborder(?:-[trblxy])?-(?:2|4|8)\b/g,
    scope: (rel) => !isPrimitive(rel),
  },
  {
    id: 'border-color',
    pattern:
      /\bborder(?:-[trblxy])?-(?:cool-grey|dark-grey|primary|red|green|blue|orange|slate|gray|zinc|neutral|stone|amber|yellow|lime|emerald|teal|cyan|sky|indigo|violet|purple|fuchsia|pink|rose)-\d+\b/g,
    scope: (rel) => !isPrimitive(rel),
  },
  {
    id: 'space-utility',
    pattern: /\bspace-[xy]-\d+\b/g,
    scope: (rel) => !isPrimitive(rel),
  },
  {
    id: 'raw-hex',
    pattern: /#[0-9a-fA-F]{3,8}\b/g,
    scope: (rel) => !isPrimitive(rel),
  },
  {
    id: 'skeleton',
    pattern: /<(?:Table|Timeline)?Skeleton\b/g,
    scope: (rel) => !isPrimitive(rel),
  },
  {
    id: 'raw-control',
    pattern: /<(?:button|input|select|textarea)\b/g,
    scope: (rel) => !isPrimitive(rel),
  },
  {
    id: 'hand-rolled-header',
    pattern: /<(?:PageHeader|HeadingGroup|PageHeadingGroup)\b/g,
    scope: (rel) => isView(rel) && !rel.endsWith('Layout.tsx'),
  },
  {
    id: 'size-prop',
    pattern: /\bsize=\{?['"](?:sm|xs)['"]\}?/g,
    scope: (rel) => !isPrimitive(rel),
  },
  {
    id: 'ghost-variant',
    pattern: /\bvariant=\{?['"]ghost['"]\}?/g,
    scope: (rel) => !isPrimitive(rel),
  },
  {
    id: 'select-in-filter',
    pattern: /<(?:Select|FormSelect)\b/g,
    scope: isFilterComponent,
  },
  {
    id: 'aria-pressed',
    pattern: /\baria-pressed\b/g,
    scope: (rel) => !isPrimitive(rel),
  },
  {
    id: 'hand-rolled-content-clip',
    pattern: /mask-image:linear-gradient/g,
    scope: (rel) => !isPrimitive(rel),
  },
]

export type TViolationCounts = Record<string, Record<string, number>>

const walk = (dir: string): string[] =>
  readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    if (entry.isDirectory()) {
      return SKIP_DIRS.has(entry.name) ? [] : walk(join(dir, entry.name))
    }
    if (!entry.isFile() || !/\.tsx?$/.test(entry.name)) return []
    if (/\.(stories|test)\.tsx?$/.test(entry.name)) return []
    return [join(dir, entry.name)]
  })

export const violationsIn = (source: string, rel: string): Record<string, number> => {
  const counts: Record<string, number> = {}
  for (const rule of RULES) {
    if (!rule.scope(rel)) continue
    const matches = source.match(rule.pattern)
    if (matches?.length) counts[rule.id] = matches.length
  }
  return counts
}

export const scan = (clientDir = CLIENT_DIR): TViolationCounts => {
  const result: TViolationCounts = {}
  for (const rule of RULES) result[rule.id] = {}

  for (const root of ROOTS) {
    for (const file of walk(join(clientDir, root))) {
      const rel = relative(clientDir, file)
      const counts = violationsIn(readFileSync(file, 'utf8'), rel)
      for (const [id, count] of Object.entries(counts)) {
        result[id]![rel] = count
      }
    }
  }

  for (const id of Object.keys(result)) {
    result[id] = Object.fromEntries(
      Object.entries(result[id]!).sort(([a], [b]) => a.localeCompare(b)),
    )
  }
  return result
}

export const ratchetFailures = (
  actual: TViolationCounts,
  baseline: TViolationCounts,
): string[] => {
  const failures: string[] = []
  const ids = [...new Set([...Object.keys(actual), ...Object.keys(baseline)])].sort()

  for (const id of ids) {
    const got = actual[id] ?? {}
    const allowed = baseline[id] ?? {}
    const files = [...new Set([...Object.keys(got), ...Object.keys(allowed)])].sort()
    for (const file of files) {
      const count = got[file] ?? 0
      const cap = allowed[file] ?? 0
      if (count > cap) {
        failures.push(
          `${id} ${file}: ${count} (baseline ${cap}). This is a new design-system violation — ` +
            'fix it instead of raising the baseline. See client/components/PATTERNS.md.',
        )
      } else if (count < cap) {
        failures.push(
          `${id} ${file}: ${count} (baseline ${cap}). The file improved — lower its count in ` +
            'design-system.baseline.json so the baseline cannot silently regress.',
        )
      }
    }
  }

  return failures
}
