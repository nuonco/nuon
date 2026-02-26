'use client'

import { useEffect, useState } from 'react'

export function useSystemTheme(): 'dark' | 'light' {
  const [theme, setTheme] = useState<'dark' | 'light'>(() => {
    if (typeof window === 'undefined') return 'light'
    return window.matchMedia('(prefers-color-scheme: dark)').matches
      ? 'dark'
      : 'light'
  })

  useEffect(() => {
    if (typeof window === 'undefined') return

    const matcher = window.matchMedia('(prefers-color-scheme: dark)')
    const update = () => setTheme(matcher.matches ? 'dark' : 'light')
    matcher.addEventListener('change', update)
    // Initialize on mount
    update()

    return () => matcher.removeEventListener('change', update)
  }, [])

  return theme
}
