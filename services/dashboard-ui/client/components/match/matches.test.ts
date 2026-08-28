import { describe, expect, it } from 'bun:test'
import { matchesSelector } from './matches'

describe('matchesSelector', () => {
  it('matches a single label', () => {
    expect(
      matchesSelector({ env: 'prod' }, { match_labels: { env: 'prod' } })
    ).toBe(true)
    expect(
      matchesSelector({ env: 'stage' }, { match_labels: { env: 'prod' } })
    ).toBe(false)
  })

  it('requires the key to exist', () => {
    expect(
      matchesSelector(
        { region: 'us-east-1' },
        { match_labels: { env: 'prod' } }
      )
    ).toBe(false)
  })

  it('ANDs across match_labels', () => {
    const sel = { match_labels: { env: 'prod', tier: 'a' } }
    expect(matchesSelector({ env: 'prod', tier: 'a' }, sel)).toBe(true)
    expect(matchesSelector({ env: 'prod', tier: 'b' }, sel)).toBe(false)
    expect(matchesSelector({ env: 'prod' }, sel)).toBe(false)
  })

  it('treats * as a wildcard value', () => {
    expect(
      matchesSelector({ env: 'prod' }, { match_labels: { env: '*' } })
    ).toBe(true)
    expect(
      matchesSelector({ region: 'us' }, { match_labels: { env: '*' } })
    ).toBe(false)
  })

  it('excludes via not_match_labels', () => {
    expect(
      matchesSelector({ env: 'prod' }, { not_match_labels: { env: 'stage' } })
    ).toBe(true)
    expect(
      matchesSelector({ env: 'stage' }, { not_match_labels: { env: 'stage' } })
    ).toBe(false)
    expect(
      matchesSelector({ env: 'prod' }, { not_match_labels: { env: '*' } })
    ).toBe(false)
  })

  it('matches everything for empty/undefined selector', () => {
    expect(matchesSelector({ env: 'prod' })).toBe(true)
    expect(matchesSelector({ env: 'prod' }, {})).toBe(true)
    expect(matchesSelector(undefined, { match_labels: { env: 'prod' } })).toBe(
      false
    )
  })
})
