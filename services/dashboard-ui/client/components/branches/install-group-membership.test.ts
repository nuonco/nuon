import { describe, expect, test } from 'bun:test'
import type { TInstall } from '@/types'
import { resolveInstallGroupMembership } from './install-group-membership'

const install = (id: string, labels: Record<string, string>): TInstall =>
  ({ id, name: id, labels }) as TInstall

describe('resolveInstallGroupMembership', () => {
  test('uses a default group selector before its fallback', () => {
    const matching = install('matching', { type: 'push' })
    const unmatched = install('unmatched', { type: 'manual' })
    const membership = resolveInstallGroupMembership(
      [matching, unmatched],
      [
        {
          name: 'push',
          default: true,
          label_selector: { match_labels: { type: 'push' } },
        },
      ]
    )

    expect(membership.installsByGroup[0]).toEqual([matching, unmatched])
    expect(membership.unassignedInstalls).toEqual([])
    expect(membership.overlappingInstalls).toEqual([])
  })

  test('detects overlap between a labeled default and another labeled group', () => {
    const matching = install('matching', { type: 'push', env: 'prod' })
    const membership = resolveInstallGroupMembership(
      [matching],
      [
        {
          name: 'push',
          default: true,
          label_selector: { match_labels: { type: 'push' } },
        },
        {
          name: 'prod',
          default: false,
          label_selector: { match_labels: { env: 'prod' } },
        },
      ]
    )

    expect(membership.installsByGroup[0]).toEqual([matching])
    expect(membership.installsByGroup[1]).toEqual([matching])
    expect(membership.overlappingInstalls).toEqual([matching])
  })
})
