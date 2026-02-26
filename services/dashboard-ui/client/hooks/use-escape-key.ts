'use client'

import { useEffect, useCallback } from 'react'

export function useEscapeKey(onEscape: () => void) {
  const handleEscape = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onEscape()
      }
    },
    [onEscape]
  )

  useEffect(() => {
    document.addEventListener('keydown', handleEscape)
    return () => {
      document.removeEventListener('keydown', handleEscape)
    }
  }, [handleEscape])
}
