import { describe, expect, test } from 'bun:test'
import type { TAppBranchInstallGroup } from '@/types'
import { installGroupMatch } from './InstallGroupMatch'

const group = (over: Partial<TAppBranchInstallGroup>): TAppBranchInstallGroup =>
  over as TAppBranchInstallGroup

describe('installGroupMatch', () => {
  test('uses the label selector when the group has labels', () => {
    expect(
      installGroupMatch(
        group({ label_selector: { match_labels: { env: 'abc' } } })
      )
    ).toEqual({ kind: 'labels', labels: { env: 'abc' } })
  })

  test('treats a group without labels as the default or pinned', () => {
    expect(installGroupMatch(group({ default: true }))).toEqual({
      kind: 'default',
    })
    expect(installGroupMatch(group({ name: 'Canary' }))).toEqual({
      kind: 'pinned',
    })
  })
})
