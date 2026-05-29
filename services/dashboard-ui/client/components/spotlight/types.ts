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
  prefix: 'app' | 'install' | 'component' | 'action' | 'org' | null
  query: string
  command: string | null
}

export const STATIC_PAGES: (SpotlightResult & { feature?: string })[] = [
  { label: 'Dashboard', path: '/', icon: 'HouseIcon', feature: 'org-dashboard' },
  { label: 'Apps', path: '/apps', icon: 'AppWindowIcon' },
  { label: 'Installs', path: '/installs', icon: 'CubeIcon' },
  { label: 'Team', path: '/team', icon: 'UsersThreeIcon' },
  { label: 'Build runner', path: '/runner', icon: 'HammerIcon' },
  { label: 'Webhooks', path: '/webhooks', icon: 'WebhooksLogoIcon' },
]

export const INSTALL_SUB_PAGES = [
  'Components',
  'Actions',
  'Runner',
  'Workflows',
  'Stacks',
  'Inputs',
  'State',
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

export const FILTER_PREFIXES = ['app:', 'install:', 'component:', 'action:', 'org:']

export const COMMANDS_BY_PREFIX: Partial<Record<NonNullable<ParsedQuery['prefix']>, string[]>> = {
  app: [
    'build all components',
  ],
  action: [
    'run',
  ],
  component: [
    'build',
    'deploy',
    'drift scan',
    'teardown',
  ],
  install: [
    'deploy all components',
    'edit inputs',
    'edit stack overrides',
    'reprovision install',
    'reprovision sandbox',
    'restart runner',
    'run adhoc action',
    'sync secrets',
    'view current inputs',
    'view state',
  ],
}

export const COMMAND_DESCRIPTIONS: Record<string, string> = {
  'build all components': 'Trigger a build for every component in the app',
  'run': 'Manually trigger an action workflow run',
  'build': 'Trigger a build for the component',
  'deploy': 'Deploy the component to the install',
  'drift scan': 'Run a plan-only deploy to detect configuration drift',
  'teardown': 'Tear down the component from the install',
  'deploy all components': 'Deploy every component in the install',
  'edit inputs': 'Update the install input values',
  'edit stack overrides': 'Override stack-level configuration for the install',
  'reprovision install': 'Re-run the install provisioning workflow',
  'reprovision sandbox': 'Re-run the sandbox provisioning workflow',
  'restart runner': 'Restart the runner process for the install',
  'run adhoc action': 'Execute a one-off adhoc action on the install',
  'sync secrets': 'Sync secrets to the install runner',
  'view current inputs': 'View the current input values for the install',
  'view state': 'View the install state object',
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
  'org:': 'org',
  'orgs:': 'org',
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

const GLOBAL_COMMANDS = [
  'run adhoc action',
  'deploy all components',
  'reprovision install',
  'sync secrets',
]

export function getAutocompletion(input: string): string | null {
  if (!input) return null
  if (input.includes('/')) {
    const parsed = parseQuery(input)
    if (!parsed.prefix) {
      if (input.startsWith('/')) {
        const after = input.slice(1).toLowerCase()
        if (!after) return null
        const match = GLOBAL_COMMANDS.find((c) => c.startsWith(after) && c !== after)
        return match ? '/' + match : null
      }
      return null
    }
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
