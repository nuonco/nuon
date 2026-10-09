import { describe, expect, test } from 'bun:test'
import { commitRangeLabel } from './CommitRange'

describe('commitRangeLabel', () => {
  test('shows a range when the commits differ', () => {
    expect(
      commitRangeLabel({
        sha: 'a1b2c3d4e5f6',
        previousSha: '9f8e7d6c5b4a',
      })
    ).toBe('9f8e7d6 → a1b2c3d')
  })

  test('shows one sha when there is no earlier commit', () => {
    expect(
      commitRangeLabel({
        sha: 'a1b2c3d4e5f6',
        previousSha: 'a1b2c3d4e5f6',
      })
    ).toBe('a1b2c3d')
    expect(commitRangeLabel({ sha: 'a1b2c3d4e5f6' })).toBe('a1b2c3d')
  })
})
