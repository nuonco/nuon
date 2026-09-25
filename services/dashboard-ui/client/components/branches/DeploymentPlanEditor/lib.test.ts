import { describe, expect, test } from 'bun:test'
import { duplicateSelectorGroupIds, selectorKey } from './lib'
import type { IInstallGroup } from './types'

const group = (
  id: string,
  mode: IInstallGroup['selection_mode'],
  labels?: Record<string, string>
): IInstallGroup => ({
  id,
  name: id,
  selection_mode: mode,
  is_default: false,
  order: 0,
  max_parallel: 1,
  auto_approve_on_policies_passing: false,
  label_selector: labels ? { match_labels: labels } : null,
})

describe('selectorKey', () => {
  test('is stable across key order', () => {
    expect(selectorKey({ env: 'prod', tier: 'web' })).toBe(
      selectorKey({ tier: 'web', env: 'prod' })
    )
  })

  test('returns null for empty or missing labels', () => {
    expect(selectorKey(null)).toBeNull()
    expect(selectorKey({})).toBeNull()
    expect(selectorKey({ ' ': 'prod' })).toBeNull()
  })
})

describe('duplicateSelectorGroupIds', () => {
  test('flags groups with the same non-empty selector', () => {
    expect(
      duplicateSelectorGroupIds([
        group('a', 'labels', { env: 'prod', tier: 'web' }),
        group('b', 'labels', { tier: 'web', env: 'prod' }),
      ])
    ).toEqual(new Set(['a', 'b']))
  })

  test('ignores empty and pinned selectors', () => {
    expect(
      duplicateSelectorGroupIds([
        group('a', 'pinned'),
        group('b', 'pinned'),
        group('c', 'labels'),
        group('d', 'labels', { env: 'prod' }),
      ])
    ).toEqual(new Set())
  })

  test('allows distinct selectors', () => {
    expect(
      duplicateSelectorGroupIds([
        group('a', 'labels', { env: 'prod' }),
        group('b', 'labels', { env: 'staging' }),
      ])
    ).toEqual(new Set())
  })
})
