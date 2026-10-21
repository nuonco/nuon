import { describe, expect, test } from 'bun:test'
import type { TAppBranchConfig } from '@/types'
import { resolveGroupDraft } from './group-draft'

const config = {
  install_groups: [
    { name: 'canary', label_selector: { match_labels: { tier: 'canary' } } },
    { name: 'default', default: true },
  ],
} as unknown as TAppBranchConfig

const stored = (value: object) => JSON.stringify(value)

describe('resolveGroupDraft', () => {
  test('restores the saved group for the same branch', () => {
    const selection = resolveGroupDraft(
      stored({ branchId: 'br-1', mode: 'labels', groupName: 'canary' }),
      'br-1',
      config
    )
    expect(selection?.mode).toBe('labels')
    expect(selection?.group?.name).toBe('canary')
  })

  test('ignores a selection saved on another branch', () => {
    expect(
      resolveGroupDraft(
        stored({ branchId: 'br-2', mode: 'explicit', groupName: 'canary' }),
        'br-1',
        config
      )
    ).toBeNull()
  })

  test('ignores a group that no longer exists', () => {
    expect(
      resolveGroupDraft(
        stored({ branchId: 'br-1', mode: 'explicit', groupName: 'gone' }),
        'br-1',
        config
      )
    ).toBeNull()
  })

  test('ignores missing or malformed drafts', () => {
    expect(resolveGroupDraft(null, 'br-1', config)).toBeNull()
    expect(resolveGroupDraft('{nope', 'br-1', config)).toBeNull()
  })
})
