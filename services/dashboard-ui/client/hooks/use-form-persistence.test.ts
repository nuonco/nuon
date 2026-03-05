import { renderHook, act } from '@testing-library/react'
import { describe, expect, test, beforeEach, afterEach, vi } from 'vitest'
import { useFormPersistence } from './use-form-persistence'
import { useRef } from 'react'

describe('useFormPersistence', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
  })

  afterEach(() => {
    localStorage.clear()
  })

  test('should initialize with no draft', () => {
    const { result } = renderHook(() => {
      const formRef = useRef<HTMLFormElement>(null)
      return useFormPersistence({
        storageKey: 'test-draft',
        formRef,
      })
    })

    expect(result.current.hasDraft).toBe(false)
    expect(result.current.draftTimestamp).toBeNull()
    expect(result.current.draftValues).toBeNull()
  })

  test('should load existing draft from localStorage', () => {
    const draftData = {
      values: { name: 'test-install', region: 'us-east-1' },
      timestamp: new Date().toISOString(),
      version: 1,
    }

    localStorage.setItem('test-draft', JSON.stringify(draftData))

    const { result } = renderHook(() => {
      const formRef = useRef<HTMLFormElement>(null)
      return useFormPersistence({
        storageKey: 'test-draft',
        formRef,
      })
    })

    expect(result.current.hasDraft).toBe(true)
    expect(result.current.draftTimestamp).toBe(draftData.timestamp)
    expect(result.current.draftValues).toEqual(draftData.values)
  })

  test('should clear draft', () => {
    const draftData = {
      values: { name: 'test-install' },
      timestamp: new Date().toISOString(),
      version: 1,
    }

    localStorage.setItem('test-draft', JSON.stringify(draftData))

    const { result } = renderHook(() => {
      const formRef = useRef<HTMLFormElement>(null)
      return useFormPersistence({
        storageKey: 'test-draft',
        formRef,
      })
    })

    expect(result.current.hasDraft).toBe(true)

    act(() => {
      result.current.clearDraft()
    })

    expect(result.current.hasDraft).toBe(false)
    expect(result.current.draftTimestamp).toBeNull()
    expect(result.current.draftValues).toBeNull()
    expect(localStorage.getItem('test-draft')).toBeNull()
  })

  test('should expire draft after TTL', () => {
    const oldTimestamp = new Date(Date.now() - 25 * 60 * 60 * 1000).toISOString()
    const draftData = {
      values: { name: 'test-install' },
      timestamp: oldTimestamp,
      version: 1,
    }

    localStorage.setItem('test-draft', JSON.stringify(draftData))

    const { result } = renderHook(() => {
      const formRef = useRef<HTMLFormElement>(null)
      return useFormPersistence({
        storageKey: 'test-draft',
        formRef,
        ttlHours: 24,
      })
    })

    expect(result.current.hasDraft).toBe(false)
    expect(localStorage.getItem('test-draft')).toBeNull()
  })

  test('should clear stale draft when configId changes', () => {
    const draftData = {
      values: { name: 'test-install' },
      timestamp: new Date().toISOString(),
      version: 1,
      configId: 'config-v1',
    }

    localStorage.setItem('test-draft', JSON.stringify(draftData))

    const { result } = renderHook(() => {
      const formRef = useRef<HTMLFormElement>(null)
      return useFormPersistence({
        storageKey: 'test-draft',
        formRef,
        configId: 'config-v2',
      })
    })

    expect(result.current.hasDraft).toBe(false)
    expect(localStorage.getItem('test-draft')).toBeNull()
  })

  test('should handle localStorage unavailable gracefully', () => {
    const getItemSpy = vi
      .spyOn(Storage.prototype, 'getItem')
      .mockImplementation(() => {
        throw new Error('localStorage not available')
      })

    const { result } = renderHook(() => {
      const formRef = useRef<HTMLFormElement>(null)
      return useFormPersistence({
        storageKey: 'test-draft',
        formRef,
      })
    })

    expect(result.current.hasDraft).toBe(false)

    getItemSpy.mockRestore()
  })

  test('should handle invalid JSON in localStorage', () => {
    localStorage.setItem('test-draft', 'invalid-json')

    const { result } = renderHook(() => {
      const formRef = useRef<HTMLFormElement>(null)
      return useFormPersistence({
        storageKey: 'test-draft',
        formRef,
      })
    })

    expect(result.current.hasDraft).toBe(false)
  })

  test('should ignore draft with wrong version', () => {
    const draftData = {
      values: { name: 'test-install' },
      timestamp: new Date().toISOString(),
      version: 999,
    }

    localStorage.setItem('test-draft', JSON.stringify(draftData))

    const { result } = renderHook(() => {
      const formRef = useRef<HTMLFormElement>(null)
      return useFormPersistence({
        storageKey: 'test-draft',
        formRef,
      })
    })

    expect(result.current.hasDraft).toBe(false)
    expect(localStorage.getItem('test-draft')).toBeNull()
  })
})
