import { describe, expect, test } from 'bun:test'
import type { TPreviewSourcePR } from '@/types'
import { openPullRequestLabel, pullRequestMatches } from './PullRequestList'

const pullRequest: TPreviewSourcePR = {
  pr_number: 128,
  title: 'Retry checkout when the payment provider times out',
  head_ref: 'feat/checkout-retries',
  head_sha: 'a1b2c3d4e5f67890abcd',
  url: 'https://github.com/acme/payments/pull/128',
  author: 'devon',
  labels: [{ name: 'payments', color: '0e8a16' }],
}

const none = new Set<string>()

describe('pullRequestMatches', () => {
  test('matches a title, number, branch, commit, author, or label', () => {
    expect(pullRequestMatches(pullRequest, 'checkout', none, none)).toBe(true)
    expect(pullRequestMatches(pullRequest, '#128', none, none)).toBe(true)
    expect(pullRequestMatches(pullRequest, 'feat/checkout', none, none)).toBe(true)
    expect(pullRequestMatches(pullRequest, 'a1b2c3d', none, none)).toBe(true)
    expect(pullRequestMatches(pullRequest, 'devon', none, none)).toBe(true)
    expect(pullRequestMatches(pullRequest, 'payments', none, none)).toBe(true)
    expect(pullRequestMatches(pullRequest, 'ledger', none, none)).toBe(false)
  })

  test('an empty branch or author selection keeps every pull request', () => {
    expect(pullRequestMatches(pullRequest, '', none, none)).toBe(true)
  })

  test('a branch selection drops other branches', () => {
    expect(
      pullRequestMatches(pullRequest, '', new Set(['feat/receipt-id']), none)
    ).toBe(false)
    expect(
      pullRequestMatches(
        pullRequest,
        'retry',
        new Set(['feat/checkout-retries']),
        none
      )
    ).toBe(true)
  })

  test('an author selection drops other authors', () => {
    expect(pullRequestMatches(pullRequest, '', none, new Set(['rory']))).toBe(
      false
    )
    expect(pullRequestMatches(pullRequest, '', none, new Set(['devon']))).toBe(
      true
    )
  })
})

describe('openPullRequestLabel', () => {
  test('uses the total when nothing is filtered out', () => {
    expect(openPullRequestLabel(7, 7)).toBe('7 open')
  })

  test('shows the visible count against the total', () => {
    expect(openPullRequestLabel(2, 7)).toBe('2 of 7 open')
  })
})
