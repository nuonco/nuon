import { describe, expect, test } from 'bun:test'
import type { TBuild } from '@/types'
import {
  buildMatchesInstallBranch,
  buildMatchesInstallConfig,
  excludePreviewBuilds,
} from './BuildSelect'

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

describe('buildMatchesInstallBranch', () => {
  test('matches the build app branch id', () => {
    expect(
      buildMatchesInstallBranch({ app_branch_id: 'br-1' } as TBuild, 'br-1')
    ).toBe(true)
  })

  test('matches the nested app branch when the build column is empty', () => {
    expect(
      buildMatchesInstallBranch(
        { app_branch_run: { app_branch: { id: 'br-1' } } } as TBuild,
        'br-1'
      )
    ).toBe(true)
  })

  test('rejects a build from another app branch', () => {
    expect(
      buildMatchesInstallBranch({ app_branch_id: 'br-other' } as TBuild, 'br-1')
    ).toBe(false)
  })

  test('stays quiet when either branch id is missing', () => {
    expect(buildMatchesInstallBranch({} as TBuild, 'br-1')).toBe(true)
    expect(
      buildMatchesInstallBranch({ app_branch_id: 'br-1' } as TBuild, undefined)
    ).toBe(true)
  })
})
