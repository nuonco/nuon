import { useEffect, useCallback } from 'react'
import { useSurfaces } from '@/hooks/use-surfaces'
import { HelpModal } from '@/components/spotlight/HelpModal'
import React from 'react'

export function useHelp() {
  const { addModal } = useSurfaces()

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'h') {
        const tag = (e.target as HTMLElement)?.tagName
        if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return

        e.preventDefault()
        addModal(React.createElement(HelpModal))
      }
    },
    [addModal]
  )

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])
}
