import { describe, expect, test } from 'bun:test'
import type { TAppConfig } from '@/types'
import { pickCreateInstallConfig } from './CreateInstallFromAppContainer'

const config = (values: Partial<TAppConfig>): TAppConfig =>
  values as TAppConfig

describe('pickCreateInstallConfig', () => {
  test('selects the newest eligible branch config without falling back', () => {
    const selected = pickCreateInstallConfig(
      [
        config({
          id: 'preview',
          status: 'active',
          labels: { source: 'git-preview-run' },
        }),
        config({ id: 'inactive', status: 'pending' }),
        config({ id: 'active', status: 'active' }),
      ],
      { requireActive: true, fallbackToFirst: false }
    )

    expect(selected?.id).toBe('active')
  })

  test('returns undefined when a branch has no eligible config', () => {
    const selected = pickCreateInstallConfig(
      [
        config({
          id: 'preview',
          status: 'active',
          labels: { source: 'git-preview-run' },
        }),
        config({ id: 'inactive', status: 'pending' }),
      ],
      { requireActive: true, fallbackToFirst: false }
    )

    expect(selected).toBeUndefined()
  })

  test('selects the apps-sync config when branch selection is skipped', () => {
    const selected = pickCreateInstallConfig(
      [
        config({ id: 'branch-config', status: 'active', app_branch_id: 'branch-1' }),
        config({ id: 'apps-sync-config', status: 'active' }),
      ],
      { requireUnbranched: true, fallbackToFirst: false }
    )

    expect(selected?.id).toBe('apps-sync-config')
  })
})
