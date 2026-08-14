import { useEffect, useState } from 'react'

const DESKTOP_QUERY = '(min-width: 768px)'

export function useIsDesktop(): boolean {
  const [isDesktop, setIsDesktop] = useState<boolean>(() => {
    if (typeof window === 'undefined') return true
    return window.matchMedia(DESKTOP_QUERY).matches
  })

  useEffect(() => {
    if (typeof window === 'undefined') return

    const matcher = window.matchMedia(DESKTOP_QUERY)
    const update = () => setIsDesktop(matcher.matches)
    matcher.addEventListener('change', update)
    update()

    return () => matcher.removeEventListener('change', update)
  }, [])

  return isDesktop
}
