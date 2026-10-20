import { useEffect, useRef, type ReactElement } from 'react'
import { useSearchParams } from 'react-router'
import type { IPanel } from '@/components/surfaces/Panel'
import { hasPanelKey } from '@/providers/surfaces-provider'
import { useSurfaces } from './use-surfaces'

type TDeepLinkablePanel = {
  key: string
  build: () => ReactElement<IPanel>
}

export const useDeepLinkedPanels = (items: TDeepLinkablePanel[]) => {
  const { addPanel, panels } = useSurfaces()
  const [searchParams] = useSearchParams()
  const openedKeys = useRef(new Set<string>())

  useEffect(() => {
    for (const item of items) {
      if (openedKeys.current.has(item.key)) continue
      if (panels.some((p) => p.key === item.key)) continue
      if (!hasPanelKey(searchParams, item.key)) continue
      openedKeys.current.add(item.key)
      addPanel(item.build(), item.key)
    }
  }, [items, panels, searchParams, addPanel])
}
