import { describe, expect, test } from 'bun:test'
import { branchSwitchSectionPath } from './branch-utils'

const base = '/org_1/apps/app_1/branches/brnch_1'

describe('branch-utils', () => {
  describe('branchSwitchSectionPath', () => {
    test('returns empty on the branch overview', () => {
      expect(branchSwitchSectionPath(base, base)).toBe('')
      expect(branchSwitchSectionPath(`${base}/`, base)).toBe('')
    })

    test('preserves single-segment section and tab routes', () => {
      expect(branchSwitchSectionPath(`${base}/runs`, base)).toBe('/runs')
      expect(branchSwitchSectionPath(`${base}/components`, base)).toBe(
        '/components'
      )
      expect(branchSwitchSectionPath(`${base}/installs`, base)).toBe('/installs')
      expect(branchSwitchSectionPath(`${base}/runbooks`, base)).toBe('/runbooks')
      expect(branchSwitchSectionPath(`${base}/inputs`, base)).toBe('/inputs')
      expect(branchSwitchSectionPath(`${base}/configs`, base)).toBe('/configs')
      expect(branchSwitchSectionPath(`${base}/plan`, base)).toBe('/plan')
      expect(branchSwitchSectionPath(`${base}/settings`, base)).toBe('/settings')
    })

    test('falls back to overview for branch-scoped detail routes', () => {
      expect(branchSwitchSectionPath(`${base}/runs/run_1`, base)).toBe('')
      expect(branchSwitchSectionPath(`${base}/components/comp_1`, base)).toBe('')
      expect(
        branchSwitchSectionPath(`${base}/components/comp_1/builds/build_1`, base)
      ).toBe('')
      expect(branchSwitchSectionPath(`${base}/install-configs/sync_1`, base)).toBe(
        ''
      )
    })

    test('does not match a different branch id sharing a prefix', () => {
      expect(
        branchSwitchSectionPath(`${base}x/runs`, base)
      ).toBe('')
    })

    test('returns empty outside the branch base path', () => {
      expect(
        branchSwitchSectionPath('/org_1/apps/app_1/branches', base)
      ).toBe('')
      expect(branchSwitchSectionPath('/org_1/apps/app_1/installs/ins_1', base)).toBe(
        ''
      )
    })
  })
})
