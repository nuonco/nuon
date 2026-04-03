import type { TIconVariant } from '@/components/common/Icon'

export type SpotlightResult = {
  label: string
  subtitle?: string
  tag?: string
  icon: TIconVariant
} & (
  | { path: string; action?: never }
  | { action: () => void; path?: never }
)

export type ParsedQuery = {
  prefix: 'app' | 'install' | 'component' | 'action' | null
  query: string
  command: string | null
}

export const STATIC_PAGES: (SpotlightResult & { feature?: string })[] = [
  { label: 'Dashboard', path: '/', icon: 'House', feature: 'org-dashboard' },
  { label: 'Apps', path: '/apps', icon: 'AppWindow' },
  { label: 'Installs', path: '/installs', icon: 'Cube' },
  { label: 'Team', path: '/team', icon: 'UsersThree' },
  { label: 'Build runner', path: '/runner', icon: 'Hammer' },
]

export const INSTALL_SUB_PAGES = [
  'Components',
  'Actions',
  'Runner',
  'Workflows',
  'Stacks',
]

export const APP_SUB_PAGES = [
  'Components',
  'Actions',
  'Roles',
  'Policies',
  'Installs',
]

export const APP_BRANCH_SUB_PAGES = [
  'Branches',
  'Sandbox',
]

export const FILTER_PREFIXES = ['app:', 'install:', 'component:', 'action:']

export const COMMANDS_BY_PREFIX: Partial<Record<NonNullable<ParsedQuery['prefix']>, string[]>> = {
  app: [
    'build all components',
  ],
  install: [
    'run adhoc action',
    'edit inputs',
    'view current inputs',
    'sync secrets',
    'reprovision install',
    'reprovision sandbox',
    'deploy all components',
    'restart runner',
  ],
}

const PREFIX_MAP: Record<string, ParsedQuery['prefix']> = {
  'app:': 'app',
  'apps:': 'app',
  'install:': 'install',
  'installs:': 'install',
  'component:': 'component',
  'components:': 'component',
  'action:': 'action',
  'actions:': 'action',
}

export function parseQuery(raw: string): ParsedQuery {
  for (const [p, prefix] of Object.entries(PREFIX_MAP)) {
    if (raw.startsWith(p)) {
      const rest = raw.slice(p.length)
      const slashIdx = rest.indexOf('/')
      if (slashIdx >= 0) {
        return { prefix, query: rest.slice(0, slashIdx).trim(), command: rest.slice(slashIdx + 1).trim() }
      }
      return { prefix, query: rest.trim(), command: null }
    }
  }
  return { prefix: null, query: raw.trim(), command: null }
}

export function getAutocompletion(input: string): string | null {
  if (!input) return null
  if (input.includes('/')) {
    const parsed = parseQuery(input)
    if (!parsed.prefix) return null
    const commands = COMMANDS_BY_PREFIX[parsed.prefix]
    if (!commands) return null
    const slashIdx = input.indexOf('/')
    const before = input.slice(0, slashIdx + 1)
    const after = input.slice(slashIdx + 1).toLowerCase()
    if (!after) return null
    const match = commands.find((c) => c.startsWith(after) && c !== after)
    return match ? before + match : null
  }
  if (input.includes(':')) return null
  const lower = input.toLowerCase()
  const match = FILTER_PREFIXES.find((p) => p.startsWith(lower) && p !== lower)
  return match ?? null
}

export function tokenMatch(text: string, query: string): boolean {
  const tokens = query.toLowerCase().split(/\s+/).filter(Boolean)
  const lower = text.toLowerCase()
  return tokens.every((t) => lower.includes(t))
}
