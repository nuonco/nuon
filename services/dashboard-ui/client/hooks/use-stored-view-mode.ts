import { useState } from 'react'

export function useStoredViewMode<T extends string>(
  storageKey: string,
  allowed: readonly T[],
  fallback: T,
) {
  const [mode, setMode] = useState<T>(() => {
    try {
      const stored = localStorage.getItem(storageKey)
      if (stored && (allowed as readonly string[]).includes(stored)) {
        return stored as T
      }
    } catch {}
    return fallback
  })

  const updateMode = (value: T) => {
    setMode(value)
    try {
      localStorage.setItem(storageKey, value)
    } catch {}
  }

  return [mode, updateMode] as const
}
