import { useSyncExternalStore } from 'react'

const mediaQuery = (query: string) =>
  typeof window === 'undefined' || !window.matchMedia
    ? undefined
    : window.matchMedia(query)

export const useMediaQuery = (query: string) =>
  useSyncExternalStore(
    (onChange) => {
      const media = mediaQuery(query)
      media?.addEventListener('change', onChange)
      return () => media?.removeEventListener('change', onChange)
    },
    () => mediaQuery(query)?.matches ?? false,
    () => false
  )
