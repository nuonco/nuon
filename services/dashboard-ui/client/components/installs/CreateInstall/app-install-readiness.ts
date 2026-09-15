import type { TApp, TComponent } from '@/types'

export type AppInstallBadge =
  | 'not-provisionable'
  | 'no-components'
  | 'no-component-builds'

export const APP_INSTALL_BADGE_LABEL: Record<AppInstallBadge, string> = {
  'not-provisionable': 'Not provisionable',
  'no-components': 'No components',
  'no-component-builds': 'No component builds',
}

export const hasRunnerConfig = (app?: TApp) =>
  !!app?.runner_config?.app_runner_type

export const latestComponentIds = (app?: TApp): string[] =>
  app?.app_configs?.[0]?.component_ids ?? []

export const hasActiveComponentBuild = (components?: TComponent[]) =>
  (components ?? []).some(
    (component) => component?.latest_build?.status_v2?.status === 'active'
  )

export const appInstallBadge = (
  app: TApp,
  components?: TComponent[]
): AppInstallBadge | undefined => {
  if (!hasRunnerConfig(app)) return 'not-provisionable'
  if (latestComponentIds(app).length === 0) return 'no-components'
  if (components && !hasActiveComponentBuild(components)) {
    return 'no-component-builds'
  }
  return undefined
}

export const shouldDefaultStackOnly = (
  app: TApp,
  components?: TComponent[]
) => {
  const badge = appInstallBadge(app, components)
  return badge === 'no-components' || badge === 'no-component-builds'
}
