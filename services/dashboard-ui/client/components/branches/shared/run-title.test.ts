import { describe, expect, test } from 'bun:test'
import type { TInstallWorkflow } from '@/types'
import { getRunTitle } from './run-title'

describe('getRunTitle', () => {
  test('uses a stable title for preview runs', () => {
    const workflow = {
      type: 'app_branches_manual_update',
      app_branch_runs: [
        {
          event_type: 'manual',
          preview: { mode: 'plan-only' },
          vcs_connection_commit: { message: 'Change app config' },
        },
      ],
    } as TInstallWorkflow

    expect(getRunTitle(workflow)).toBe('Manual Preview')
  })

  test('keeps the manual app config title for non-preview runs', () => {
    const workflow = {
      type: 'app_branches_manual_update',
      app_branch_runs: [{}],
    } as TInstallWorkflow

    expect(getRunTitle(workflow)).toBe('Manual app config update')
  })

  test('keeps the commit title for automated previews', () => {
    const workflow = {
      type: 'app_branches_manual_update',
      app_branch_runs: [
        {
          event_type: 'pull_request',
          preview: { mode: 'plan-only', source: 'pr' },
          vcs_connection_commit: { message: 'Update payment service' },
        },
      ],
    } as TInstallWorkflow

    expect(getRunTitle(workflow)).toBe('Update payment service')
  })
})
