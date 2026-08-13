import { useState } from 'react'

export function useStoredRecord<V>(
  storageKey: string,
  fallback: Record<string, V> = {},
) {
  const [record, setRecord] = useState<Record<string, V>>(() => {
    try {
      const stored = localStorage.getItem(storageKey)
      if (stored) return { ...fallback, ...JSON.parse(stored) }
    } catch {}
    return fallback
  })

  const setValue = (key: string, value: V) => {
    setRecord((prev) => {
      const next = { ...prev, [key]: value }
      try {
        localStorage.setItem(storageKey, JSON.stringify(next))
      } catch {}
      return next
    })
  }

  return [record, setValue] as const
}
