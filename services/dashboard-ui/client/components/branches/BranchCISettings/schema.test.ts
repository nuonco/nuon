import { describe, expect, test } from 'bun:test'
import {
  branchCISettingsSchema,
  parsePathPatterns,
  pathFilterFromRunConfig,
  runConfigWithPathFilter,
} from './schema'

describe('parsePathPatterns', () => {
  test('trims lines and drops blanks', () => {
    expect(parsePathPatterns(' apps/acme/** \n\napps/acme/docs/**\n')).toEqual([
      'apps/acme/**',
      'apps/acme/docs/**',
    ])
  })
})

describe('pathFilterFromRunConfig', () => {
  test('defaults to empty lists', () => {
    expect(pathFilterFromRunConfig(undefined)).toEqual({
      allowedPaths: '',
      deniedPaths: '',
    })
  })

  test('reads paths and paths_ignore into the two lists', () => {
    expect(
      pathFilterFromRunConfig({
        mode: 'push',
        paths_ignore: ['docs/**', '**/*.md'],
      })
    ).toEqual({
      allowedPaths: '',
      deniedPaths: 'docs/**\n**/*.md',
    })
  })
})

describe('runConfigWithPathFilter', () => {
  test('saves the two lists as paths and paths_ignore', () => {
    expect(
      runConfigWithPathFilter(
        { mode: 'on_github_label', github_label: 'deploy' },
        { allowedPaths: '', deniedPaths: 'docs/**\n**/*.md\n' }
      )
    ).toEqual({
      mode: 'on_github_label',
      github_label: 'deploy',
      paths: [],
      paths_ignore: ['docs/**', '**/*.md'],
    })
  })

  test('empty lists clear both filters', () => {
    expect(
      runConfigWithPathFilter(
        { mode: 'push', paths: ['apps/acme/**'] },
        { allowedPaths: '', deniedPaths: '' }
      )
    ).toEqual({ mode: 'push', paths: [], paths_ignore: [] })
  })
})

const settings = (allowedPaths: string, deniedPaths: string) => ({
  ignoreChangesRegex: '',
  sendStatusesOnIgnore: false,
  allowedPaths,
  deniedPaths,
})

describe('branchCISettingsSchema', () => {
  test('rejects both lists being set', () => {
    const result = branchCISettingsSchema.safeParse(
      settings('apps/acme/**', 'docs/**')
    )
    expect(result.success).toBe(false)
  })

  test('accepts either list on its own', () => {
    expect(
      branchCISettingsSchema.safeParse(settings('', 'docs/**')).success
    ).toBe(true)
    expect(
      branchCISettingsSchema.safeParse(settings('apps/acme/**', '')).success
    ).toBe(true)
  })
})
