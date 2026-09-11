import { describe, expect, test } from 'bun:test'
import { hasLabelSelector, matchesSelector } from './label-selector'

describe('matchesSelector', () => {
  test('matches a single label', () => {
    expect(
      matchesSelector({ env: 'prod' }, { match_labels: { env: 'prod' } })
    ).toBe(true)
    expect(
      matchesSelector({ env: 'stage' }, { match_labels: { env: 'prod' } })
    ).toBe(false)
  })

  test('requires a match_labels key to exist', () => {
    expect(
      matchesSelector(
        { region: 'us-east-1' },
        { match_labels: { env: 'prod' } }
      )
    ).toBe(false)
    expect(
      matchesSelector(undefined, { match_labels: { env: 'prod' } })
    ).toBe(false)
  })

  test('ANDs across match_labels', () => {
    const selector = { match_labels: { env: 'prod', tier: 'a' } }
    expect(matchesSelector({ env: 'prod', tier: 'a' }, selector)).toBe(true)
    expect(matchesSelector({ env: 'prod', tier: 'b' }, selector)).toBe(false)
    expect(matchesSelector({ env: 'prod' }, selector)).toBe(false)
  })

  test('treats * as a wildcard value that still requires the key', () => {
    expect(
      matchesSelector({ env: 'prod' }, { match_labels: { env: '*' } })
    ).toBe(true)
    expect(
      matchesSelector({ region: 'us' }, { match_labels: { env: '*' } })
    ).toBe(false)
  })

  test('excludes via not_match_labels', () => {
    expect(
      matchesSelector({ env: 'prod' }, { not_match_labels: { env: 'stage' } })
    ).toBe(true)
    expect(
      matchesSelector({ env: 'stage' }, { not_match_labels: { env: 'stage' } })
    ).toBe(false)
  })

  test('does not exclude when a not_match_labels key is missing', () => {
    expect(
      matchesSelector(
        { region: 'us-east-1' },
        { not_match_labels: { env: 'stage' } }
      )
    ).toBe(true)
  })

  test('rejects when not_match_labels uses * and the key is present', () => {
    expect(
      matchesSelector({ env: 'prod' }, { not_match_labels: { env: '*' } })
    ).toBe(false)
    expect(
      matchesSelector({ region: 'us' }, { not_match_labels: { env: '*' } })
    ).toBe(true)
  })

  test('ANDs match_labels with not_match_labels', () => {
    const selector = {
      match_labels: { env: '*' },
      not_match_labels: { env: 'stage' },
    }
    expect(matchesSelector({ env: 'prod' }, selector)).toBe(true)
    expect(matchesSelector({ env: 'stage' }, selector)).toBe(false)
  })

  test('matches everything for a nil selector', () => {
    expect(matchesSelector({ env: 'prod' })).toBe(true)
    expect(matchesSelector({ env: 'prod' }, null)).toBe(true)
    expect(matchesSelector({ env: 'prod' }, undefined)).toBe(true)
    expect(matchesSelector(undefined)).toBe(true)
  })

  test('matches everything for an empty selector', () => {
    expect(matchesSelector({ env: 'prod' }, {})).toBe(true)
  })
})

describe('hasLabelSelector', () => {
  test('is false for nil or empty selectors', () => {
    expect(hasLabelSelector()).toBe(false)
    expect(hasLabelSelector(null)).toBe(false)
    expect(hasLabelSelector({})).toBe(false)
    expect(hasLabelSelector({ match_labels: {}, not_match_labels: {} })).toBe(
      false
    )
  })

  test('is true when either polarity has a key', () => {
    expect(hasLabelSelector({ match_labels: { env: 'prod' } })).toBe(true)
    expect(hasLabelSelector({ not_match_labels: { env: 'stage' } })).toBe(true)
  })
})
