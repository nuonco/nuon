import { describe, expect, test } from 'bun:test'
import type { TAppBranchInstallGroup } from '@/types'
import { appBranchGroupFromSelection } from './GroupStep'

describe('appBranchGroupFromSelection', () => {
  test('returns the selected group name for the app branch connection', () => {
    const group = {
      id: 'group-production',
      name: 'production',
      label_selector: { match_labels: { env: 'production' } },
    } as TAppBranchInstallGroup

    expect(appBranchGroupFromSelection({ mode: 'explicit', group })).toBe(
      'production'
    )
  })

  test('returns undefined for the default group', () => {
    const group = {
      id: 'group-default',
      name: 'default',
      default: true,
    } as TAppBranchInstallGroup

    expect(
      appBranchGroupFromSelection({ mode: 'default', group })
    ).toBeUndefined()
  })
})
