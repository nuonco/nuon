import type { TIconVariant } from '../components/atoms/Icon'
import { settingsHref } from './hrefs'

export type TModuleId =
  | 'apps'
  | 'installs'
  | 'team'
  | 'connections'
  | 'webhooks'
  | 'triggers'
  | 'api-tokens'
  | 'service-accounts'
  | 'oidc-federation'

export type TModuleArea = 'product' | 'access' | 'integrations'

export type TModuleReadiness = 'built' | 'partial' | 'scaffold'

export type TModuleNavGroup = 'primary' | 'secondary'

export interface IModuleNav {
  path: string
  label: string
  shortcut: string
  group: TModuleNavGroup
}

export interface IModuleSettingsSection {
  path: string
  label: string
}

export interface IModule {
  id: TModuleId
  name: string
  description: string
  area: TModuleArea
  readiness: TModuleReadiness
  icon: TIconVariant
  nav?: IModuleNav
  settings?: IModuleSettingsSection
}

export const MODULE_FEATURE_PREFIX = 'disable-module-'

export const moduleFeature = (id: TModuleId) => `${MODULE_FEATURE_PREFIX}${id}`

export const MODULES: readonly IModule[] = [
  {
    id: 'apps',
    name: 'Apps',
    description:
      'Configure apps, connect branches, and review plans before they roll out.',
    area: 'product',
    readiness: 'partial',
    icon: 'AppWindowIcon',
    nav: { path: 'apps', label: 'Apps', shortcut: 'g a', group: 'primary' },
  },
  {
    id: 'installs',
    name: 'Installs',
    description:
      'Create installs in customer cloud accounts and follow what changed.',
    area: 'product',
    readiness: 'partial',
    icon: 'CubeIcon',
    nav: {
      path: 'installs',
      label: 'Installs',
      shortcut: 'g i',
      group: 'primary',
    },
  },
  {
    id: 'team',
    name: 'Team',
    description: 'Invite members and manage their roles in the org.',
    area: 'access',
    readiness: 'built',
    icon: 'UsersThreeIcon',
    nav: { path: 'teams', label: 'Team', shortcut: 'g t', group: 'secondary' },
  },
  {
    id: 'connections',
    name: 'Connections',
    description: 'Connect the GitHub accounts that app configs are read from.',
    area: 'integrations',
    readiness: 'scaffold',
    icon: 'PlugsConnectedIcon',
    settings: { path: '', label: 'Connections' },
  },
  {
    id: 'webhooks',
    name: 'Webhooks',
    description: 'Send workflow lifecycle events to external endpoints.',
    area: 'integrations',
    readiness: 'scaffold',
    icon: 'WebhooksLogoIcon',
    settings: { path: 'webhooks', label: 'Webhooks' },
  },
  {
    id: 'triggers',
    name: 'Triggers',
    description: 'Run actions when install and workflow events happen.',
    area: 'integrations',
    readiness: 'scaffold',
    icon: 'LightningIcon',
    settings: { path: 'triggers', label: 'Triggers' },
  },
  {
    id: 'api-tokens',
    name: 'API tokens',
    description: 'Issue and revoke API tokens for the CLI and integrations.',
    area: 'access',
    readiness: 'scaffold',
    icon: 'KeyIcon',
    settings: { path: 'api-tokens', label: 'API tokens' },
  },
  {
    id: 'service-accounts',
    name: 'Service accounts',
    description: 'Manage non-human accounts that automate against the API.',
    area: 'access',
    readiness: 'scaffold',
    icon: 'RobotIcon',
    settings: { path: 'service-accounts', label: 'Service accounts' },
  },
  {
    id: 'oidc-federation',
    name: 'OIDC federation',
    description:
      'Trust install stack identities through OIDC federation policies.',
    area: 'access',
    readiness: 'scaffold',
    icon: 'ShieldCheckIcon',
    settings: { path: 'oidc', label: 'OIDC federation' },
  },
]

export const MODULE_IDS: readonly TModuleId[] = MODULES.map(
  (module) => module.id
)

export const ALL_MODULES: ReadonlySet<TModuleId> = new Set(MODULE_IDS)

export const SETTINGS_MODULE_IDS: readonly TModuleId[] = MODULES.flatMap(
  (module) => (module.settings ? [module.id] : [])
)

export const MODULE_AREAS: readonly TModuleArea[] = [
  'product',
  'access',
  'integrations',
]

export const MODULE_AREA_LABELS: Record<TModuleArea, string> = {
  product: 'Product',
  access: 'Access',
  integrations: 'Integrations',
}

export const MODULE_READINESS_LABELS: Record<TModuleReadiness, string> = {
  built: 'Built',
  partial: 'Partly built',
  scaffold: 'Scaffold',
}

export const moduleById = (id: TModuleId) =>
  MODULES.find((module) => module.id === id)

export const moduleHref = (orgId: string, module: IModule) => {
  if (module.settings) return settingsHref(orgId, module.settings.path)
  return `/${orgId}/${module.nav?.path ?? module.id}`
}

export type TOrgFeatures = Readonly<Record<string, boolean>> | undefined

export const isModuleEnabled = (features: TOrgFeatures, id: TModuleId) =>
  !features?.[moduleFeature(id)]

export const enabledModules = (
  features: TOrgFeatures
): ReadonlySet<TModuleId> =>
  new Set(MODULE_IDS.filter((id) => isModuleEnabled(features, id)))

export const withModule = (
  current: ReadonlySet<TModuleId>,
  id: TModuleId,
  enabled: boolean
): ReadonlySet<TModuleId> => {
  const next = new Set(current)
  if (enabled) next.add(id)
  else next.delete(id)
  return next
}

export const featuresPatch = (
  current: ReadonlySet<TModuleId>,
  next: ReadonlySet<TModuleId>
): Record<string, boolean> => {
  const patch: Record<string, boolean> = {}
  for (const id of MODULE_IDS) {
    if (current.has(id) !== next.has(id)) {
      patch[moduleFeature(id)] = !next.has(id)
    }
  }
  return patch
}

export type TModulePresetId =
  | 'full'
  | 'core'
  | 'install-operations'
  | 'app-delivery'

export interface IModulePreset {
  id: TModulePresetId
  name: string
  description: string
  modules: readonly TModuleId[]
}

export const MODULE_PRESETS: readonly IModulePreset[] = [
  {
    id: 'full',
    name: 'Full platform',
    description: 'Every module Lite ships.',
    modules: MODULE_IDS,
  },
  {
    id: 'core',
    name: 'Core',
    description:
      'Apps, installs, and team. Nuon manages settings on the org behalf.',
    modules: ['apps', 'installs', 'team'],
  },
  {
    id: 'install-operations',
    name: 'Install operations',
    description:
      'An operator console for running customer installs. App config ships from git or the CLI.',
    modules: ['installs', 'team'],
  },
  {
    id: 'app-delivery',
    name: 'App delivery',
    description:
      'Ship app config through branches. Installs are created outside this dashboard.',
    modules: ['apps', 'connections', 'webhooks', 'triggers', 'team'],
  },
]

export const presetById = (id: TModulePresetId) =>
  MODULE_PRESETS.find((preset) => preset.id === id)

export const presetModules = (id: TModulePresetId): ReadonlySet<TModuleId> =>
  new Set(presetById(id)?.modules ?? [])

export const matchPreset = (
  enabled: ReadonlySet<TModuleId>
): TModulePresetId | 'custom' =>
  MODULE_PRESETS.find(
    (preset) =>
      preset.modules.length === enabled.size &&
      preset.modules.every((id) => enabled.has(id))
  )?.id ?? 'custom'

export const FORCED_FEATURES_KEY = 'FORCED_ENABLED_FEATURES'

export const deploymentConfig = (enabled: ReadonlySet<TModuleId>) => {
  const pinned = MODULE_IDS.filter((id) => !enabled.has(id)).map(moduleFeature)
  const note = pinned.length
    ? '# ctl-api config. Append these to the deployment forced_enabled_features value.'
    : '# Every module ships. Nothing to pin in forced_enabled_features.'
  return `${note}\n${FORCED_FEATURES_KEY}: "${pinned.join(',')}"`
}
