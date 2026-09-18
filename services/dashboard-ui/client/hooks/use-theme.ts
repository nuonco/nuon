import { useContext, useEffect, useState } from 'react'
import {
  ThemeContext,
  systemColorScheme,
  type TColorScheme,
} from '@/providers/theme-provider'

export function useTheme() {
  const ctx = useContext(ThemeContext)
  if (!ctx) {
    throw new Error('useTheme must be used within a ThemeProvider')
  }
  return ctx
}

export function useColorScheme(): TColorScheme {
  const ctx = useContext(ThemeContext)
  const [systemTheme, setSystemTheme] = useState<TColorScheme>(systemColorScheme)

  useEffect(() => {
    if (ctx) return

    const matcher = window.matchMedia('(prefers-color-scheme: dark)')
    const update = () => setSystemTheme(matcher.matches ? 'dark' : 'light')
    matcher.addEventListener('change', update)
    update()

    return () => matcher.removeEventListener('change', update)
  }, [ctx])

  return ctx?.colorScheme ?? systemTheme
}
