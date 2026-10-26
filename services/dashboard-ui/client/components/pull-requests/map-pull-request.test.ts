import { describe, expect, test } from 'bun:test'
import type { TInstallWorkflow, TPreviewSourcePR } from '@/types'
import {
  previewConfigStatus,
  previewRunStatus,
  pullRequestRuns,
  toPlaygroundPullRequest,
} from './map-pull-request'

const source: TPreviewSourcePR = {
  pr_number: 482,
  title: 'Bump the api component',
  head_sha: '9f2c1ab',
  head_ref: 'dana/node-22',
  url: 'https://github.com/acme/payments/pull/482',
}

const workflow = (
  createdAt: string,
  status: string
): TInstallWorkflow =>
  ({
    id: `wf-${createdAt}`,
    created_at: createdAt,
    finished_at: createdAt,
    status: { status },
    app_branch_runs: [
      {
        id: `run-${createdAt}`,
        pr_number: 482,
        created_at: createdAt,
        started_at: createdAt,
        status,
        metadata: { trigger: 'pull_request', pr_number: 482 },
        vcs_connection_commit: {
          sha: '9f2c1ab',
          message: 'Bump the api component',
          author_name: 'Dana Eng',
        },
        preview: {
          mode: 'plan-only',
          install_name: 'staging',
          resolved_preview_config: {
            mode: 'plan-only',
            install_id: 'inst-staging',
            install_name: 'staging',
          },
        },
      },
    ],
  }) as TInstallWorkflow

describe('map-pull-request', () => {
  test('maps workflow statuses onto preview run statuses', () => {
    expect(previewRunStatus('in-progress')).toBe('running')
    expect(previewRunStatus('success')).toBe('success')
    expect(previewRunStatus('error')).toBe('failed')
    expect(previewRunStatus('cancelled')).toBe('cancelled')
    expect(previewRunStatus(undefined)).toBe('pending')
  })

  test('numbers a pull request’s runs from oldest to newest', () => {
    const runs = pullRequestRuns(
      [workflow('2026-10-04T16:04:00Z', 'success'), workflow('2026-10-02T12:00:00Z', 'error')],
      482
    )
    expect(runs.map((run) => run.run_number)).toEqual([2, 1])
    expect(runs[0]?.status).toBe('success')
    expect(runs[1]?.status).toBe('failed')
  })

  test('treats a branch config without an install as needing input', () => {
    expect(previewConfigStatus({ mode: 'plan-only' })).toBe('needs-input')
    expect(
      previewConfigStatus({ mode: 'plan-only', install_id: 'inst-staging' })
    ).toBe('ready-from-defaults')
    expect(previewConfigStatus({ mode: 'build-only' })).toBe('ready-from-defaults')
  })

  test('builds the page model from the open pull request', () => {
    const runs = pullRequestRuns([workflow('2026-10-04T16:04:00Z', 'success')], 482)
    const pullRequest = toPlaygroundPullRequest({
      source,
      orgId: 'org-1',
      appId: 'app-1',
      branchId: 'br-1',
      repo: 'acme/payments',
      baseRef: 'main',
      runs,
      latest: workflow('2026-10-04T16:04:00Z', 'success').app_branch_runs?.[0],
    })
    expect(pullRequest.repo_name).toBe('payments')
    expect(pullRequest.author_login).toBe('Dana Eng')
    expect(pullRequest.resolved_preview_config.install_name).toBe('staging')
    expect(pullRequest.config_status).toBe('ready-from-defaults')
  })
})
