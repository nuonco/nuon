import { useNavigate } from 'react-router'
import { useCallback } from 'react'
import { useSurfaces } from './use-surfaces'

export function useRemovePanelByKey() {
  const { panels, removePanel } = useSurfaces()
  const navigate = useNavigate()

  return useCallback(
    (key: string) => {
      const panel = panels?.find((p) => p?.key === key)
      if (panel) {
        const params = new URLSearchParams(window.location.search)
        params.delete('panel')
        navigate(`?${params.toString()}`, { replace: true })
        removePanel(panel.id)
      }
    },
    [panels, removePanel, navigate]
  )
}
