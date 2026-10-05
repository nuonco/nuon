import { describe, expect, test } from 'bun:test'
import type { TBuild } from '@/types'
import { buildMatchesInstallConfig, excludePreviewBuilds } from './BuildSelect'

describe('excludePreviewBuilds', () => {
  test('removes preview builds from deploy choices', () => {
    const builds = [
      { id: 'bld-preview', is_preview: true },
      { id: 'bld-main', is_preview: false },
      { id: 'bld-legacy' },
    ] as TBuild[]

    expect(excludePreviewBuilds(builds).map((build) => build.id)).toEqual([
      'bld-main',
      'bld-legacy',
    ])
  })
})

describe('buildMatchesInstallConfig', () => {
  test('matches when the build and install share an app config', () => {
    const build = {
      component_config_connection: { app_config_id: 'cfg-1' },
    } as TBuild

    expect(buildMatchesInstallConfig(build, 'cfg-1')).toBe(true)
  })

  test('does not match when the build was created from another app config', () => {
    const build = {
      component_config_connection: { app_config_id: 'cfg-other' },
    } as TBuild

    expect(buildMatchesInstallConfig(build, 'cfg-1')).toBe(false)
  })

  test('stays quiet when either config id is missing', () => {
    expect(buildMatchesInstallConfig({} as TBuild, 'cfg-1')).toBe(true)
    expect(
      buildMatchesInstallConfig(
        { component_config_connection: { app_config_id: 'cfg-1' } } as TBuild,
        undefined
      )
    ).toBe(true)
  })
})
