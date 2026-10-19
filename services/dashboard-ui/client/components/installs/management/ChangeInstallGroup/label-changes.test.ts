import { describe, expect, test } from 'bun:test'
import type { TAppBranchInstallGroup } from '@/types'
import { labelChangesForGroupMove } from './label-changes'

const group = (match: Record<string, string>): TAppBranchInstallGroup =>
  ({
    name: 'preview',
    label_selector: { match_labels: match },
  }) as TAppBranchInstallGroup

describe('labelChangesForGroupMove', () => {
  test('replaces selector labels and leaves other labels out of the change', () => {
    expect(
      labelChangesForGroupMove(
        { env: 'staging', region: 'us-east-1' },
        group({ env: 'preview' })
      )
    ).toEqual([{ key: 'env', from: 'staging', to: 'preview' }])
  })

  test('adds a selector label the install does not have', () => {
    expect(
      labelChangesForGroupMove(
        { region: 'us-east-1' },
        group({ env: 'preview' })
      )
    ).toEqual([{ key: 'env', to: 'preview' }])
  })

  test('skips wildcards and labels that already match', () => {
    expect(
      labelChangesForGroupMove(
        { env: 'preview', channel: 'stable' },
        group({ env: 'preview', tier: '*' })
      )
    ).toEqual([])
  })
})
