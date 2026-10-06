import { describe, expect, test } from 'bun:test'
import { externalGitUrl, gitReferenceUrl } from './vcs-urls'

const sha = 'e5aef07c91b24d0a8f3310c0e5aef07c91b24d0a'

describe('gitReferenceUrl', () => {
  test('links a GitHub slug commit and pull request', () => {
    expect(gitReferenceUrl('acme/platform', { type: 'commit', sha })).toBe(
      `https://github.com/acme/platform/commit/${sha}`
    )
    expect(
      gitReferenceUrl('acme/platform', { type: 'pull-request', number: 104 })
    ).toBe('https://github.com/acme/platform/pull/104')
  })

  test('accepts a full repository URL and strips .git', () => {
    expect(
      gitReferenceUrl('https://github.com/acme/platform.git', {
        type: 'commit',
        sha,
      })
    ).toBe(`https://github.com/acme/platform/commit/${sha}`)
  })

  test('uses host-specific paths', () => {
    expect(
      gitReferenceUrl('https://gitlab.com/acme/platform', {
        type: 'pull-request',
        number: 12,
      })
    ).toBe('https://gitlab.com/acme/platform/-/merge_requests/12')
    expect(
      gitReferenceUrl('https://bitbucket.org/acme/platform', {
        type: 'tag',
        tag: 'v1.2.0',
      })
    ).toBe('https://bitbucket.org/acme/platform/src/v1.2.0')
  })

  test('returns undefined without a repository or reference', () => {
    expect(gitReferenceUrl(undefined, { type: 'commit', sha })).toBeUndefined()
    expect(gitReferenceUrl('acme/platform', { type: 'commit', sha: '' })).toBeUndefined()
    expect(
      gitReferenceUrl('acme/platform', { type: 'pull-request', number: 0 })
    ).toBeUndefined()
  })
})

describe('externalGitUrl', () => {
  test('adds a scheme to a bare host URL', () => {
    expect(externalGitUrl('github.com/acme/platform/pull/104')).toBe(
      'https://github.com/acme/platform/pull/104'
    )
  })

  test('keeps an absolute URL', () => {
    expect(externalGitUrl('https://github.com/acme/platform/commit/abc')).toBe(
      'https://github.com/acme/platform/commit/abc'
    )
  })
})
