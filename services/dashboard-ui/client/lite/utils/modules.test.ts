import { describe, expect, test } from 'bun:test'
import {
  ALL_MODULES,
  MODULE_IDS,
  MODULE_PRESETS,
  SETTINGS_MODULE_IDS,
  deploymentConfig,
  enabledModules,
  featuresPatch,
  matchPreset,
  moduleFeature,
  moduleHref,
  moduleById,
  presetModules,
  withModule,
} from './modules'

describe('module resolution', () => {
  test('every module ships when the org carries no module flags', () => {
    expect([...enabledModules(undefined)]).toEqual([...MODULE_IDS])
    expect([...enabledModules({ 'app-branches': true })]).toEqual([
      ...MODULE_IDS,
    ])
  })

  test('a true disable flag hides exactly that module', () => {
    const enabled = enabledModules({
      [moduleFeature('apps')]: true,
      [moduleFeature('installs')]: false,
    })

    expect(enabled.has('apps')).toBe(false)
    expect(enabled.has('installs')).toBe(true)
    expect(enabled.size).toBe(MODULE_IDS.length - 1)
  })

  test('the patch carries only the modules that changed, inverted', () => {
    const current = ALL_MODULES
    const next = withModule(withModule(current, 'apps', false), 'team', false)

    expect(featuresPatch(current, next)).toEqual({
      [moduleFeature('apps')]: true,
      [moduleFeature('team')]: true,
    })
    expect(featuresPatch(next, current)).toEqual({
      [moduleFeature('apps')]: false,
      [moduleFeature('team')]: false,
    })
    expect(featuresPatch(current, current)).toEqual({})
  })
})

describe('presets', () => {
  test('the full set matches the full preset and a trimmed set its preset', () => {
    expect(matchPreset(ALL_MODULES)).toBe('full')
    expect(matchPreset(presetModules('install-operations'))).toBe(
      'install-operations'
    )
    expect(matchPreset(new Set(['apps']))).toBe('custom')
  })

  test('every preset names only registered modules', () => {
    for (const preset of MODULE_PRESETS) {
      for (const id of preset.modules) expect(moduleById(id)).toBeDefined()
    }
  })
})

describe('deployment config', () => {
  test('pins the hidden modules through forced_enabled_features', () => {
    const config = deploymentConfig(presetModules('install-operations'))

    expect(config).toContain('FORCED_ENABLED_FEATURES: "')
    expect(config).toContain(moduleFeature('apps'))
    expect(config).toContain(moduleFeature('webhooks'))
    expect(config).not.toContain(moduleFeature('installs'))
    expect(config).not.toContain(moduleFeature('team'))
  })

  test('pins nothing when every module ships', () => {
    expect(deploymentConfig(ALL_MODULES)).toContain(
      'FORCED_ENABLED_FEATURES: ""'
    )
  })
})

describe('hrefs', () => {
  test('settings modules live under settings and the rest at the org root', () => {
    expect(moduleHref('org_a', moduleById('connections')!)).toBe(
      '/org_a/settings'
    )
    expect(moduleHref('org_a', moduleById('webhooks')!)).toBe(
      '/org_a/settings/webhooks'
    )
    expect(moduleHref('org_a', moduleById('team')!)).toBe('/org_a/teams')
    expect(moduleHref('org_a', moduleById('apps')!)).toBe('/org_a/apps')
  })

  test('connections stays the first settings section', () => {
    expect(SETTINGS_MODULE_IDS[0]).toBe('connections')
  })
})
