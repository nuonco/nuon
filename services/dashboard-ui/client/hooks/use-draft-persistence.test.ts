import { renderHook, act } from '@testing-library/react'
import { describe, expect, test, beforeEach, afterEach } from 'bun:test'
import { useDraftPersistence, DRAFT_VERSION } from './use-draft-persistence'

describe('useDraftPersistence', () => {
  beforeEach(() => localStorage.clear())
  afterEach(() => localStorage.clear())

  test('initializes with no draft', () => {
    const { result } = renderHook(() =>
      useDraftPersistence({ storageKey: 'k', values: { name: '' } })
    )
    expect(result.current.hasDraft).toBe(false)
    expect(result.current.draftTimestamp).toBeNull()
    expect(result.current.draftValues).toBeNull()
  })

  test('loads a matching-version draft', () => {
    const draft = {
      values: { name: 'saved' },
      timestamp: new Date().toISOString(),
      version: DRAFT_VERSION,
    }
    localStorage.setItem('k', JSON.stringify(draft))

    const { result } = renderHook(() =>
      useDraftPersistence({ storageKey: 'k', values: { name: '' } })
    )
    expect(result.current.hasDraft).toBe(true)
    expect(result.current.draftValues).toEqual({ name: 'saved' })
  })

  test('ignores + removes a stale-version draft', () => {
    localStorage.setItem(
      'k',
      JSON.stringify({
        values: { name: 'x' },
        timestamp: new Date().toISOString(),
        version: 1,
      })
    )
    const { result } = renderHook(() =>
      useDraftPersistence({ storageKey: 'k', values: { name: '' } })
    )
    expect(result.current.hasDraft).toBe(false)
    expect(localStorage.getItem('k')).toBeNull()
  })

  test('expires a draft past its TTL', () => {
    const old = new Date(Date.now() - 25 * 60 * 60 * 1000).toISOString()
    localStorage.setItem(
      'k',
      JSON.stringify({
        values: { name: 'x' },
        timestamp: old,
        version: DRAFT_VERSION,
      })
    )
    const { result } = renderHook(() =>
      useDraftPersistence({
        storageKey: 'k',
        values: { name: '' },
        ttlHours: 24,
      })
    )
    expect(result.current.hasDraft).toBe(false)
  })

  test('invalidates a draft when configId differs', () => {
    localStorage.setItem(
      'k',
      JSON.stringify({
        values: { name: 'x' },
        timestamp: new Date().toISOString(),
        version: DRAFT_VERSION,
        configId: 'a',
      })
    )
    const { result } = renderHook(() =>
      useDraftPersistence({
        storageKey: 'k',
        values: { name: '' },
        configId: 'b',
      })
    )
    expect(result.current.hasDraft).toBe(false)
    expect(localStorage.getItem('k')).toBeNull()
  })

  test('does not persist untouched seed values on mount', async () => {
    renderHook(() =>
      useDraftPersistence({ storageKey: 'k', values: { name: 'seed' } })
    )
    await new Promise((r) => setTimeout(r, 600))
    expect(localStorage.getItem('k')).toBeNull()
  })

  test('persists after values change', async () => {
    const { rerender } = renderHook(
      ({ values }) => useDraftPersistence({ storageKey: 'k', values }),
      { initialProps: { values: { name: 'seed' } } }
    )
    rerender({ values: { name: 'changed' } })
    await new Promise((r) => setTimeout(r, 600))
    const stored = JSON.parse(localStorage.getItem('k') as string)
    expect(stored.values).toEqual({ name: 'changed' })
    expect(stored.version).toBe(DRAFT_VERSION)
  })

  test('clearDraft removes storage and resets state', () => {
    localStorage.setItem(
      'k',
      JSON.stringify({
        values: { name: 'x' },
        timestamp: new Date().toISOString(),
        version: DRAFT_VERSION,
      })
    )
    const { result } = renderHook(() =>
      useDraftPersistence({ storageKey: 'k', values: { name: '' } })
    )
    expect(result.current.hasDraft).toBe(true)
    act(() => result.current.clearDraft())
    expect(result.current.hasDraft).toBe(false)
    expect(result.current.draftValues).toBeNull()
    expect(localStorage.getItem('k')).toBeNull()
  })
})
