import { describe, expect, test } from 'bun:test'
import { composeTitle } from './PageTitle'

describe('composeTitle', () => {
  test('joins segments with a pipe', () => {
    expect(composeTitle(['Components', 'acme-app'])).toBe(
      'Components | acme-app'
    )
  })

  test('omits undefined, null, false, and empty segments', () => {
    expect(composeTitle(['Components', undefined])).toBe('Components')
    expect(composeTitle(['Components', null])).toBe('Components')
    expect(composeTitle(['Components', false])).toBe('Components')
    expect(composeTitle(['Components', ''])).toBe('Components')
  })

  test('collapses to empty string when all segments are unset', () => {
    expect(composeTitle([undefined, null, false, ''])).toBe('')
  })

  test('keeps only the specific segment when the owning entity is still loading', () => {
    const installName: string | undefined = undefined
    expect(composeTitle(['Deploys', installName])).toBe('Deploys')
  })

  test('preserves order (specific first)', () => {
    expect(composeTitle(['prod configs', 'acme-app'])).toBe(
      'prod configs | acme-app'
    )
  })
})
