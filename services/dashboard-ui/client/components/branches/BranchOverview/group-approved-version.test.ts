import { expect, test } from 'bun:test'
import type { TInstallWorkflow } from '@/types'
import { lastApprovedGroupVersions } from './group-approved-version'

const run = ({
  id,
  group,
  response,
  status = 'success',
  sha = 'abcdef1234567890',
  message = 'Ship the cache\n\nLonger body that stays off the row.',
  createdAt = '2026-10-15T12:00:00Z',
  planOnly = false,
  preview = false,
}: {
  id: string
  group: string
  response?: string
  status?: string
  sha?: string
  message?: string
  createdAt?: string
  planOnly?: boolean
  preview?: boolean
}): TInstallWorkflow =>
  ({
    id,
    created_at: createdAt,
    plan_only: planOnly,
    app_branch_runs: [
      {
        app_config_id: `cfg-${id}`,
        created_at: createdAt,
        head_sha: sha,
        plan_only: preview,
        run_type: preview ? 'git-preview-run' : 'manual-run',
        metadata: { trigger: 'manual', head_sha: sha },
        vcs_connection_commit: {
          sha,
          message,
          author_name: 'Avery Chen',
          created_at: createdAt,
        },
      },
    ],
    steps: [
      {
        name: `plan install group: ${group}`,
        status: { status },
        approval: response ? { response: { type: response } } : undefined,
      },
    ],
  }) as TInstallWorkflow

test('uses the newest approved plan for each group', () => {
  const versions = lastApprovedGroupVersions([
    run({
      id: 'new',
      group: 'Canary',
      response: 'approve',
      sha: '1111111111111111',
      createdAt: '2026-10-15T18:00:00Z',
    }),
    run({
      id: 'old',
      group: 'Canary',
      response: 'approve',
      sha: '2222222222222222',
      createdAt: '2026-10-14T18:00:00Z',
    }),
    run({
      id: 'east',
      group: 'US East',
      status: 'pending',
    }),
  ])

  expect(versions.canary?.configId).toBe('cfg-new')
  expect(versions.canary?.sha).toBe('1111111111111111')
  expect(versions.canary?.message).toContain('Ship the cache')
  expect(versions['us east']).toBeUndefined()
})

test('keeps an older apply when the latest plan was auto-skipped', () => {
  const versions = lastApprovedGroupVersions([
    run({ id: 'skipped', group: 'Default', status: 'auto-skipped' }),
    run({
      id: 'applied',
      group: 'Default',
      status: 'success',
      sha: '4444444444444444',
    }),
  ])

  expect(versions.default?.configId).toBe('cfg-applied')
})

test('keeps an older approval when the latest plan is still waiting', () => {
  const versions = lastApprovedGroupVersions([
    run({ id: 'waiting', group: 'Default', status: 'approval-awaiting' }),
    run({
      id: 'approved',
      group: 'Default',
      response: 'approve',
      status: 'approved',
      sha: '3333333333333333',
    }),
  ])

  expect(versions.default?.configId).toBe('cfg-approved')
  expect(versions.default?.sha).toBe('3333333333333333')
})

test('ignores denied plans and preview runs', () => {
  const versions = lastApprovedGroupVersions([
    run({
      id: 'denied',
      group: 'Canary',
      response: 'deny',
      status: 'approval-denied',
    }),
    run({ id: 'preview', group: 'Canary', response: 'approve', preview: true }),
    run({ id: 'plan', group: 'Canary', response: 'approve', planOnly: true }),
  ])

  expect(versions.canary).toBeUndefined()
})
