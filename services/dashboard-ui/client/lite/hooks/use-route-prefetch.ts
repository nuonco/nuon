import { useEffect, useRef } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { prefetchRoute } from '../utils/route-prefetch'

export const useRoutePrefetch = () => {
  const queryClient = useQueryClient()
  const lastHref = useRef<string | null>(null)

  useEffect(() => {
    const onIntent = (event: Event) => {
      const target = event.target
      if (!(target instanceof Element)) return

      const anchor = target.closest('a[href]')
      if (!(anchor instanceof HTMLAnchorElement)) return
      if (anchor.hasAttribute('download') || anchor.target === '_blank') return

      const href = anchor.getAttribute('href') ?? ''
      if (href === lastHref.current) return

      lastHref.current = href
      prefetchRoute(queryClient, href)
    }

    document.addEventListener('pointerover', onIntent)
    document.addEventListener('focusin', onIntent)

    return () => {
      document.removeEventListener('pointerover', onIntent)
      document.removeEventListener('focusin', onIntent)
    }
  }, [queryClient])
}
