import { useEffect, useCallback } from 'react'
import { useSurfaces } from '@/hooks/use-surfaces'
import { SpotlightModalContainer } from '@/components/spotlight/Spotlight'
import React from 'react'

export function useSpotlight() {
  const { addModal } = useSurfaces()

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        const tag = (e.target as HTMLElement)?.tagName
        if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return

        e.preventDefault()
        addModal(React.createElement(SpotlightModalContainer))
      }
    },
    [addModal]
  )

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])
}
