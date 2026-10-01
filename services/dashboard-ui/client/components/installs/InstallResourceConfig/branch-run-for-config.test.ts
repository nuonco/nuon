import { describe, expect, test } from 'bun:test'
import type { TInstallAppConfigVersion } from '@/types'
import { branchRunForConfig, branchRunHref } from './branch-run-for-config'

const version = (
  overrides: Partial<TInstallAppConfigVersion>
): TInstallAppConfigVersion =>
  ({
    id: 'iacv-1',
    created_at: '2026-09-01T00:00:00Z',
    new_app_config_id: 'cfg-v2',
    app_branch_run: { id: 'run-v2', head_sha: 'abc1234' },
    ...overrides,
  }) as TInstallAppConfigVersion

describe('branchRunForConfig', () => {
  test('returns the run that produced the requested app config', () => {
    const run = branchRunForConfig(
      [
        version({
          id: 'iacv-3',
          created_at: '2026-09-03T00:00:00Z',
          new_app_config_id: 'cfg-v3',
          app_branch_run: { id: 'run-v3' },
        }),
        version({}),
      ],
      'cfg-v2'
    )

    expect(run?.id).toBe('run-v2')
  })

  test('prefers the latest application of the same app config', () => {
    const run = branchRunForConfig(
      [
        version({
          id: 'iacv-old',
          created_at: '2026-08-01T00:00:00Z',
          app_branch_run: { id: 'run-old' },
        }),
        version({
          id: 'iacv-new',
          created_at: '2026-09-02T00:00:00Z',
          app_branch_run: { id: 'run-new' },
        }),
      ],
      'cfg-v2'
    )

    expect(run?.id).toBe('run-new')
  })

  test('returns nothing when that config has no branch run', () => {
    expect(
      branchRunForConfig([version({ app_branch_run: undefined })], 'cfg-v2')
    ).toBeUndefined()
  })
})

describe('branchRunHref', () => {
  test('links to the branch run', () => {
    expect(
      branchRunHref({
        orgId: 'org-1',
        appId: 'app-1',
        run: { id: 'run-1', app_branch: { id: 'brn-1' } },
      })
    ).toBe('/org-1/apps/app-1/branches/brn-1/runs/run-1')
  })
})
