import { describe, expect, test } from 'bun:test'
import type { TInstallWorkflow } from '@/types'
import { getRunTitle } from './run-title'

describe('getRunTitle', () => {
  test('uses the commit message for manual preview runs', () => {
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

    expect(getRunTitle(workflow)).toBe('Change app config')
  })

  test('uses run.name for manual preview runs without a commit message', () => {
    const workflow = {
      type: 'app_branches_manual_update',
      name: 'Run',
      app_branch_runs: [
        {
          event_type: 'manual',
          preview: { mode: 'plan-only' },
        },
      ],
    } as TInstallWorkflow

    expect(getRunTitle(workflow)).toBe('Run')
  })

  test('normalizes the legacy manual workflow name', () => {
    const workflow = {
      name: 'Manual run',
      app_branch_runs: [
        { event_type: 'manual', preview: { mode: 'plan-only' } },
      ],
    } as TInstallWorkflow

    expect(getRunTitle(workflow)).toBe('Run')
  })

  test('falls back to type label when run.name is also absent', () => {
    const workflow = {
      type: 'app_branches_manual_update',
      app_branch_runs: [
        {
          event_type: 'manual',
          preview: { mode: 'plan-only' },
        },
      ],
    } as TInstallWorkflow

    expect(getRunTitle(workflow)).toBe('Manual app config update')
  })

  test('keeps the manual app config title for non-preview runs', () => {
    const workflow = {
      type: 'app_branches_manual_update',
      app_branch_runs: [{}],
    } as TInstallWorkflow

    expect(getRunTitle(workflow)).toBe('Manual app config update')
  })

  test('uses the PR number for automated previews', () => {
    const workflow = {
      type: 'app_branches_manual_update',
      app_branch_runs: [
        {
          event_type: 'pull_request',
          pr_number: 23,
          preview: { mode: 'plan-only', source: 'pr' },
          vcs_connection_commit: { message: 'Update payment service' },
        },
      ],
    } as TInstallWorkflow

    expect(getRunTitle(workflow)).toBe('PR #23')
  })
})
