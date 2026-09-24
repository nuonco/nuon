import { describe, expect, test } from 'bun:test'
import type { TInstallWorkflow } from '@/types'
import { getRunTitle, getRunTrigger } from './run-title'

describe('getRunTitle', () => {
  test('uses the preview title for manual preview runs with a commit', () => {
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

    expect(getRunTitle(workflow)).toBe('Preview run')
  })

  test('uses the preview title instead of the generic run name', () => {
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

    expect(getRunTitle(workflow)).toBe('Preview run')
  })

  test('normalizes the legacy manual workflow name', () => {
    const workflow = {
      name: 'Manual run',
      app_branch_runs: [
        { event_type: 'manual', preview: { mode: 'plan-only' } },
      ],
    } as TInstallWorkflow

    expect(getRunTitle(workflow)).toBe('Preview run')
  })

  test('uses the preview title when run.name is absent', () => {
    const workflow = {
      type: 'app_branches_manual_update',
      app_branch_runs: [
        {
          event_type: 'manual',
          preview: { mode: 'plan-only' },
        },
      ],
    } as TInstallWorkflow

    expect(getRunTitle(workflow)).toBe('Preview run')
  })

  test('keeps the manual app config title for non-preview runs', () => {
    const workflow = {
      type: 'app_branches_manual_update',
      app_branch_runs: [{}],
    } as TInstallWorkflow

    expect(getRunTitle(workflow)).toBe('Manual app config update')
  })

  test('uses trigger titles for manual and push runs', () => {
    const manual = {
      app_branch_runs: [{ metadata: { trigger: 'manual' } }],
    } as TInstallWorkflow
    const push = {
      app_branch_runs: [
        {
          metadata: { trigger: 'push' },
          vcs_connection_commit: { message: 'Update app' },
        },
      ],
    } as TInstallWorkflow

    expect(getRunTitle(manual)).toBe('Manual run')
    expect(getRunTitle(push)).toBe('Commit pushed')
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

  test('uses metadata trigger before the legacy event type', () => {
    expect(
      getRunTrigger({
        event_type: 'manual',
        metadata: { trigger: 'push' },
      })
    ).toBe('push')
  })

  test('uses tag metadata for tag run titles', () => {
    const workflow = {
      app_branch_runs: [
        {
          metadata: { trigger: 'tag', tag: 'v2.4.0' },
          vcs_connection_commit: { message: 'Release app' },
        },
      ],
    } as TInstallWorkflow

    expect(getRunTitle(workflow)).toBe('Tag v2.4.0 Pushed')
  })

  test('uses the PR number for label-triggered runs', () => {
    const workflow = {
      app_branch_runs: [
        {
          metadata: {
            trigger: 'github_label',
            pr_number: 142,
            github_label: 'deploy-preview',
          },
        },
      ],
    } as TInstallWorkflow

    expect(getRunTitle(workflow)).toBe('PR #142 Merged')
  })
})
