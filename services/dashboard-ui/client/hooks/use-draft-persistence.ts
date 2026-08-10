import { useCallback, useEffect, useRef, useState } from 'react'

interface IDraft<T> {
  values: T
  timestamp: string
  version: number
  configId?: string
}

export interface IUseDraftPersistence<T> {
  storageKey: string
  values: T
  enabled?: boolean
  ttlHours?: number
  configId?: string
}

export interface IUseDraftPersistenceReturn<T> {
  hasDraft: boolean
  draftTimestamp: string | null
  draftValues: T | null
  clearDraft: () => void
}

export const DRAFT_VERSION = 2
const DEFAULT_TTL_HOURS = 24
const SAVE_DEBOUNCE_MS = 500

const isLocalStorageAvailable = (): boolean => {
  if (typeof window === 'undefined') return false
  try {
    const test = '__localStorage_test__'
    localStorage.setItem(test, test)
    localStorage.removeItem(test)
    return true
  } catch {
    return false
  }
}

const loadDraft = <T>(
  storageKey: string,
  configId: string | undefined,
  ttlHours: number
): IDraft<T> | null => {
  if (!isLocalStorageAvailable()) return null

  try {
    const stored = localStorage.getItem(storageKey)
    if (!stored) return null

    const draft = JSON.parse(stored) as IDraft<T>

    if (draft.version !== DRAFT_VERSION) {
      localStorage.removeItem(storageKey)
      return null
    }

    if (configId && draft.configId !== configId) {
      localStorage.removeItem(storageKey)
      return null
    }

    const age = Date.now() - new Date(draft.timestamp).getTime()
    if (age > ttlHours * 60 * 60 * 1000) {
      localStorage.removeItem(storageKey)
      return null
    }

    return draft
  } catch (error) {
    console.warn('Failed to load form draft:', error)
    return null
  }
}

export function useDraftPersistence<T extends object>({
  storageKey,
  values,
  enabled = true,
  ttlHours = DEFAULT_TTL_HOURS,
  configId,
}: IUseDraftPersistence<T>): IUseDraftPersistenceReturn<T> {
  const [draft] = useState<IDraft<T> | null>(() =>
    enabled ? loadDraft<T>(storageKey, configId, ttlHours) : null
  )
  const [cleared, setCleared] = useState(false)
  const saveTimeoutRef = useRef<ReturnType<typeof setTimeout>>()

  const serialized = JSON.stringify(values)
  const initialSerializedRef = useRef(serialized)

  useEffect(() => {
    if (!enabled || !isLocalStorageAvailable()) return
    if (serialized === initialSerializedRef.current) return

    if (saveTimeoutRef.current) clearTimeout(saveTimeoutRef.current)
    saveTimeoutRef.current = setTimeout(() => {
      try {
        const next: IDraft<T> = {
          values: JSON.parse(serialized) as T,
          timestamp: new Date().toISOString(),
          version: DRAFT_VERSION,
          ...(configId && { configId }),
        }
        localStorage.setItem(storageKey, JSON.stringify(next))
      } catch (error) {
        if (error instanceof Error && error.name === 'QuotaExceededError') {
          console.warn('localStorage quota exceeded, draft not saved')
        } else {
          console.warn('Failed to save form draft:', error)
        }
      }
    }, SAVE_DEBOUNCE_MS)

    return () => {
      if (saveTimeoutRef.current) clearTimeout(saveTimeoutRef.current)
    }
  }, [enabled, storageKey, configId, serialized])

  const clearDraft = useCallback(() => {
    if (saveTimeoutRef.current) clearTimeout(saveTimeoutRef.current)
    setCleared(true)
    if (!isLocalStorageAvailable()) return
    try {
      localStorage.removeItem(storageKey)
    } catch (error) {
      console.warn('Failed to clear form draft:', error)
    }
  }, [storageKey])

  return {
    hasDraft: !cleared && !!draft,
    draftTimestamp: cleared ? null : (draft?.timestamp ?? null),
    draftValues: cleared ? null : (draft?.values ?? null),
    clearDraft,
  }
}
