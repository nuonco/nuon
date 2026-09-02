import { createContext, useCallback, useEffect, useMemo, useState } from 'react'

export type TThemePreference = 'light' | 'dark' | 'high-contrast' | 'system'
export type TTheme = 'light' | 'dark' | 'high-contrast'

export const THEME_STORAGE_KEY = 'nuon-lite-theme'

interface IThemeContext {
  preference: TThemePreference
  theme: TTheme
  setPreference: (preference: TThemePreference) => void
}

export const ThemeContext = createContext<IThemeContext | undefined>(undefined)

const PREFERENCES: TThemePreference[] = [
  'light',
  'dark',
  'high-contrast',
  'system',
]

const isPreference = (value: unknown): value is TThemePreference =>
  PREFERENCES.includes(value as TThemePreference)

const darkQuery = () => window.matchMedia('(prefers-color-scheme: dark)')

const readPreference = (): TThemePreference => {
  try {
    const stored = localStorage.getItem(THEME_STORAGE_KEY)
    return isPreference(stored) ? stored : 'system'
  } catch {
    return 'system'
  }
}

export const ThemeProvider = ({ children }: { children: React.ReactNode }) => {
  const [preference, setStoredPreference] =
    useState<TThemePreference>(readPreference)
  const [systemTheme, setSystemTheme] = useState<TTheme>(() =>
    darkQuery().matches ? 'dark' : 'light'
  )

  useEffect(() => {
    const query = darkQuery()
    const onChange = (event: MediaQueryListEvent) =>
      setSystemTheme(event.matches ? 'dark' : 'light')
    query.addEventListener('change', onChange)
    return () => query.removeEventListener('change', onChange)
  }, [])

  useEffect(() => {
    const root = document.documentElement
    if (preference === 'system') {
      root.removeAttribute('data-theme')
    } else {
      root.setAttribute('data-theme', preference)
    }
  }, [preference])

  const setPreference = useCallback((next: TThemePreference) => {
    try {
      localStorage.setItem(THEME_STORAGE_KEY, next)
    } catch {}
    setStoredPreference(next)
  }, [])

  const theme = preference === 'system' ? systemTheme : preference

  const value = useMemo(
    () => ({ preference, theme, setPreference }),
    [preference, theme, setPreference]
  )

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}
