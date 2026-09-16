import { useEffect, useCallback } from 'react'

export function useEscapeKey(onEscape: () => void, enabled = true) {
  const handleEscape = useCallback(
    (e: KeyboardEvent) => {
      if (!enabled || e.key !== 'Escape') return
      e.stopPropagation()
      onEscape()
    },
    [enabled, onEscape]
  )

  useEffect(() => {
    if (!enabled) return
    document.addEventListener('keydown', handleEscape)
    return () => {
      document.removeEventListener('keydown', handleEscape)
    }
  }, [enabled, handleEscape])
}
