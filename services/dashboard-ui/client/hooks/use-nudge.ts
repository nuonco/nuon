import { useCallback, useEffect, useState } from 'react'

const DEFAULT_NUDGE_DURATION_MS = 8000

export function useNudge(
  trigger: boolean,
  durationMs = DEFAULT_NUDGE_DURATION_MS
) {
  const [isOpen, setIsOpen] = useState(false)

  useEffect(() => {
    if (!trigger) {
      setIsOpen(false)
      return
    }
    setIsOpen(true)
    const timer = setTimeout(() => setIsOpen(false), durationMs)
    return () => clearTimeout(timer)
  }, [trigger, durationMs])

  const close = useCallback(() => setIsOpen(false), [])

  return { isOpen, close }
}
