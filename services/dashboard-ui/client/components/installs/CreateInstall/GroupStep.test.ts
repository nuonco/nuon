import { describe, expect, test } from 'bun:test'
import type { TAppBranchInstallGroup } from '@/types'
import {
  concreteMatchLabels,
  hasLabelConflict,
  mergeGroupLabels,
} from './GroupStep'

const group = (
  matchLabels: Record<string, string>,
  opts: Partial<TAppBranchInstallGroup> = {}
): TAppBranchInstallGroup => ({
  id: 'g1',
  name: 'Test group',
  label_selector: { match_labels: matchLabels },
  ...opts,
})

describe('concreteMatchLabels', () => {
  test('returns labels where value is not a wildcard', () => {
    const result = concreteMatchLabels(
      group({ env: 'production', region: '*', tier: 'web' })
    )
    expect(result).toEqual({ env: 'production', tier: 'web' })
  })

  test('returns empty object for wildcard-only group', () => {
    const result = concreteMatchLabels(group({ env: '*', region: '*' }))
    expect(result).toEqual({})
  })

  test('returns all labels when none are wildcards', () => {
    const result = concreteMatchLabels(
      group({ env: 'staging', region: 'us-east-1' })
    )
    expect(result).toEqual({ env: 'staging', region: 'us-east-1' })
  })

  test('returns empty when group has no label selector', () => {
    const g: TAppBranchInstallGroup = {
      id: 'g1',
      name: 'No selector',
      all_installs: true,
    }
    expect(concreteMatchLabels(g)).toEqual({})
  })
})

describe('hasLabelConflict', () => {
  test('returns true when an install label conflicts with a concrete group label', () => {
    const conflict = hasLabelConflict(
      group({ env: 'production', tier: 'web' }),
      { env: 'staging', tier: 'web' }
    )
    expect(conflict).toBe(true)
  })

  test('returns false when install labels match group labels exactly', () => {
    const conflict = hasLabelConflict(
      group({ env: 'production' }),
      { env: 'production', region: 'us-east-1' }
    )
    expect(conflict).toBe(false)
  })

  test('returns false when install has no overlap with group labels', () => {
    const conflict = hasLabelConflict(
      group({ env: 'production' }),
      { region: 'us-east-1' }
    )
    expect(conflict).toBe(false)
  })

  test('ignores wildcard values when checking conflicts', () => {
    const conflict = hasLabelConflict(
      group({ env: '*', region: 'us-east-1' }),
      { env: 'staging', region: 'eu-west-1' }
    )
    // env=* is wildcard, not included in concrete. region conflict only.
    expect(conflict).toBe(true)
  })

  test('returns false when only wildcards are present', () => {
    const conflict = hasLabelConflict(group({ env: '*' }), {
      env: 'anything',
    })
    expect(conflict).toBe(false)
  })
})

describe('mergeGroupLabels', () => {
  test('merges concrete group labels on top of existing install labels', () => {
    const result = mergeGroupLabels(
      { region: 'us-east-1', team: 'platform' },
      group({ env: 'production', region: '*' })
    )
    expect(result).toEqual({
      region: 'us-east-1',
      team: 'platform',
      env: 'production',
    })
  })

  test('group labels override conflicting install labels', () => {
    const result = mergeGroupLabels(
      { env: 'staging' },
      group({ env: 'production' })
    )
    expect(result).toEqual({ env: 'production' })
  })

  test('returns existing labels unchanged when group is null', () => {
    const existing = { env: 'staging', region: 'us-east-1' }
    const result = mergeGroupLabels(existing, null)
    expect(result).toEqual(existing)
  })
})
