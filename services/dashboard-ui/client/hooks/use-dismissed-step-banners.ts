import { useCallback, useEffect, useState } from 'react'

const STORAGE_KEY = 'nuon:dismissed-step-banners'
const TTL_MS = 30 * 24 * 60 * 60 * 1000

type DismissedStore = Record<string, number>

const loadStore = (): DismissedStore => {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return {}
    const parsed = JSON.parse(raw) as DismissedStore
    const now = Date.now()
    return Object.fromEntries(
      Object.entries(parsed).filter(([, ts]) => now - ts < TTL_MS)
    )
  } catch {
    return {}
  }
}

const persistStore = (store: DismissedStore) => {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(store))
  } catch {
    // localStorage may be unavailable (private mode, quota) — dismissals
    // simply fall back to in-memory state for the session.
  }
}

export const useDismissedStepBanners = () => {
  const [store, setStore] = useState<DismissedStore>(loadStore)

  useEffect(() => {
    persistStore(store)
  }, [])

  const dismiss = useCallback((stepId: string) => {
    setStore((prev) => {
      const next = { ...prev, [stepId]: Date.now() }
      persistStore(next)
      return next
    })
  }, [])

  const isDismissed = useCallback((stepId: string) => stepId in store, [store])

  return { isDismissed, dismiss }
}
