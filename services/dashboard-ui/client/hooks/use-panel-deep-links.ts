import { useEffect, useRef, type ReactElement } from 'react'
import { useSearchParams } from 'react-router'
import type { IPanel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import { panelKeyStack } from '@/providers/surfaces-provider'

export type TPanelResolver = (key: string) => ReactElement<IPanel> | null | undefined

// Return null for keys you don't own or can't build yet; the resolver is
// retried on every render until it returns a panel.
export const usePanelDeepLinks = (resolve: TPanelResolver) => {
  const { panels, addPanel } = useSurfaces()
  const [searchParams] = useSearchParams()
  const requested = useRef(new Set<string>())
  const stack = panelKeyStack(searchParams)

  useEffect(() => {
    for (const key of requested.current) {
      if (!stack.includes(key)) requested.current.delete(key)
    }
    for (const key of stack) {
      if (requested.current.has(key)) continue
      if (panels.some((panel) => panel.key === key)) continue
      const panel = resolve(key)
      if (!panel) continue
      requested.current.add(key)
      addPanel(panel, key)
    }
  })
}
