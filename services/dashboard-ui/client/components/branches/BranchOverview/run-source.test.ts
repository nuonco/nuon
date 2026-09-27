import { describe, expect, test } from 'bun:test'
import { resolveRunSource } from './run-source'

const repo = 'acme/platform'

describe('resolveRunSource', () => {
  test('labeled pull request', () => {
    expect(
      resolveRunSource(
        {
          pr_number: 482,
          base_branch: 'main',
          metadata: { trigger: 'github_label', github_label: 'deploy' },
        },
        repo
      )
    ).toEqual({
      kind: 'pull-request',
      number: 482,
      url: 'https://github.com/acme/platform/pull/482',
      label: 'deploy',
      baseBranch: 'main',
    })
  })

  test('pull request from a squash commit message', () => {
    const source = resolveRunSource(
      {
        metadata: { trigger: 'push' },
        vcs_connection_commit: { message: 'Add cache component (#471)' },
      },
      repo
    )
    expect(source).toMatchObject({ kind: 'pull-request', number: 471 })
  })

  test('tag push', () => {
    expect(
      resolveRunSource(
        { metadata: { trigger: 'tag', git_ref: 'refs/tags/v1.4.2' } },
        repo
      )
    ).toEqual({
      kind: 'tag',
      tag: 'v1.4.2',
      url: 'https://github.com/acme/platform/releases/tag/v1.4.2',
    })
  })

  test('plain commit', () => {
    expect(
      resolveRunSource(
        {
          metadata: { trigger: 'push' },
          vcs_connection_commit: { message: 'Bump api' },
        },
        repo
      )
    ).toEqual({ kind: 'commit' })
  })
})
