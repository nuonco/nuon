import { describe, expect, test } from 'bun:test'
import {
  countText,
  hasActiveFilters,
  healthSet,
  parseFleetFilters,
  parseFleetView,
  rolloutSegments,
} from './fleet-utils'

describe('parseFleetFilters', () => {
  test('reads known filters from the URL', () => {
    const params = new URLSearchParams(
      'q=%20acme%20&labels=env:prod&health=unhealthy,degraded&runner_status=not-active&rollout=behind'
    )
    expect(parseFleetFilters(params)).toEqual({
      q: 'acme',
      labels: 'env:prod',
      health: 'unhealthy,degraded',
      runner_status: 'not-active',
      rollout: 'behind',
    })
  })

  test('drops values the API would reject', () => {
    const params = new URLSearchParams(
      'runner_status=idle&rollout=sideways&q=%20'
    )
    expect(parseFleetFilters(params)).toEqual({
      q: undefined,
      labels: undefined,
      health: undefined,
      runner_status: undefined,
      rollout: undefined,
    })
  })
})

test('parseFleetView defaults to table', () => {
  expect(parseFleetView(new URLSearchParams())).toBe('table')
  expect(parseFleetView(new URLSearchParams('view=board'))).toBe('board')
  expect(parseFleetView(new URLSearchParams('view=grid'))).toBe('table')
})

test('hasActiveFilters ignores empty values', () => {
  expect(hasActiveFilters({})).toBe(false)
  expect(hasActiveFilters({ q: '' })).toBe(false)
  expect(hasActiveFilters({ rollout: 'applied' })).toBe(true)
})

test('healthSet splits the comma list', () => {
  expect([...healthSet('unhealthy,,degraded')]).toEqual([
    'unhealthy',
    'degraded',
  ])
  expect(healthSet(undefined).size).toBe(0)
})

test('countText shows matches only while filtered', () => {
  const counts = { total: 12, matched: 3 }
  expect(countText(counts, false)).toBe('12 installs')
  expect(countText({ total: 1 }, false)).toBe('1 install')
  expect(countText(counts, true)).toBe('3 of 12 match')
})

test('rolloutSegments skips empty states', () => {
  expect(rolloutSegments({ applied: 2, behind: 0, unknown: 1 })).toEqual([
    { state: 'applied', count: 2 },
    { state: 'unknown', count: 1 },
  ])
})
