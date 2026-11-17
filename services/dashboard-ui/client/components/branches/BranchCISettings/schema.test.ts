import { describe, expect, test } from 'bun:test'
import {
  branchCISettingsSchema,
  MAX_PATH_PATTERNS,
  parsePathPatterns,
  pathFilterFromRunConfig,
  runConfigWithPathFilter,
} from './schema'

describe('parsePathPatterns', () => {
  test('trims lines and drops blanks', () => {
    expect(parsePathPatterns(' apps/acme/** \n\n!apps/acme/docs/**\n')).toEqual(
      ['apps/acme/**', '!apps/acme/docs/**']
    )
  })
})

describe('pathFilterFromRunConfig', () => {
  test('defaults to empty run paths', () => {
    expect(pathFilterFromRunConfig(undefined)).toEqual({
      pathFilterMode: 'paths',
      pathPatterns: '',
    })
  })

  test('reads paths_ignore', () => {
    expect(
      pathFilterFromRunConfig({
        mode: 'push',
        paths_ignore: ['docs/**', '**/*.md'],
      })
    ).toEqual({
      pathFilterMode: 'paths_ignore',
      pathPatterns: 'docs/**\n**/*.md',
    })
  })
})

describe('runConfigWithPathFilter', () => {
  test('keeps the run mode and clears the other list', () => {
    expect(
      runConfigWithPathFilter(
        {
          mode: 'on_github_label',
          github_label: 'deploy',
          paths_ignore: ['docs/**'],
        },
        { pathFilterMode: 'paths', pathPatterns: 'apps/acme/**\n' }
      )
    ).toEqual({
      mode: 'on_github_label',
      github_label: 'deploy',
      paths: ['apps/acme/**'],
      paths_ignore: [],
    })
  })
})

describe('branchCISettingsSchema', () => {
  test('rejects too many patterns', () => {
    const result = branchCISettingsSchema.safeParse({
      ignoreChangesRegex: '',
      sendStatusesOnIgnore: false,
      pathFilterMode: 'paths',
      pathPatterns: Array.from(
        { length: MAX_PATH_PATTERNS + 1 },
        (_, i) => `p${i}/**`
      ).join('\n'),
    })
    expect(result.success).toBe(false)
  })
})
