'use client'

import { useRouter } from 'next/navigation'
import { useCallback } from 'react'
import { useSurfaces } from './use-surfaces'

export function useRemovePanelByKey() {
  const { panels, removePanel } = useSurfaces()
  const router = useRouter()

  return useCallback(
    (key: string) => {
      const panel = panels?.find((p) => p?.key === key)
      if (panel) {
        const params = new URLSearchParams(window.location.search)
        params.delete('panel')
        router.replace(`?${params.toString()}`, { scroll: false })
        removePanel(panel.id)
      }
    },
    [panels, removePanel, router]
  )
}
