import { describe, expect, test } from 'vitest'
import { labelsToQueryString, parseLabelsQuery } from './parse'

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
