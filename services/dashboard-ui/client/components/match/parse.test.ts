import { describe, expect, test } from 'bun:test'
import { labelsToQueryString, parseLabelsQuery } from './parse'
import { describeMatch } from './types'

describe('parseLabelsQuery', () => {
  test('empty input returns empty map', () => {
    expect(parseLabelsQuery('')).toEqual({})
    expect(parseLabelsQuery('   ')).toEqual({})
  })

  test('single key=value pair', () => {
    expect(parseLabelsQuery('env=prod')).toEqual({ env: 'prod' })
  })

  test('multiple key=value pairs comma-separated', () => {
    expect(parseLabelsQuery('env=prod, tier=critical, owner=alice')).toEqual({
      env: 'prod',
      tier: 'critical',
      owner: 'alice',
    })
  })

  test('wildcard value', () => {
    expect(parseLabelsQuery('owner=*')).toEqual({ owner: '*' })
  })

  test('bare key (no separator) becomes wildcard', () => {
    expect(parseLabelsQuery('env')).toEqual({ env: '*' })
    expect(parseLabelsQuery('env, tier=critical')).toEqual({
      env: '*',
      tier: 'critical',
    })
  })

  test('trims whitespace around parts, keys, and values', () => {
    expect(parseLabelsQuery('  env  =  prod  ,  tier  =  high  ')).toEqual({
      env: 'prod',
      tier: 'high',
    })
  })

  test('empty parts (extra commas) are skipped', () => {
    expect(parseLabelsQuery('env=prod,,, tier=high,,')).toEqual({
      env: 'prod',
      tier: 'high',
    })
  })

  test('empty key after trim is skipped (treated as invalid)', () => {
    // `=value` has an empty key — drop it.
    expect(parseLabelsQuery('=foo, env=prod')).toEqual({ env: 'prod' })
  })

  test('colon separator works (and is preferred over equals)', () => {
    expect(parseLabelsQuery('env:prod')).toEqual({ env: 'prod' })
  })

  test('value may contain equals (split on first separator only)', () => {
    expect(parseLabelsQuery('env=a=b=c')).toEqual({ env: 'a=b=c' })
  })

  test('later duplicate keys overwrite earlier ones', () => {
    expect(parseLabelsQuery('env=stage, env=prod')).toEqual({ env: 'prod' })
  })
})

describe('labelsToQueryString', () => {
  test('empty / undefined input returns empty string', () => {
    expect(labelsToQueryString({})).toBe('')
    expect(labelsToQueryString(undefined)).toBe('')
  })

  test('single key=value', () => {
    expect(labelsToQueryString({ env: 'prod' })).toBe('env=prod')
  })

  test('multiple keys are sorted alphabetically', () => {
    expect(
      labelsToQueryString({ tier: 'critical', env: 'prod', owner: 'alice' })
    ).toBe('env=prod, owner=alice, tier=critical')
  })

  test('round-trip: parse → stringify is stable', () => {
    const raw = 'env=prod, owner=*, tier=critical'
    expect(labelsToQueryString(parseLabelsQuery(raw))).toBe(raw)
  })

  test('wildcard value renders as k=*', () => {
    expect(labelsToQueryString({ env: '*' })).toBe('env=*')
  })
})

describe('describeMatch', () => {
  test('undefined => Org-wide', () => {
    expect(describeMatch(undefined)).toBe('Org-wide')
  })

  test('no populated kind => Org-wide', () => {
    expect(describeMatch({})).toBe('Org-wide')
  })

  test('include labels only', () => {
    expect(
      describeMatch({
        installs: { selector: { match_labels: { env: 'prod', tier: '*' } } },
      })
    ).toBe('Installs: env=prod, tier=*')
  })

  test('exclude labels only renders with "not " prefix', () => {
    expect(
      describeMatch({
        installs: { selector: { not_match_labels: { env: 'stage' } } },
      })
    ).toBe('Installs: not env=stage')
  })

  test('include + exclude joined with semicolon', () => {
    expect(
      describeMatch({
        components: {
          selector: {
            match_labels: { env: 'prod' },
            not_match_labels: { canary: '*' },
          },
        },
      })
    ).toBe('Components: env=prod; not canary=*')
  })

  test('ids fall back to count-noun summary', () => {
    expect(describeMatch({ installs: { ids: ['a', 'b', 'c'] } })).toBe(
      '3 installs'
    )
    expect(describeMatch({ installs: { ids: ['a'] } })).toBe('1 install')
  })

  test('empty TargetMatch{} renders as Any', () => {
    expect(describeMatch({ actions: {} })).toBe('Any actions')
  })

  test('empty selector maps fall through to Any (matches nothing summary)', () => {
    // describeMatch treats empty selectors as no labels; falls through
    // to ids → empty → Any. The server rejects this shape on submit.
    expect(
      describeMatch({
        installs: { selector: { match_labels: {}, not_match_labels: {} } },
      })
    ).toBe('Any installs')
  })
})
